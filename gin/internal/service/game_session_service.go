package service

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	repopg "gin/internal/repository/postgres"
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

	if s.walletRepository != nil {
		userNumericID, err := strconv.ParseInt(strings.TrimSpace(userID), 10, 64)
		if err != nil {
			return game.JoinResponse{}, fmt.Errorf(message.Unauthorized)
		}

		record, err := s.walletRepository.FindByUserAndUnit(ctx, userNumericID, 1)
		if err != nil {
			return game.JoinResponse{}, err
		}

		available, err := subtractDecimal(record.Balance, record.LockedBalance)
		if err != nil {
			return game.JoinResponse{}, err
		}

		if compareDecimal(available, "0") <= 0 {
			return game.JoinResponse{}, fmt.Errorf(message.InsufficientBalancePlay)
		}
	}

	connectionID := id.New()
	now := time.Now()

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
