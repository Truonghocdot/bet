package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/message"

	goredis "github.com/redis/go-redis/v9"
)

type PlayRoomService struct {
	gameRepository   *repopg.GameRepository
	walletRepository *repopg.WalletRepository
	walletService    *WalletService
	redis            *goredis.Client
	broker           *realtime.Broker
	roomStateTTL     time.Duration
}

type cachedRoomState struct {
	Version     int64                  `json:"version"`
	GeneratedAt time.Time              `json:"generated_at"`
	Source      string                 `json:"source"`
	Payload     game.RoomStateResponse `json:"payload"`
}

func NewPlayRoomService(
	gameRepository *repopg.GameRepository,
	walletRepository *repopg.WalletRepository,
	walletService *WalletService,
	redis *goredis.Client,
	broker *realtime.Broker,
) *PlayRoomService {
	return &PlayRoomService{
		gameRepository:   gameRepository,
		walletRepository: walletRepository,
		walletService:    walletService,
		redis:            redis,
		broker:           broker,
		roomStateTTL:     2 * time.Hour,
	}
}

func (s *PlayRoomService) ListRooms(ctx context.Context) (game.RoomListResponse, error) {
	rooms, err := s.gameRepository.ListRooms(ctx)
	if err != nil {
		return game.RoomListResponse{}, err
	}

	items := make([]game.RoomItem, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, game.RoomItem{
			Code:             room.Code,
			GameType:         toGameTypeSlug(room.GameType),
			DurationSeconds:  room.DurationSeconds,
			BetCutoffSeconds: room.BetCutoffSeconds,
			Status:           toRoomStatusLabel(room.Status),
			SortOrder:        room.SortOrder,
		})
	}

	return game.RoomListResponse{
		Message: message.RoomListSuccess,
		Items:   items,
	}, nil
}

func (s *PlayRoomService) GetRoomState(ctx context.Context, roomCode string) (game.RoomStateResponse, error) {
	if strings.TrimSpace(roomCode) == "" {
		return game.RoomStateResponse{}, fmt.Errorf(message.GameRoomCodeRequired)
	}

	if cached, hit := s.readRoomStateCache(ctx, roomCode); hit {
		if !isRoomStateCacheFresh(cached) {
		} else {
			cached.ServerTime = clock.Now()
			log.Printf("[realtime][room.cache.hit] room_code=%s", roomCode)
			return cached, nil
		}
	}

	response, err := s.RefreshRoomState(ctx, roomCode, "cache_rebuild")
	if err != nil {
		return game.RoomStateResponse{}, err
	}
	response.ServerTime = clock.Now()
	return response, nil
}

func isRoomStateCacheFresh(response game.RoomStateResponse) bool {
	now := clock.Now()
	period := response.CurrentPeriod
	if period.ID == 0 || strings.TrimSpace(period.PeriodNo) == "" {
		return false
	}

	status := strings.ToUpper(strings.TrimSpace(period.Status))
	if status != "SCHEDULED" && status != "OPEN" && status != "LOCKED" {
		return false
	}

	// Once draw_at has passed, the cache is already outdated for bet placement.
	if !period.DrawAt.After(now) {
		return false
	}

	// Guard against obviously drifted snapshots that point too far in the past.
	if period.OpenAt.After(now.Add(15 * time.Minute)) {
		return false
	}

	return true
}

func (s *PlayRoomService) RefreshRoomState(ctx context.Context, roomCode string, source string) (game.RoomStateResponse, error) {
	response, err := s.buildRoomState(ctx, roomCode)
	if err != nil {
		return game.RoomStateResponse{}, err
	}

	s.writeRoomStateCache(ctx, roomCode, response, source)
	if err := s.broker.Publish(ctx, realtime.PlayRoomTopic(roomCode), "room.state", response); err != nil {
		log.Printf("[realtime][room.publish.error] room_code=%s source=%s err=%v", roomCode, source, err)
	}
	if err := s.broker.Publish(ctx, realtime.AdminRoomsTopic(), "admin.rooms.changed", map[string]any{
		"room_code": roomCode,
		"source":    source,
		"at":        clock.Now(),
	}); err != nil {
		log.Printf("[realtime][admin.rooms.publish.error] room_code=%s source=%s err=%v", roomCode, source, err)
	}

	return response, nil
}

