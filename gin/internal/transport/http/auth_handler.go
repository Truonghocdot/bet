package http

import (
	"encoding/json"
	"errors"
	"net/http"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/auth"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidRegisterPayload})
		return
	}

	response, err := h.authService.Register(r.Context(), request, extractRequestMeta(r))
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidLoginPayload})
		return
	}

	response, err := h.authService.Login(r.Context(), request, extractRequestMeta(r))
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	profile, err := h.authService.Me(r.Context(), claims.UserID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var request auth.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidForgotPayload})
		return
	}

	response, err := h.authService.ForgotPassword(r.Context(), request, extractRequestMeta(r))
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusAccepted, response)
}

func (h *AuthHandler) VerifyForgotPasswordOTP(w http.ResponseWriter, r *http.Request) {
	var request auth.VerifyForgotPasswordOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidVerifyOTPPayload})
		return
	}

	response, err := h.authService.VerifyForgotPasswordOTP(r.Context(), request)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var request auth.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidResetPayload})
		return
	}

	response, err := h.authService.ResetPassword(r.Context(), request)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repopg.ErrEmailExists),
		errors.Is(err, repopg.ErrPhoneExists),
		errors.Is(err, repopg.ErrRefCodeNotFound),
		errors.Is(err, repopg.ErrInvalidSelfReferral):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.InvalidCredentials})
	case errors.Is(err, service.ErrRateLimited):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"message": message.TooManyRequests})
	case errors.Is(err, service.ErrLoginLocked):
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"message": message.LoginTemporarilyLocked})
	case errors.Is(err, service.ErrOTPInvalid),
		errors.Is(err, service.ErrOTPExpired),
		errors.Is(err, service.ErrOTPLocked),
		errors.Is(err, service.ErrResetTokenInvalid),
		errors.Is(err, repopg.ErrOTPNotFound),
		errors.Is(err, repopg.ErrOTPExpired),
		errors.Is(err, repopg.ErrOTPLocked),
		errors.Is(err, repopg.ErrResetTokenInvalid):
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
	case errors.Is(err, repopg.ErrAccountNotFound), errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
	case errors.Is(err, repopg.ErrUserDisabled):
		writeJSON(w, http.StatusForbidden, map[string]string{"message": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
	}
}
