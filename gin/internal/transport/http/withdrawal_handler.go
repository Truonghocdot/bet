package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	type withdrawalPublic struct {
		ID        int64     `json:"id"`
		Unit      int       `json:"unit"`
		Amount    string    `json:"amount"`
		Fee       string    `json:"fee"`
		NetAmount string    `json:"net_amount"`
		Status    int       `json:"status"`
		Reason    string    `json:"reason_rejected,omitempty"`
		CreatedAt time.Time `json:"created_at"`
	}

	out := make([]withdrawalPublic, 0, len(requests))
	for _, it := range requests {
		out = append(out, withdrawalPublic{
			ID:        it.ID,
			Unit:      it.Unit,
			Amount:    it.Amount,
			Fee:       it.Fee,
			NetAmount: it.NetAmount,
			Status:    it.Status,
			Reason:    it.ReasonRejected,
			CreatedAt: it.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": out})
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

	// Client không được thấy tên ngân hàng/tên chủ tài khoản để tránh tự chỉnh sửa qua F12.
	type accountPublic struct {
		ID            int64     `json:"id"`
		Unit          int       `json:"unit"`
		AccountNumber string    `json:"account_number"`
		IsDefault     bool      `json:"is_default"`
		CreatedAt     time.Time `json:"created_at"`
	}

	out := make([]accountPublic, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, accountPublic{
			ID:            a.ID,
			Unit:          a.Unit,
			AccountNumber: a.AccountNumber,
			IsDefault:     a.IsDefault,
			CreatedAt:     a.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": out})
}

func (h *WithdrawalHandler) handleAddAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var req withdrawal.SetupAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Dữ liệu tài khoản rút tiền không hợp lệ"})
		return
	}

	if req.Unit != 1 && req.Unit != 2 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Loại ví không hợp lệ"})
		return
	}

	if strings.TrimSpace(req.AccountNumber) == "" || strings.TrimSpace(req.AccountName) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Thiếu thông tin tài khoản nhận tiền"})
		return
	}

	// Không cho phép khách tự sửa thông tin: nếu đã liên kết 1 tài khoản cho unit này thì chặn tạo mới.
	existing, err := h.withdrawalService.ListUserAccounts(r.Context(), claims.UserID)
	if err == nil {
		for _, a := range existing {
			if a.Unit == req.Unit {
				writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": "Bạn đã liên kết tài khoản nhận tiền, không thể thay đổi."})
				return
			}
		}
	}

	id, err := h.withdrawalService.AddAccount(r.Context(), claims.UserID, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "message": "Thêm phương thức thành công"})
}

func (h *WithdrawalHandler) handleDeleteAccount(w http.ResponseWriter, r *http.Request, id int64) {
	_ = id
	// Không cho phép khách tự xoá/sửa thông tin nhận tiền sau khi đã liên kết.
	writeJSON(w, http.StatusForbidden, map[string]string{"message": "Bạn không được phép thay đổi tài khoản nhận tiền."})
}

func (h *WithdrawalHandler) handleSubmitWithdrawal(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var req withdrawal.SubmitWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Lệnh rút không hợp lệ"})
		return
	}
	if req.AccountWithdrawalInfoID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Lệnh rút không hợp lệ"})
		return
	}
	if strings.TrimSpace(req.Amount) == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Lệnh rút không hợp lệ"})
		return
	}

	requestID, err := h.withdrawalService.SubmitWithdrawalRequest(r.Context(), claims.UserID, req)
	if err != nil {
		if errors.Is(err, repopg.ErrWithdrawalAccountNotFound) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Lệnh rút không hợp lệ"})
			return
		}
		if errors.Is(err, repopg.ErrInsufficientBalance) {
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Lệnh rút không hợp lệ"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Lệnh rút không hợp lệ"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":      requestID,
		"message": "Đã tạo lệnh rút tiền, vui lòng chờ hệ thống xét duyệt.",
	})
}
