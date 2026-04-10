package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gate/internal/domain/event"
	ginclient "gate/internal/integration/gin"
)

type WebhookService struct {
	ginClient *ginclient.Client
}

func NewWebhookService(ginClient *ginclient.Client) *WebhookService {
	return &WebhookService{ginClient: ginClient}
}

func (s *WebhookService) HandleDepositWebhook(ctx context.Context, provider string, payload map[string]any) (event.WebhookEvent, error) {
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

	if s.ginClient != nil {
		if err := s.forwardToGin(ctx, provider, payload); err != nil {
			return webhookEvent, err
		}
	}

	return webhookEvent, nil
}

func (s *WebhookService) forwardToGin(ctx context.Context, provider string, payload map[string]any) error {
	request := event.DepositApplyRequest{
		Provider:       provider,
		ProviderStatus: firstNonEmptyString(payload, []string{"provider_status", "status", "state", "payment_status", "code"}),
		ClientRef:      firstNonEmptyString(payload, []string{"client_ref", "order_code", "orderCode", "reference", "ref"}),
		ProviderTxnID:  firstNonEmptyString(payload, []string{"provider_txn_id", "transaction_id", "transactionId", "txid", "tx_hash", "txHash"}),
		Amount:         firstNonEmptyString(payload, []string{"amount", "paid_amount", "transfer_amount"}),
		Currency:       strings.ToUpper(firstNonEmptyString(payload, []string{"currency"})),
		PaidAt:         time.Now(),
		Raw:            payload,
	}

	if request.Currency == "" {
		request.Currency = "VND"
	}

	if request.ClientRef == "" && request.ProviderTxnID == "" {
		return fmt.Errorf("client_ref or provider_txn_id is required")
	}

	return s.ginClient.ApplyDeposit(ctx, request)
}

func firstNonEmptyString(payload map[string]any, keys []string) string {
	for _, key := range keys {
		if value, ok := payload[key]; ok {
			trimmed := strings.TrimSpace(fmt.Sprint(value))
			if trimmed != "" && trimmed != "<nil>" {
				return trimmed
			}
		}
	}

	return ""
}
