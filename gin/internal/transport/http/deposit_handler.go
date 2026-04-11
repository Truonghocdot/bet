package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/deposit"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type DepositHandler struct {
	depositService *service.DepositService
	internalToken  string
}

func NewDepositHandler(depositService *service.DepositService, internalToken string) *DepositHandler {
	return &DepositHandler{
		depositService: depositService,
		internalToken:  internalToken,
	}
}

func (h *DepositHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/deposits/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 1 && r.Method == http.MethodGet {
		h.handleStatus(w, r, parts[0])
		return
	}

	if len(parts) == 2 && parts[1] == "stream" && r.Method == http.MethodGet {
		h.handleStatusStream(w, r, parts[0])
		return
	}

	if len(parts) != 2 || parts[1] != "init" || r.Method != http.MethodPost {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
		return
	}

	switch parts[0] {
	case string(deposit.DepositMethodVietQR):
		h.handleInitVietQR(w, r)
	case string(deposit.DepositMethodUSDT):
		h.handleInitUSDT(w, r)
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
	}
}

func (h *DepositHandler) Apply(w http.ResponseWriter, r *http.Request) {
	if h.internalToken == "" || r.Header.Get("X-Internal-Token") != h.internalToken {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.DepositInternalTokenInvalid})
		return
	}

	var request deposit.ApplyDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidDepositPayload})
		return
	}

	response, err := h.depositService.ApplyDeposit(r.Context(), request)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, response)
}

func (h *DepositHandler) handleInitVietQR(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var request deposit.DepositInitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidDepositPayload})
		return
	}

	response, err := h.depositService.InitVietQRDeposit(r.Context(), claims.UserID, request)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *DepositHandler) handleInitUSDT(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var request deposit.DepositInitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidDepositPayload})
		return
	}

	response, err := h.depositService.InitUSDTDeposit(r.Context(), claims.UserID, request)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *DepositHandler) handleStatus(w http.ResponseWriter, r *http.Request, clientRef string) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	response, err := h.depositService.GetDepositStatus(r.Context(), claims.UserID, clientRef)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *DepositHandler) handleStatusStream(w http.ResponseWriter, r *http.Request, clientRef string) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	initialResponse, err := h.depositService.GetDepositStatus(r.Context(), claims.UserID, clientRef)
	if err != nil {
		h.writeError(w, err)
		return
	}

	stream, err := newSSEStream(w)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	pollTicker := time.NewTicker(3 * time.Second)
	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer pollTicker.Stop()
	defer heartbeatTicker.Stop()

	lastPayload := ""
	emitStatus := func(response deposit.DepositStatusResponse) (bool, error) {
		payload, err := json.Marshal(response)
		if err != nil {
			return false, err
		}
		payloadKey := string(payload)
		if payloadKey != lastPayload {
			lastPayload = payloadKey
			if err := stream.Event("deposit.status", response); err != nil {
				return false, err
			}
		}

		terminal := response.Transaction.Status == 2 || response.Transaction.Status == 3 || response.Transaction.Status == 4
		return terminal, nil
	}

	terminal, err := emitStatus(initialResponse)
	if err != nil {
		return
	}
	if terminal {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-pollTicker.C:
			response, err := h.depositService.GetDepositStatus(r.Context(), claims.UserID, clientRef)
			if err != nil {
				return
			}
			terminal, err := emitStatus(response)
			if err != nil {
				return
			}
			if terminal {
				return
			}
		case <-heartbeatTicker.C:
			if err := stream.KeepAlive(); err != nil {
				return
			}
		}
	}
}

func (h *DepositHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repopg.ErrDepositNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"message": err.Error()})
	case errors.Is(err, repopg.ErrDepositProviderInvalid),
		errors.Is(err, repopg.ErrDepositReceivingAccount),
		errors.Is(err, repopg.ErrDepositWalletNotFound),
		errors.Is(err, repopg.ErrDepositAmountInvalid):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
	}
}
