package http

import (
	"net/http"

	"gate/internal/service"
)

func NewRouter(webhookService *service.WebhookService, notificationService *service.NotificationService, tcgService *service.TCGService) http.Handler {
	mux := http.NewServeMux()

	healthHandler := NewHealthHandler()
	webhookHandler := NewWebhookHandler(webhookService)
	nowPaymentsHandler := NewNowPaymentsHandler(webhookService)
	notificationHandler := NewNotificationHandler(notificationService)
	tcgHandler := NewTCGHandler(tcgService, webhookService.InternalToken())

	mux.HandleFunc("GET /healthz", healthHandler.ServeHTTP)
	mux.HandleFunc("POST /v1/webhooks/deposits/", webhookHandler.ServeHTTP)
	mux.HandleFunc("POST /internal/v1/nowpayments/deposits/create", nowPaymentsHandler.CreateDeposit)
	mux.HandleFunc("POST /internal/v1/tcg/players/register", tcgHandler.RegisterPlayer)
	mux.HandleFunc("POST /internal/v1/tcg/wallets/balance", tcgHandler.Balance)
	mux.HandleFunc("POST /internal/v1/tcg/wallets/transfer", tcgHandler.Transfer)
	mux.HandleFunc("POST /internal/v1/tcg/wallets/transfer-out-all", tcgHandler.TransferOutAll)
	mux.HandleFunc("POST /internal/v1/tcg/wallets/transfer-status", tcgHandler.TransferStatus)
	mux.HandleFunc("POST /internal/v1/tcg/games/launch", tcgHandler.LaunchGame)
	mux.HandleFunc("POST /v1/notifications/", notificationHandler.ServeHTTP)

	return RecoverMiddleware(mux)
}
