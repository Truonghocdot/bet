package service

import (
	"context"
	"fmt"
	"strings"

	"gin/internal/domain/telegram"
	repopg "gin/internal/repository/postgres"
)

type TelegramService struct {
	repository *repopg.TelegramRepository
	siteCode   string
}

func NewTelegramService(repository *repopg.TelegramRepository, siteCode string) *TelegramService {
	return &TelegramService{repository: repository, siteCode: strings.TrimSpace(siteCode)}
}

func (s *TelegramService) RecordGroupEvent(ctx context.Context, request telegram.GroupEvent) error {
	if strings.TrimSpace(request.SiteCode) != s.siteCode {
		return fmt.Errorf("telegram site code mismatch")
	}
	if request.BotStatus != "member" && request.BotStatus != "administrator" && request.BotStatus != "left" && request.BotStatus != "kicked" {
		return fmt.Errorf("unsupported telegram bot status")
	}
	return s.repository.UpsertGroupEvent(ctx, request)
}

func (s *TelegramService) ListActiveTargets(ctx context.Context, siteCode string) ([]telegram.Target, error) {
	if strings.TrimSpace(siteCode) != s.siteCode {
		return nil, fmt.Errorf("telegram site code mismatch")
	}
	return s.repository.ListActiveTargets(ctx, siteCode)
}

func (s *TelegramService) MarkTargetError(ctx context.Context, request telegram.TargetError) error {
	if strings.TrimSpace(request.SiteCode) != s.siteCode {
		return fmt.Errorf("telegram site code mismatch")
	}
	return s.repository.MarkTargetError(ctx, request)
}
