package http

import (
	"net/http"

	"gate/internal/service"
)

func NewRouter(webhookService *service.WebhookService, notificationService *service.NotificationService) http.Handler {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler()
	webhookHandler := NewWebhookHandler(webhookService)
	nowPaymentsHandler := NewNowPaymentsHandler(webhookService)
	notificationHandler := NewNotificationHandler(notificationService)

	mux.HandleFunc("GET /healthz", healthHandler.ServeHTTP)
	mux.HandleFunc("POST /v1/webhooks/deposits/", webhookHandler.ServeHTTP)
	mux.HandleFunc("POST /internal/v1/nowpayments/deposits/create", nowPaymentsHandler.CreateDeposit)
	mux.HandleFunc("POST /v1/notifications/", notificationHandler.ServeHTTP)

	return RecoverMiddleware(mux)
}