func (s *PlayRoomService) ListRoomHistory(ctx context.Context, roomCode string, page, pageSize int) (game.HistoryListResponse, error) {
	records, total, err := s.gameRepository.ListRoomRounds(ctx, roomCode, page, pageSize)
	if err != nil {
		return game.HistoryListResponse{}, err
	}

	items := make([]game.HistoryListItem, 0, len(records))
	for _, record := range records {
		items = append(items, game.HistoryListItem{
			PeriodNo:    record.PeriodNo,
			PeriodIndex: record.PeriodIndex,
			Result:      record.Result,
			BigSmall:    record.BigSmall,
			Color:       record.Color,
			DrawAt:      record.DrawAt,
			Status:      record.Status,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		})
	}

	return game.HistoryListResponse{
		Message:    message.RoomHistorySuccess,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: calcTotalPages(total, pageSize),
		Items:      items,
	}, nil
}

func (s *PlayRoomService) ListMyRoomBets(ctx context.Context, userID int64, roomCode string, page, pageSize int) (game.BetTicketHistoryResponse, error) {
	records, total, err := s.gameRepository.ListRoomBetTickets(ctx, userID, roomCode, page, pageSize)
	if err != nil {
		return game.BetTicketHistoryResponse{}, err
	}

	items := make([]game.BetTicketHistoryItem, 0, len(records))
	for _, record := range records {
		summary := summarizeBetTicket(record.ItemsJSON)
		profitLoss := record.ProfitLoss
		if strings.TrimSpace(profitLoss) == "" {
			profitLoss = "0"
		}
		items = append(items, game.BetTicketHistoryItem{
			ID:             record.ID,
			PeriodID:       record.PeriodID,
			PeriodNo:       record.PeriodNo,
			PeriodIndex:    record.PeriodIndex,
			Result:         summary.Result,
			BigSmall:       summary.BigSmall,
			Color:          summary.Color,
			Stake:          record.TotalStake,
			OriginalAmount: record.OriginalAmount,
			TaxAmount:      record.TaxAmount,
			NetAmount:      record.NetAmount,
			ActualPayout:   record.ActualPayout,
			ProfitLoss:     profitLoss,
			SettledAt:      nullTimePtr(record.SettledAt),
			Status:         toBetStatusLabel(record.Status),
			ItemsCount:     summary.ItemsCount,
			CreatedAt:      record.CreatedAt,
		})
	}

	return game.BetTicketHistoryResponse{
		Message:    "Lấy lịch sử cược thành công",
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: calcTotalPages(total, pageSize),
		Items:      items,
	}, nil
}

