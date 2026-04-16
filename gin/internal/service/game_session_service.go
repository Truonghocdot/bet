package service

import (
	"context"
	"fmt"

	"gin/internal/domain/game"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/id"
	"gin/internal/support/message"
	"gin/internal/ws"
)

type GameSessionService struct {
	hub              *ws.Hub
	walletRepository *repopg.WalletRepository
}

func NewGameSessionService(hub *ws.Hub, walletRepository *repopg.WalletRepository) *GameSessionService {
	return &GameSessionService{hub: hub, walletRepository: walletRepository}
}

func (s *GameSessionService) JoinGame(ctx context.Context, gameType game.GameType, userID string) (game.JoinResponse, error) {
	if userID == "" {
		return game.JoinResponse{}, fmt.Errorf(message.UserIDRequired)
	}

	connectionID := id.New()
	now := clock.Now()

	s.hub.Upsert(ws.Session{
		ConnectionID: connectionID,
		UserID:       userID,
		GameType:     string(gameType),
		JoinedAt:     now,
		LastSeenAt:   now,
	})

	return game.JoinResponse{
		ConnectionID: connectionID,
		GameType:     gameType,
		JoinedAt:     now,
		Message:      message.JoinGameSuccess,
	}, nil
}

func (s *GameSessionService) ValidateConnection(connectionID, userID string, gameType game.GameType) error {
	session, ok := s.hub.Get(connectionID)
	if !ok {
		return fmt.Errorf(message.ConnectionSessionMissing)
	}

	if session.UserID != userID {
		return fmt.Errorf(message.ConnectionUserMismatch)
	}

	if session.GameType != string(gameType) {
		return fmt.Errorf(message.ConnectionGameMismatch)
	}

	return nil
}
