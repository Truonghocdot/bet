package http

import (
	"encoding/json"
	"log"
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
		log.Printf("[gate][nowpayments.create.error] reason=unauthorized path=%s", r.URL.Path)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}

	var request service.CreateNowPaymentsDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("[gate][nowpayments.create.error] reason=decode_failed path=%s err=%v", r.URL.Path, err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid nowpayments payload"})
		return
	}

	response, err := h.webhookService.CreateNowPaymentsDeposit(r.Context(), request)
	if err != nil {
		log.Printf("[gate][nowpayments.create.error] reason=create_failed client_ref=%s err=%v", strings.TrimSpace(request.ClientRef), err)
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
