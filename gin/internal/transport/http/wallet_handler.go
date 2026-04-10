package http

import (
	"errors"
	"net/http"

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
