package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	authmiddleware "gin/internal/auth/middleware"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/message"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	page, pageSize := readNotificationPagination(r)
	response, err := h.notificationService.List(r.Context(), claims.UserID, page, pageSize)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	rawID := strings.TrimSpace(r.PathValue("id"))
	notificationID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || notificationID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.NotificationIDInvalid})
		return
	}

	response, err := h.notificationService.MarkRead(r.Context(), claims.UserID, notificationID)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *NotificationHandler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
	case errors.Is(err, repopg.ErrNotificationNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.NotificationNotFound})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
	}
}

func readNotificationPagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 10

	if raw := strings.TrimSpace(r.URL.Query().Get("page")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			page = value
		}
	}

	if raw := strings.TrimSpace(r.URL.Query().Get("page_size")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			pageSize = value
		}
	}

	return page, pageSize
}