func (s *PlayRoomService) PlaceRoomBet(
	ctx context.Context,
	userID int64,
	roomCode string,
	req game.RoomBetRequest,
	placedIP string,
	placedDevice string,
	connectionID string,
) (game.RoomBetResponse, error) {
	log.Printf(
		"[play.bet.service.start] room=%s user_id=%d request_id=%s period_id=%s connection_id=%s items=%d",
		strings.TrimSpace(roomCode),
		userID,
		strings.TrimSpace(req.RequestID),
		strings.TrimSpace(req.PeriodID),
		strings.TrimSpace(connectionID),
		len(req.Items),
	)
	if strings.TrimSpace(roomCode) == "" {
		return game.RoomBetResponse{}, fmt.Errorf(message.GameRoomCodeRequired)
	}
	if strings.TrimSpace(req.RequestID) == "" {
		return game.RoomBetResponse{}, fmt.Errorf(message.RequestIDRequired)
	}
	if strings.TrimSpace(req.PeriodID) == "" {
		return game.RoomBetResponse{}, fmt.Errorf(message.PeriodIDRequired)
	}
	if len(req.Items) == 0 {
		return game.RoomBetResponse{}, fmt.Errorf(message.BetItemsRequired)
	}

	if _, err := s.gameRepository.FindRoomByCode(ctx, roomCode); err != nil {
		return game.RoomBetResponse{}, err
	}

	periodID, err := repopg.ParsePeriodID(req.PeriodID)
	if err != nil {
		return game.RoomBetResponse{}, err
	}

	totalStake, err := sumBetItems(req.Items)
	if err != nil {
		return game.RoomBetResponse{}, err
	}

	wallet, err := s.walletRepository.FindByUserAndUnit(ctx, userID, 1)
	if err != nil {
		return game.RoomBetResponse{}, err
	}

	available := strings.TrimSpace(wallet.Balance)
	if available == "" {
		available = "0"
	}

	if compareDecimal(available, totalStake) < 0 {
		log.Printf(
			"[play.bet.service.reject] room=%s user_id=%d request_id=%s reason=insufficient_balance available=%s stake=%s",
			roomCode,
			userID,
			req.RequestID,
			available,
			totalStake,
		)
		return game.RoomBetResponse{}, fmt.Errorf(message.InsufficientBalanceBet)
	}

	ticket, err := s.gameRepository.CreateBetTicket(ctx, repopg.CreateBetTicketParams{
		UserID:       userID,
		RoomCode:     roomCode,
		PeriodID:     periodID,
		RequestID:    req.RequestID,
		ConnectionID: strings.TrimSpace(connectionID),
		TotalStake:   totalStake,
		Items:        toBetTicketItems(req.Items),
		PlacedIP:     placedIP,
		PlacedDevice: placedDevice,
	})
	if err != nil {
		log.Printf(
			"[play.bet.service.error] room=%s user_id=%d request_id=%s stage=create_ticket err=%v",
			roomCode,
			userID,
			req.RequestID,
			err,
		)
		return game.RoomBetResponse{}, err
	}

	if s.walletService != nil {
		if err := s.walletService.PublishSummary(ctx, ticket.UserID); err != nil {
			log.Printf("[realtime][wallet.publish.error] user_id=%d source=place_room_bet err=%v", ticket.UserID, err)
		}
	}
	if s.broker != nil {
		if err := s.broker.Publish(ctx, realtime.PlayRoomBetsTopic(roomCode, ticket.UserID), "bets.updated", map[string]any{
			"type":      "bet_placed",
			"room_code": roomCode,
			"ticket_id": ticket.ID,
			"user_id":   ticket.UserID,
			"at":        clock.Now(),
		}); err != nil {
			log.Printf("[realtime][bets.publish.error] room_code=%s user_id=%d ticket_id=%d err=%v", roomCode, ticket.UserID, ticket.ID, err)
		}
	}

	log.Printf(
		"[play.bet.service.ok] room=%s user_id=%d request_id=%s ticket_id=%d total_stake=%s",
		roomCode,
		userID,
		req.RequestID,
		ticket.ID,
		totalStake,
	)
	return game.RoomBetResponse{
		RequestID:  req.RequestID,
		RoomCode:   roomCode,
		Status:     "accepted",
		AcceptedAt: clock.Now(),
		Message:    message.BetAccepted,
	}, nil
}

func toGameTypeSlug(gameType int) string {
	switch gameType {
	case 1:
		return "wingo"
	case 2:
		return "k3"
	case 3:
		return "lottery"
	default:
		return "unknown"
	}
}

func toRoomStatusLabel(status int) string {
	switch status {
	case 1:
		return "ACTIVE"
	case 2:
		return "INACTIVE"
	default:
		return "UNKNOWN"
	}
}

func toPeriodStatusLabel(status int) string {
	switch status {
	case 1:
		return "SCHEDULED"
	case 2:
		return "OPEN"
	case 3:
		return "LOCKED"
	case 4:
		return "DRAWN"
	case 5:
		return "SETTLED"
	case 6:
		return "CANCELED"
	default:
		return "UNKNOWN"
	}
}

