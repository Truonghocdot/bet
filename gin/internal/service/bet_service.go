package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	"gin/internal/event/outbox"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/id"
	"gin/internal/support/message"
)

type BetService struct {
	publisher        outbox.Publisher
	sessionService   *GameSessionService
	gameRepository   *repopg.GameRepository
	walletRepository *repopg.WalletRepository
}

func NewBetService(
	publisher outbox.Publisher,
	sessionService *GameSessionService,
	gameRepository *repopg.GameRepository,
	walletRepository *repopg.WalletRepository,
) *BetService {
	return &BetService{
		publisher:        publisher,
		sessionService:   sessionService,
		gameRepository:   gameRepository,
		walletRepository: walletRepository,
	}
}

func (s *BetService) PlaceBet(ctx context.Context, connectionID string, request game.PlaceBetRequest) (game.PlaceBetResponse, error) {
	if request.RequestID == "" {
		return game.PlaceBetResponse{}, fmt.Errorf(message.RequestIDRequired)
	}

	if request.UserID == "" {
		return game.PlaceBetResponse{}, fmt.Errorf(message.UserIDRequired)
	}

	if request.GameType == "" {
		return game.PlaceBetResponse{}, fmt.Errorf(message.GameTypeRequired)
	}

	if request.PeriodID == "" {
		return game.PlaceBetResponse{}, fmt.Errorf(message.PeriodIDRequired)
	}

	if len(request.Items) == 0 {
		return game.PlaceBetResponse{}, fmt.Errorf(message.BetItemsRequired)
	}

	if err := s.sessionService.ValidateConnection(connectionID, request.UserID, request.GameType); err != nil {
		return game.PlaceBetResponse{}, err
	}

	roomCode, ok := game.DefaultRoomCode(request.GameType)
	if !ok {
		return game.PlaceBetResponse{}, fmt.Errorf(message.GameRoomNotFound)
	}

	userID, err := strconv.ParseInt(strings.TrimSpace(request.UserID), 10, 64)
	if err != nil {
		return game.PlaceBetResponse{}, fmt.Errorf(message.Unauthorized)
	}

	periodID, err := repopg.ParsePeriodID(request.PeriodID)
	if err != nil {
		return game.PlaceBetResponse{}, err
	}

	totalStake, err := sumBetItems(request.Items)
	if err != nil {
		return game.PlaceBetResponse{}, err
	}

	if s.walletRepository != nil {
		wallet, err := s.walletRepository.FindByUserAndUnit(ctx, userID, 1)
		if err != nil {
			return game.PlaceBetResponse{}, err
		}

		available, err := subtractDecimal(wallet.Balance, wallet.LockedBalance)
		if err != nil {
			return game.PlaceBetResponse{}, err
		}

		if compareDecimal(available, totalStake) < 0 {
			return game.PlaceBetResponse{}, fmt.Errorf(message.InsufficientBalanceBet)
		}
	}

	ticket, err := s.gameRepository.CreateBetTicket(ctx, repopg.CreateBetTicketParams{
		UserID:       userID,
		RoomCode:     roomCode,
		PeriodID:     periodID,
		RequestID:    request.RequestID,
		ConnectionID: connectionID,
		TotalStake:   totalStake,
		Items:        toBetTicketItems(request.Items),
	})
	if err != nil {
		return game.PlaceBetResponse{}, err
	}

	event := outbox.Event{
		ID:         id.New(),
		Name:       "bet.placed",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"connection_id": connectionID,
			"user_id":       request.UserID,
			"game_type":     request.GameType,
			"period_id":     periodID,
			"room_code":     roomCode,
			"request_id":    request.RequestID,
			"items_count":   len(request.Items),
			"ticket_id":     ticket.ID,
		},
	}

	if err := s.publisher.Publish(ctx, event); err != nil {
		return game.PlaceBetResponse{}, err
	}

	return game.PlaceBetResponse{
		RequestID:    request.RequestID,
		ConnectionID: connectionID,
		GameType:     request.GameType,
		Status:       "accepted",
		AcceptedAt:   time.Now(),
		Message:      message.BetAccepted,
	}, nil
}

func (s *BetService) ListGameHistory(ctx context.Context, gameType game.GameType, page, pageSize int) (game.HistoryListResponse, error) {
	records, total, err := s.gameRepository.ListGameRounds(ctx, string(gameType), page, pageSize)
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
		Message:    "Lấy lịch sử game thành công",
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: calcTotalPages(total, pageSize),
		Items:      items,
	}, nil
}

