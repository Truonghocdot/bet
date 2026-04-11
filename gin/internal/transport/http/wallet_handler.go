package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type WalletHandler struct {
	walletService *service.WalletService
}

func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{walletService: walletService}
}

func (h *WalletHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
		return
	}

	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	response, err := h.walletService.Summary(r.Context(), claims.UserID)
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

	pollTicker := time.NewTicker(5 * time.Second)
	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer pollTicker.Stop()
	defer heartbeatTicker.Stop()

	lastPayload := ""
	emitSummary := func(response any) error {
		payload, err := json.Marshal(response)
		if err != nil {
			return err
		}
		payloadKey := string(payload)
		if payloadKey == lastPayload {
			return nil
		}

		lastPayload = payloadKey
		return stream.Event("wallet.summary", response)
	}

	if err := emitSummary(initialResponse); err != nil {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-pollTicker.C:
			response, err := h.walletService.Summary(r.Context(), claims.UserID)
			if err != nil {
				return
			}
			if err := emitSummary(response); err != nil {
				return
			}
		case <-heartbeatTicker.C:
			if err := stream.KeepAlive(); err != nil {
				return
			}
		}
	}
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