func (s *PlayRoomService) buildRoomState(ctx context.Context, roomCode string) (game.RoomStateResponse, error) {
	room, err := s.gameRepository.FindRoomByCode(ctx, roomCode)
	if err != nil {
		return game.RoomStateResponse{}, err
	}

	period, err := s.gameRepository.GetCurrentPeriodByRoom(ctx, roomCode)
	if err != nil {
		if errors.Is(err, repopg.ErrPeriodNotFound) {
			if _, ensureErr := s.gameRepository.EnsureRoomPeriods(ctx, room, clock.Now()); ensureErr != nil {
				log.Printf("[realtime][room.cache.rebuild.error] room_code=%s stage=ensure_period err=%v", roomCode, ensureErr)
			}

			period, err = s.gameRepository.GetCurrentPeriodByRoom(ctx, roomCode)
			if errors.Is(err, repopg.ErrPeriodNotFound) {
				period, err = s.gameRepository.GetNearestUpcomingPeriodByRoom(ctx, roomCode)
			}
		}
		if err != nil {
			log.Printf("[realtime][room.cache.rebuild.error] room_code=%s stage=current_period err=%v", roomCode, err)
			return game.RoomStateResponse{}, err
		}
	}

	results, err := s.gameRepository.ListRoomRecentRounds(ctx, roomCode, 20)
	if err != nil {
		return game.RoomStateResponse{}, err
	}

	recentResults := make([]game.HistoryListItem, 0, len(results))
	for _, row := range results {
		recentResults = append(recentResults, game.HistoryListItem{
			PeriodNo:    row.PeriodNo,
			PeriodIndex: row.PeriodIndex,
			Result:      row.Result,
			BigSmall:    row.BigSmall,
			Color:       row.Color,
			DrawAt:      row.DrawAt,
			Status:      row.Status,
			CreatedAt:   row.CreatedAt,
			UpdatedAt:   row.UpdatedAt,
		})
	}

	return game.RoomStateResponse{
		Message:    message.RoomStateSuccess,
		ServerTime: clock.Now(),
		Room: game.RoomItem{
			Code:             room.Code,
			GameType:         toGameTypeSlug(room.GameType),
			DurationSeconds:  room.DurationSeconds,
			BetCutoffSeconds: room.BetCutoffSeconds,
			Status:           toRoomStatusLabel(room.Status),
			SortOrder:        room.SortOrder,
		},
		CurrentPeriod: game.RoomPeriod{
			ID:          period.ID,
			PeriodNo:    period.PeriodNo,
			PeriodIndex: period.PeriodIndex,
			Status:      toPeriodStatusLabel(period.Status),
			OpenAt:      period.OpenAt,
			BetLockAt:   period.BetLockAt,
			DrawAt:      period.DrawAt,
		},
		RecentResults: recentResults,
	}, nil
}

func (s *PlayRoomService) roomStateCacheKey(roomCode string) string {
	return fmt.Sprintf("cache:play:room_state:%s", strings.TrimSpace(roomCode))
}

func (s *PlayRoomService) readRoomStateCache(ctx context.Context, roomCode string) (game.RoomStateResponse, bool) {
	if s.redis == nil || strings.TrimSpace(roomCode) == "" {
		return game.RoomStateResponse{}, false
	}

	raw, err := s.redis.Get(ctx, s.roomStateCacheKey(roomCode)).Bytes()
	if err != nil || len(raw) == 0 {
		return game.RoomStateResponse{}, false
	}

	var cached cachedRoomState
	if err := json.Unmarshal(raw, &cached); err != nil {
		return game.RoomStateResponse{}, false
	}
	if cached.Payload.Room.Code == "" || cached.Payload.CurrentPeriod.ID == 0 {
		return game.RoomStateResponse{}, false
	}

	return cached.Payload, true
}

func (s *PlayRoomService) writeRoomStateCache(ctx context.Context, roomCode string, response game.RoomStateResponse, source string) {
	if s.redis == nil || strings.TrimSpace(roomCode) == "" {
		return
	}

	payload, err := json.Marshal(cachedRoomState{
		Version:     time.Now().UnixNano(),
		GeneratedAt: clock.Now(),
		Source:      source,
		Payload:     response,
	})
	if err != nil {
		log.Printf("[realtime][room.cache.write.error] room_code=%s stage=marshal err=%v", roomCode, err)
		return
	}

	if err := s.redis.Set(ctx, s.roomStateCacheKey(roomCode), payload, s.roomStateTTL).Err(); err != nil {
		log.Printf("[realtime][room.cache.write.error] room_code=%s stage=set err=%v", roomCode, err)
		return
	}
}

func ParseUserID(raw string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf(message.UserIDRequired)
	}
	return parsed, nil
}