func (s *BetService) ListMyBets(ctx context.Context, userID int64, gameType game.GameType, page, pageSize int) (game.BetTicketHistoryResponse, error) {
	records, total, err := s.gameRepository.ListBetTickets(ctx, userID, string(gameType), page, pageSize)
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

func sumBetItems(items []game.BetItem) (string, error) {
	total := new(big.Rat)
	for _, item := range items {
		amount, err := parseDecimal(item.Stake)
		if err != nil {
			return "", err
		}
		total.Add(total, amount)
	}

	return total.FloatString(8), nil
}

type betSummary struct {
	Result     string
	BigSmall   string
	Color      string
	ItemsCount int
}

func summarizeBetTicket(raw []byte) betSummary {
	if len(raw) == 0 {
		return betSummary{Result: "Chờ xử lý", BigSmall: "—", Color: "—", ItemsCount: 0}
	}

	var items []repopg.BetTicketItemRecord
	if err := json.Unmarshal(raw, &items); err != nil {
		return betSummary{Result: "Chờ xử lý", BigSmall: "—", Color: "—", ItemsCount: 0}
	}

	result := "Chờ xử lý"
	bigSmall := "—"
	color := "—"
	for _, item := range items {
		label := strings.TrimSpace(item.OptionKey)
		if result == "Chờ xử lý" && label != "" {
			result = normalizeResultLabel(item.OptionType, label)
		}
		switch strings.ToUpper(strings.TrimSpace(item.OptionType)) {
		case "COLOR":
			color = normalizeColorLabel(label)
		case "BIG_SMALL":
			bigSmall = normalizeBigSmallLabel(label)
		case "NUMBER", "SUM", "ODD_EVEN", "COMBINATION":
			if result == "Chờ xử lý" {
				result = label
			}
		}
	}

	return betSummary{Result: result, BigSmall: bigSmall, Color: color, ItemsCount: len(items)}
}

func normalizeResultLabel(optionType, label string) string {
	switch strings.ToUpper(strings.TrimSpace(optionType)) {
	case "COLOR":
		return normalizeColorLabel(label)
	case "BIG_SMALL":
		return normalizeBigSmallLabel(label)
	default:
		return label
	}
}

func normalizeColorLabel(label string) string {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "xanh"):
		return "Xanh"
	case strings.Contains(lower, "đỏ"):
		return "Đỏ"
	case strings.Contains(lower, "tím"):
		return "Tím"
	default:
		return label
	}
}

func normalizeBigSmallLabel(label string) string {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "lớn"):
		return "Lớn"
	case strings.Contains(lower, "nhỏ"):
		return "Nhỏ"
	default:
		return label
	}
}

func toBetTicketItems(items []game.BetItem) []repopg.BetTicketItemRecord {
	result := make([]repopg.BetTicketItemRecord, 0, len(items))
	for _, item := range items {
		result = append(result, repopg.BetTicketItemRecord{
			OptionType: item.OptionType,
			OptionKey:  item.OptionKey,
			Stake:      item.Stake,
		})
	}
	return result
}

func calcTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return pages
}

func toBetStatusLabel(status int) string {
	switch status {
	case 1:
		return "PENDING"
	case 2:
		return "WON"
	case 3:
		return "LOST"
	case 4:
		return "VOID"
	case 5:
		return "HALF_WON"
	case 6:
		return "HALF_LOST"
	case 7:
		return "CANCELED"
	case 8:
		return "CASHED_OUT"
	default:
		return "UNKNOWN"
	}
}

func subtractDecimal(left, right string) (string, error) {
	lv, err := parseDecimal(left)
	if err != nil {
		return "", err
	}
	rv, err := parseDecimal(right)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Sub(lv, rv).FloatString(8), nil
}

func compareDecimal(left, right string) int {
	lv, err := parseDecimal(left)
	if err != nil {
		return -1
	}
	rv, err := parseDecimal(right)
	if err != nil {
		return -1
	}
	return lv.Cmp(rv)
}

func parseDecimal(value string) (*big.Rat, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(strings.TrimSpace(value)); !ok {
		return nil, fmt.Errorf("invalid decimal value: %s", value)
	}
	return rat, nil
}
