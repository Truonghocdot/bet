package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/wallet"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type WalletHandler struct {
	walletService *service.WalletService
	broker        *realtime.Broker
}

func NewWalletHandler(walletService *service.WalletService, broker *realtime.Broker) *WalletHandler {
	return &WalletHandler{walletService: walletService, broker: broker}
}

func (h *WalletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
		return
	}

	var userID int64
	if claims, ok := authmiddleware.CurrentClaims(r.Context()); ok {
		userID = claims.UserID
	}

	response, err := h.walletService.Summary(r.Context(), userID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *WalletHandler) Stream(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	initialResponse, err := h.walletService.Summary(r.Context(), claims.UserID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	stream, err := newSSEStream(w)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	if err := stream.Event("wallet.summary", initialResponse); err != nil {
		return
	}

	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.WalletUserTopic(claims.UserID))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}
	defer unsubscribe()

	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Event != "wallet.summary" {
				continue
			}
			if err := stream.EventRaw(update.Event, update.Data); err != nil {
				return
			}
		case <-heartbeatTicker.C:
			if err := stream.KeepAlive(); err != nil {
				return
			}
		}
	}
}

func (h *WalletHandler) Exchange(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var req wallet.ExchangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[wallet.exchange.decode.error] err=%v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Dữ liệu không hợp lệ"})
		return
	}

	response, err := h.walletService.Exchange(r.Context(), claims.UserID, req)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *WalletHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
	case errors.Is(err, repopg.ErrAccountNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.AccountNotFound})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
	}
}
