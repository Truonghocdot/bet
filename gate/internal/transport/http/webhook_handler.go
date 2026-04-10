package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"gate/internal/service"
)

type WebhookHandler struct {
	webhookService *service.WebhookService
}

func NewWebhookHandler(webhookService *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	provider := strings.TrimPrefix(r.URL.Path, "/v1/webhooks/deposits/")
	if provider == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "provider not found"})
		return
	}

	var payload map[string]any
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid webhook payload"})
		return
	}

	event, err := h.webhookService.HandleDepositWebhook(r.Context(), provider, payload)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, event)
}
