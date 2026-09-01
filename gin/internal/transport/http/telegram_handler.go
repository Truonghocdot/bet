package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"gin/internal/domain/telegram"
	"gin/internal/service"
)

type TelegramHandler struct {
	service        *service.TelegramService
	depositService *service.DepositService
	internalToken  string
}

func NewTelegramHandler(service *service.TelegramService, depositService *service.DepositService, internalToken string) *TelegramHandler {
	return &TelegramHandler{service: service, depositService: depositService, internalToken: strings.TrimSpace(internalToken)}
}

func (h *TelegramHandler) authorize(w http.ResponseWriter, r *http.Request) bool {
	if h.internalToken == "" || r.Header.Get("X-Internal-Token") != h.internalToken {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return false
	}
	return true
}

func (h *TelegramHandler) GroupEvent(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	var request telegram.GroupEvent
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid telegram group event"})
		return
	}
	if err := h.service.RecordGroupEvent(r.Context(), request); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}

func (h *TelegramHandler) Targets(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	targets, err := h.service.ListActiveTargets(r.Context(), r.URL.Query().Get("site_code"))
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"targets": targets})
}

func (h *TelegramHandler) TargetError(w http.ResponseWriter, r *http.Request) {
	if !h.authorize(w, r) {
		return
	}
	var request telegram.TargetError
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid telegram target error"})
		return
	}
	if err := h.service.MarkTargetError(r.Context(), request); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]bool{"ok": true})
}
