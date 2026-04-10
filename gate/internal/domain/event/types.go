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
