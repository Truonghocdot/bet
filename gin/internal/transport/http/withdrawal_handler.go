package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/withdrawal"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type WithdrawalHandler struct {
	withdrawalService *service.WithdrawalService
}

func NewWithdrawalHandler(withdrawalService *service.WithdrawalService) *WithdrawalHandler {
	return &WithdrawalHandler{
		withdrawalService: withdrawalService,
	}
}

func (h *WithdrawalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/withdrawals")
	if path != "" && path != "/" {
		path = strings.TrimSuffix(path, "/")
	}
	
	if path == "/accounts" {
		if r.Method == http.MethodGet {
			h.handleListAccounts(w, r)
			return
		}
		if r.Method == http.MethodPost {
			h.handleAddAccount(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/accounts/") && r.Method == http.MethodDelete {
		idStr := strings.TrimPrefix(path, "/accounts/")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			h.handleDeleteAccount(w, r, id)
			return
		}
	}

	if (path == "" || path == "/") && r.Method == http.MethodPost {
		h.handleSubmitWithdrawal(w, r)
		return
	}

	if (path == "" || path == "/") && r.Method == http.MethodGet {
		h.handleListHistory(w, r)
		return
	}

	writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
}

func (h *WithdrawalHandler) handleListHistory(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	requests, err := h.withdrawalService.ListHistory(r.Context(), claims.UserID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	if requests == nil {
		requests = []withdrawal.WithdrawalRequest{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": requests})
}

func (h *WithdrawalHandler) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	accounts, err := h.withdrawalService.ListUserAccounts(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	// Always return an array
	if accounts == nil {
		accounts = []withdrawal.AccountWithdrawalInfo{}
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": accounts})
}

func (h *WithdrawalHandler) handleAddAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var req withdrawal.SetupAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid payload format"})
		return
	}

	if req.Unit != 1 && req.Unit != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Unit must be 1 (VND) or 2 (USDT)"})
		return
	}

	if strings.TrimSpace(req.AccountNumber) == "" || strings.TrimSpace(req.AccountName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Missing required fields"})
		return
	}

	id, err := h.withdrawalService.AddAccount(r.Context(), claims.UserID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "message": "Thêm phương thức thành công"})
}

func (h *WithdrawalHandler) handleDeleteAccount(w http.ResponseWriter, r *http.Request, id int64) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	err := h.withdrawalService.DeleteAccount(r.Context(), claims.UserID, id)
	if err != nil {
		if errors.Is(err, repopg.ErrWithdrawalAccountNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Tài khoản không tồn tại"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Xoá phương thức thành công"})
}

func (h *WithdrawalHandler) handleSubmitWithdrawal(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var req withdrawal.SubmitWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid payload format"})
		return
	}

	requestID, err := h.withdrawalService.SubmitWithdrawalRequest(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, repopg.ErrInsufficientBalance) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Số dư không đủ để rút số tiền này"})
			return
		}
		if errors.Is(err, repopg.ErrWithdrawalAccountNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"message": "Tài khoản nhận tiền không tồn tại"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":      requestID,
		"message": "Đã tạo lệnh rút tiền, vui lòng chờ hệ thống xét duyệt.",
	})
}
