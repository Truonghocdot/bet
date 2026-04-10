package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"gate/internal/domain/event"
	"gate/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	channel := strings.TrimPrefix(r.URL.Path, "/v1/notifications/")
	if channel == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": "channel not found"})
		return
	}

	var request event.NotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid notification payload"})
		return
	}

	request.Channel = channel

	if err := h.notificationService.Send(r.Context(), request); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"channel": request.Channel,
		"target":  request.Target,
		"status":  "accepted",
	})
}
