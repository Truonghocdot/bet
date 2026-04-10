package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"gate/internal/domain/event"
)

type WebhookService struct{}

func NewWebhookService() *WebhookService {
	return &WebhookService{}
}

func (s *WebhookService) HandleDepositWebhook(_ context.Context, provider string, payload map[string]any) (event.WebhookEvent, error) {
	if provider == "" {
		return event.WebhookEvent{}, fmt.Errorf("provider is required")
	}

	webhookEvent := event.WebhookEvent{
		Provider:   provider,
		Type:       "deposit.callback",
		ReceivedAt: time.Now(),
		Payload:    payload,
	}

	log.Printf("[gate] webhook provider=%s payload=%v", provider, payload)

	return webhookEvent, nil
}
