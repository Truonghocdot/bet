package http

import (
	"encoding/json"
	"io"
	"log"
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

	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[gate][webhook.error] provider=%s reason=read_body_failed err=%v", provider, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid webhook payload"})
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		log.Printf("[gate][webhook.error] provider=%s reason=unmarshal_failed err=%v", provider, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid webhook payload"})
		return
	}

	event, err := h.webhookService.HandleDepositWebhook(r.Context(), provider, payload, rawBody, r.Header)
	if err != nil {
		log.Printf("[gate][webhook.error] provider=%s reason=handle_failed err=%v", provider, err)
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, event)
}
