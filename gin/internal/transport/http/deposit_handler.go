package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
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
	path := strings.TrimPrefix(r.URL.Path, "/v1/deposits")
	path = strings.Trim(path, "/")
	parts := make([]string, 0)
	if path != "" {
		parts = strings.Split(path, "/")
	}

	if len(parts) == 0 && r.Method == http.MethodGet {
		h.handleListHistory(w, r)
		return
	}

	if len(parts) == 1 && r.Method == http.MethodGet {
		h.handleStatus(w, r, parts[0])
		return
	}

	if len(parts) == 2 && parts[1] == "stream" && r.Method == http.MethodGet {
		h.handleStatusStream(w, r, parts[0])
		return
	}

	if len(parts) == 2 && parts[1] == "banks" && r.Method == http.MethodGet {
		h.handleBanks(w, r, parts[0])
		return
	}

	if len(parts) == 2 && parts[1] == "cancel" && r.Method == http.MethodPost {
		h.handleCancel(w, r, parts[0])
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

func (h *DepositHandler) handleListHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	page, pageSize := parsePaginationQuery(r, 10)
	response, err := h.depositService.ListHistory(r.Context(), claims.UserID, page, pageSize)
	if err != nil {
		h.writeError(r, w, err, "list_history")
		return
	}

	if response.Data == nil {
		response.Data = []deposit.DepositTransaction{}
	}

	type depositHistoryPublic struct {
		ID        int64     `json:"id"`
		Unit      int       `json:"unit"`
		Amount    string    `json:"amount"`
		Status    int       `json:"status"`
		CreatedAt time.Time `json:"created_at"`
	}

	out := make([]depositHistoryPublic, 0, len(response.Data))
	for _, it := range response.Data {
		out = append(out, depositHistoryPublic{
			ID:        it.ID,
			Unit:      it.Unit,
			Amount:    it.Amount,
			Status:    it.Status,
			CreatedAt: it.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"page":        response.Page,
		"page_size":   response.PageSize,
		"total":       response.Total,
		"total_pages": response.TotalPages,
		"data":        out,
	})
}

func parsePaginationQuery(r *http.Request, defaultPageSize int) (int, int) {
	page := 1
	pageSize := defaultPageSize

	if value := strings.TrimSpace(r.URL.Query().Get("page")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("page_size")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}
	if value := strings.TrimSpace(r.URL.Query().Get("offset")); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 && pageSize > 0 {
			page = (parsed / pageSize) + 1
		}
	}

	return page, pageSize
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
		h.writeError(r, w, err, "apply")
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
		h.writeError(r, w, err, "init_vietqr")
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
		h.writeError(r, w, err, "init_usdt")
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *DepositHandler) handleBanks(w http.ResponseWriter, r *http.Request, method string) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	switch method {
	case string(deposit.DepositMethodVietQR):
		response, err := h.depositService.ListVietQrBanks(r.Context())
		if err != nil {
			h.writeError(r, w, err, "list_banks")
			return
		}

		writeJSON(w, http.StatusOK, response)
	default:
		_ = claims
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
	}
}

func (h *DepositHandler) handleStatus(w http.ResponseWriter, r *http.Request, clientRef string) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	response, err := h.depositService.GetDepositStatus(r.Context(), claims.UserID, clientRef)
	if err != nil {
		h.writeError(r, w, err, "status")
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *DepositHandler) handleCancel(w http.ResponseWriter, r *http.Request, idStr string) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var txnID int64
	if _, err := fmt.Sscanf(idStr, "%d", &txnID); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "ID giao dịch không hợp lệ"})
		return
	}

	response, err := h.depositService.CancelDeposit(r.Context(), claims.UserID, txnID)
	if err != nil {
		h.writeError(r, w, err, "cancel")
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
		h.writeError(r, w, err, "status_stream_init")
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

func (h *DepositHandler) writeError(r *http.Request, w http.ResponseWriter, err error, operation string) {
	statusCode := http.StatusInternalServerError
	messageText := message.InternalServerError

	switch {
	case errors.Is(err, repopg.ErrDepositNotFound):
		statusCode = http.StatusNotFound
		messageText = message.DepositIntentNotFound
	case errors.Is(err, repopg.ErrDepositAlreadyDone):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositAlreadyCompleted
	case errors.Is(err, repopg.ErrDepositProviderInvalid):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositProviderInvalid
	case errors.Is(err, repopg.ErrDepositReceivingAccount):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositReceivingAccountMissing
	case errors.Is(err, repopg.ErrDepositWalletNotFound):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositWalletMissing
	case errors.Is(err, repopg.ErrDepositAmountInvalid):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositAmountInvalid
	case errors.Is(err, service.ErrDepositUSDTNotAvailable):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositUSDTNotAvailable
	case errors.Is(err, service.ErrDepositUSDTTemporarilyClosed):
		statusCode = http.StatusUnprocessableEntity
		messageText = message.DepositUSDTTemporarilyClosed
	case errors.Is(err, repopg.ErrDepositCancelForbidden):
		statusCode = http.StatusUnprocessableEntity
		messageText = err.Error()
	}

	log.Printf(
		"[deposit][handler.error] op=%s method=%s path=%s status=%d err=%+v",
		operation,
		r.Method,
		r.URL.Path,
		statusCode,
		err,
	)

	// Trả về message lỗi chi tiết hơn nếu là môi trường local/dev
	// hoặc giúp admin debug nhanh hơn lúc này.
	finalMessage := messageText
	if statusCode == http.StatusInternalServerError {
		finalMessage = fmt.Sprintf("%s: %v", message.InternalServerError, err)
	}

	writeJSON(w, statusCode, map[string]any{"message": finalMessage, "error_detail": err.Error()})
}
