package http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/chat"
	"gin/internal/realtime"
	"gin/internal/service"

	"github.com/gorilla/websocket"
)

type ChatHandler struct {
	service *service.ChatService
	broker  *realtime.Broker
}

func NewChatHandler(chatService *service.ChatService, broker *realtime.Broker) *ChatHandler {
	return &ChatHandler{service: chatService, broker: broker}
}

func (h *ChatHandler) List(w http.ResponseWriter, r *http.Request) {
	before, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("before")), 10, 64)
	limit, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("limit")), 10, 64)
	response, err := h.service.List(r.Context(), before, limit)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *ChatHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Vui lòng đăng nhập."})
		return
	}
	var payload chat.CreateRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8*1024))
	if err := decoder.Decode(&payload); err != nil {
		h.writeError(w, service.ErrChatInvalid)
		return
	}
	item, err := h.service.Create(r.Context(), claims.UserID, chatClientIP(r), payload.Body)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, chat.CreateResponse{Message: item})
}

func (h *ChatHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	if _, err := h.service.List(r.Context(), 0, 1); err != nil {
		h.writeError(w, err)
		return
	}
	conn, err := chatWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(75 * time.Second)) })
	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.ChatGlobalTopic(h.service.RoomCode()))
	if err != nil {
		return
	}
	defer unsubscribe()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-done:
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if err := conn.WriteJSON(map[string]any{"event": update.Event, "data": json.RawMessage(update.Data)}); err != nil {
				return
			}
		case <-ping.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)); err != nil {
				return
			}
		}
	}
}

func (h *ChatHandler) writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "Không thể tải phòng chat."
	switch {
	case errors.Is(err, service.ErrChatDisabled):
		status = http.StatusNotFound
		message = "Phòng chat hiện chưa mở."
	case errors.Is(err, service.ErrChatBanned):
		status = http.StatusForbidden
		message = "Bạn đang bị khóa quyền chat."
	case errors.Is(err, service.ErrChatRateLimited):
		status = http.StatusTooManyRequests
		message = "Bạn gửi tin quá nhanh, vui lòng thử lại sau."
	case errors.Is(err, service.ErrChatInvalid):
		status = http.StatusBadRequest
		message = "Tin nhắn không hợp lệ."
	}
	writeJSON(w, status, map[string]string{"message": message, "code": err.Error()})
}

func chatClientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); forwarded != "" {
		return forwarded
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host
	}
	return strings.TrimSpace(r.RemoteAddr)
}

var chatWSUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 2048, CheckOrigin: func(*http.Request) bool { return true }}
