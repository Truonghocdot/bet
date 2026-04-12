package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"gate/internal/service"
)

type NowPaymentsHandler struct {
	webhookService *service.WebhookService
}

func NewNowPaymentsHandler(webhookService *service.WebhookService) *NowPaymentsHandler {
	return &NowPaymentsHandler{webhookService: webhookService}
}

func (h *NowPaymentsHandler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthorized(r.Header.Get("X-Internal-Token")) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.CreateNowPaymentsDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid nowpayments payload"})
		return
	}

	response, err := h.webhookService.CreateNowPaymentsDeposit(r.Context(), request)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *NowPaymentsHandler) isAuthorized(token string) bool {
	expected := strings.TrimSpace(h.webhookService.InternalToken())
	if expected == "" {
		return false
	}

	return strings.TrimSpace(token) == expected
}
