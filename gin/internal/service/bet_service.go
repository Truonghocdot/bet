package service

import (
	"context"
	"fmt"
	"time"

	"gin/internal/domain/game"
	"gin/internal/event/outbox"
	"gin/internal/support/id"
	"gin/internal/support/message"
)

type BetService struct {
	publisher      outbox.Publisher
	sessionService *GameSessionService
}

func NewBetService(publisher outbox.Publisher, sessionService *GameSessionService) *BetService {
	return &BetService{
		publisher:      publisher,
		sessionService: sessionService,
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

	event := outbox.Event{
		ID:         id.New(),
		Name:       "bet.placed",
		OccurredAt: time.Now(),
		Payload: map[string]any{
			"connection_id": connectionID,
			"user_id":       request.UserID,
			"game_type":     request.GameType,
			"period_id":     request.PeriodID,
			"request_id":    request.RequestID,
			"items_count":   len(request.Items),
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
