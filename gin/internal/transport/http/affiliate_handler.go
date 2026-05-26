package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	authmiddleware "gin/internal/auth/middleware"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type AffiliateHandler struct {
	affiliateService *service.AffiliateService
}

func NewAffiliateHandler(affiliateService *service.AffiliateService) *AffiliateHandler {
	return &AffiliateHandler{affiliateService: affiliateService}
}

func (h *AffiliateHandler) Summary(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	res, err := h.affiliateService.Summary(r.Context(), claims.UserID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *AffiliateHandler) ManagedUsers(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	res, err := h.affiliateService.ManagedUsers(r.Context(), claims.UserID, claims.Role)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			writeJSON(w, http.StatusForbidden, map[string]string{"message": "Bạn không có quyền xem danh sách user trực thuộc"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *AffiliateHandler) ManagedUserTransactions(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	userID, err := strconv.ParseInt(r.PathValue("user_id"), 10, 64)
	if err != nil || userID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Mã người chơi không hợp lệ"})
		return
	}

	res, err := h.affiliateService.ManagedUserTransactions(r.Context(), claims.UserID, claims.Role, userID)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			writeJSON(w, http.StatusForbidden, map[string]string{"message": "Bạn không có quyền xem giao dịch của người chơi này"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	writeJSON(w, http.StatusOK, res)
}

type becomeAgencyRequest struct {
	StaffRefCode string `json:"staff_ref_code"`
}

func (h *AffiliateHandler) BecomeAgency(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	var req becomeAgencyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Dữ liệu nâng cấp đại lý không hợp lệ"})
		return
	}

	res, err := h.affiliateService.BecomeAgency(r.Context(), claims.UserID, claims.Role, req.StaffRefCode)
	if err != nil {
		switch {
		case errors.Is(err, repopg.ErrStaffInviteInvalid):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": message.StaffInviteCodeInvalid})
			return
		case errors.Is(err, repopg.ErrReferralAlreadyUsed):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": message.ReferralAlreadyUsed})
			return
		case errors.Is(err, repopg.ErrInvalidSelfReferral):
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": message.InvalidSelfReferral})
			return
		case errors.Is(err, service.ErrUnauthorized):
			writeJSON(w, http.StatusForbidden, map[string]string{"message": "Bạn không có quyền thực hiện thao tác này"})
			return
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
			return
		}
	}

	writeJSON(w, http.StatusOK, res)
}
