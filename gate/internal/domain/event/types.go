package event

import "time"

type WebhookEvent struct {
	Provider   string         `json:"provider"`
	Type       string         `json:"type"`
	ReceivedAt time.Time      `json:"received_at"`
	Payload    map[string]any `json:"payload"`
}

type NotificationRequest struct {
	Channel string         `json:"channel"`
	Target  string         `json:"target"`
	Subject string         `json:"subject"`
	Message string         `json:"message"`
	Meta    map[string]any `json:"meta"`
}

type DepositApplyRequest struct {
	Provider       string         `json:"provider"`
	ProviderStatus string         `json:"provider_status"`
	ClientRef      string         `json:"client_ref"`
	ProviderTxnID  string         `json:"provider_txn_id"`
	Amount         string         `json:"amount"`
	Currency       string         `json:"currency"`
	PaidAt         time.Time      `json:"paid_at"`
	Raw            map[string]any `json:"raw"`
}
