package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/message"
)

type PlayRoomService struct {
	gameRepository   *repopg.GameRepository
	walletRepository *repopg.WalletRepository
}

func NewPlayRoomService(gameRepository *repopg.GameRepository, walletRepository *repopg.WalletRepository) *PlayRoomService {
	return &PlayRoomService{
		gameRepository:   gameRepository,
		walletRepository: walletRepository,
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

	room, err := s.gameRepository.FindRoomByCode(ctx, roomCode)
	if err != nil {
		return game.RoomStateResponse{}, err
	}

	period, err := s.gameRepository.GetCurrentPeriodByRoom(ctx, roomCode)
	if err != nil {
		return game.RoomStateResponse{}, err
	}

	results, err := s.gameRepository.ListRoomRecentRounds(ctx, roomCode, 20)
	if err != nil {
		return game.RoomStateResponse{}, err
	}

	recentResults := make([]game.HistoryListItem, 0, len(results))
	for _, row := range results {
		recentResults = append(recentResults, game.HistoryListItem{
			PeriodNo:  row.PeriodNo,
			Result:    row.Result,
			BigSmall:  row.BigSmall,
			Color:     row.Color,
			DrawAt:    row.DrawAt,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}

	return game.RoomStateResponse{
		Message:    message.RoomStateSuccess,
		ServerTime: time.Now(),
		Room: game.RoomItem{
			Code:             room.Code,
			GameType:         toGameTypeSlug(room.GameType),
			DurationSeconds:  room.DurationSeconds,
			BetCutoffSeconds: room.BetCutoffSeconds,
			Status:           toRoomStatusLabel(room.Status),
			SortOrder:        room.SortOrder,
		},
		CurrentPeriod: game.RoomPeriod{
			ID:        period.ID,
			PeriodNo:  period.PeriodNo,
			Status:    toPeriodStatusLabel(period.Status),
			OpenAt:    period.OpenAt,
			BetLockAt: period.BetLockAt,
			DrawAt:    period.DrawAt,
		},
		RecentResults: recentResults,
	}, nil
}

func (s *PlayRoomService) ListRoomHistory(ctx context.Context, roomCode string, page, pageSize int) (game.HistoryListResponse, error) {
	records, total, err := s.gameRepository.ListRoomRounds(ctx, roomCode, page, pageSize)
	if err != nil {
		return game.HistoryListResponse{}, err
	}

	items := make([]game.HistoryListItem, 0, len(records))
	for _, record := range records {
		items = append(items, game.HistoryListItem{
			PeriodNo:  record.PeriodNo,
			Result:    record.Result,
			BigSmall:  record.BigSmall,
			Color:     record.Color,
			DrawAt:    record.DrawAt,
			Status:    record.Status,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
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
		items = append(items, game.BetTicketHistoryItem{
			ID:         record.ID,
			PeriodNo:   record.PeriodNo,
			Result:     summary.Result,
			BigSmall:   summary.BigSmall,
			Color:      summary.Color,
			Stake:      record.TotalStake,
			Status:     toBetStatusLabel(record.Status),
			ItemsCount: summary.ItemsCount,
			CreatedAt:  record.CreatedAt,
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
) (game.RoomBetResponse, error) {
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

	available, err := subtractDecimal(wallet.Balance, wallet.LockedBalance)
	if err != nil {
		return game.RoomBetResponse{}, err
	}

	if compareDecimal(available, totalStake) < 0 {
		return game.RoomBetResponse{}, fmt.Errorf(message.InsufficientBalanceBet)
	}

	if _, err := s.gameRepository.CreateBetTicket(ctx, repopg.CreateBetTicketParams{
		UserID:       userID,
		RoomCode:     roomCode,
		PeriodID:     periodID,
		RequestID:    req.RequestID,
		ConnectionID: "",
		TotalStake:   totalStake,
		Items:        toBetTicketItems(req.Items),
		PlacedIP:     placedIP,
		PlacedDevice: placedDevice,
	}); err != nil {
		return game.RoomBetResponse{}, err
	}

	return game.RoomBetResponse{
		RequestID:  req.RequestID,
		RoomCode:   roomCode,
		Status:     "accepted",
		AcceptedAt: time.Now(),
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

func ParseUserID(raw string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf(message.UserIDRequired)
	}
	return parsed, nil
}
