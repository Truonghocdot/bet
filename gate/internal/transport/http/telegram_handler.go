package http

import (
	"crypto/subtle"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"gate/internal/service"
)

type TelegramHandler struct {
	notifier      *service.TelegramNotifier
	internalToken string
}

func NewTelegramHandler(notifier *service.TelegramNotifier, internalToken string) *TelegramHandler {
	return &TelegramHandler{notifier: notifier, internalToken: strings.TrimSpace(internalToken)}
}

func (h *TelegramHandler) Test(w http.ResponseWriter, r *http.Request) {
	if h.notifier == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "telegram notifier is unavailable"})
		return
	}
	if h.internalToken == "" || r.Header.Get("X-Internal-Token") != h.internalToken {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "internal token invalid"})
		return
	}
	var request struct {
		SiteCode string `json:"site_code"`
		ChatID   int64  `json:"chat_id"`
		Message  string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid telegram test payload"})
		return
	}
	if strings.TrimSpace(request.SiteCode) == "" || request.SiteCode != h.notifier.SiteCode() || request.ChatID == 0 || strings.TrimSpace(request.Message) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "site_code, chat_id and message are required"})
		return
	}
	if err := h.notifier.SendTest(r.Context(), request.ChatID, request.Message); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"message": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *TelegramHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.notifier == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"message": "telegram notifier is unavailable"})
		return
	}
	secret := h.notifier.WebhookSecret()
	provided := strings.TrimSpace(r.Header.Get("X-Telegram-Bot-Api-Secret-Token"))
	if secret == "" || provided == "" || subtle.ConstantTimeCompare([]byte(secret), []byte(provided)) != 1 {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "telegram webhook secret invalid"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid telegram update"})
		return
	}
	if err := h.notifier.HandleWebhookUpdate(r.Context(), body); err != nil {
		log.Printf("[telegram][webhook.error] err=%v", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
