package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/wheel"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"

	"github.com/gorilla/websocket"
)

type WheelHandler struct {
	service *service.WheelService
	broker  *realtime.Broker
}

func NewWheelHandler(wheelService *service.WheelService, broker *realtime.Broker) *WheelHandler {
	return &WheelHandler{service: wheelService, broker: broker}
}

func (h *WheelHandler) ListInvitations(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Vui lòng đăng nhập."})
		return
	}
	response, err := h.service.ListInvitations(r.Context(), claims.UserID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *WheelHandler) MarkSeen(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Vui lòng đăng nhập."})
		return
	}
	if err := h.service.MarkSeen(r.Context(), strings.TrimSpace(r.PathValue("id")), claims.UserID); err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Đã ghi nhận."})
}

func (h *WheelHandler) Launch(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Vui lòng đăng nhập."})
		return
	}
	response, err := h.service.Launch(r.Context(), strings.TrimSpace(r.PathValue("id")), claims.UserID)
	if err != nil {
		log.Printf("[wheel.launch] user_id=%d invitation=%q err=%v", claims.UserID, strings.TrimSpace(r.PathValue("id")), err)
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *WheelHandler) Exchange(w http.ResponseWriter, r *http.Request) {
	var payload wheel.ExchangeRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&payload) != nil || strings.TrimSpace(payload.LaunchCode) == "" {
		h.writeError(w, service.ErrWheelTokenInvalid)
		return
	}
	response, err := h.service.Exchange(r.Context(), payload.LaunchCode)
	if err != nil {
		log.Printf("[wheel.exchange] err=%v", err)
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *WheelHandler) MainSocketTicket(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Vui lòng đăng nhập."})
		return
	}
	response, err := h.service.CreateUserSocketTicket(r.Context(), claims.UserID)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *WheelHandler) UserEventsWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, err := h.service.ConsumeUserSocketTicket(r.Context(), r.URL.Query().Get("ticket"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	conn, err := wheelWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	items, err := h.service.ListInvitations(r.Context(), userID)
	if err == nil {
		_ = conn.WriteJSON(map[string]any{"event": "wheel.invitations", "data": items})
	}
	h.streamTopic(r, conn, realtime.UserEventTopic(userID), false)
}

func (h *WheelHandler) Me(w http.ResponseWriter, r *http.Request) {
	access, ok := h.eventAccess(w, r)
	if !ok {
		return
	}
	state, err := h.service.State(r.Context(), access)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *WheelHandler) Start(w http.ResponseWriter, r *http.Request) {
	access, ok := h.eventAccess(w, r)
	if !ok {
		return
	}
	state, err := h.service.Start(r.Context(), access)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *WheelHandler) Spin(w http.ResponseWriter, r *http.Request) {
	access, ok := h.eventAccess(w, r)
	if !ok {
		return
	}
	roundNo, err := strconv.Atoi(strings.TrimSpace(r.PathValue("round")))
	if err != nil {
		h.writeError(w, repopg.ErrWheelRoundOrder)
		return
	}
	response, err := h.service.Spin(r.Context(), access, roundNo)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *WheelHandler) SessionSocketTicket(w http.ResponseWriter, r *http.Request) {
	access, ok := h.eventAccess(w, r)
	if !ok {
		return
	}
	response, err := h.service.CreateSessionSocketTicket(r.Context(), access)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *WheelHandler) SessionWebSocket(w http.ResponseWriter, r *http.Request) {
	access, sessionID, err := h.service.ConsumeSessionSocketTicket(r.Context(), r.URL.Query().Get("ticket"))
	if err != nil {
		h.writeError(w, err)
		return
	}
	conn, err := wheelWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	if state, stateErr := h.service.State(r.Context(), access); stateErr == nil {
		_ = conn.WriteJSON(map[string]any{"event": "wheel.state", "data": state})
	}
	topic := realtime.WheelInvitationTopic(access.InvitationID)
	closeOnCompletion := false
	if sessionID != nil {
		topic = realtime.WheelSessionTopic(*sessionID)
		closeOnCompletion = true
	}
	h.streamTopic(r, conn, topic, closeOnCompletion)
}

func (h *WheelHandler) ListChat(w http.ResponseWriter, r *http.Request) {
	access, ok := h.eventAccess(w, r)
	if !ok {
		return
	}
	before, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("before")), 10, 64)
	limit, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("limit")), 10, 64)
	response, err := h.service.ListChat(r.Context(), access, before, limit)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *WheelHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	access, ok := h.eventAccess(w, r)
	if !ok {
		return
	}
	var payload wheel.ChatCreateRequest
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&payload) != nil {
		h.writeError(w, service.ErrWheelChatInvalid)
		return
	}
	response, err := h.service.CreateChat(r.Context(), access, chatClientIP(r), payload.Body)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *WheelHandler) eventAccess(w http.ResponseWriter, r *http.Request) (wheel.Access, bool) {
	parts := strings.SplitN(strings.TrimSpace(r.Header.Get("Authorization")), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		h.writeError(w, service.ErrWheelTokenInvalid)
		return wheel.Access{}, false
	}
	access, err := h.service.Authenticate(r.Context(), strings.TrimSpace(parts[1]))
	if err != nil {
		h.writeError(w, err)
		return wheel.Access{}, false
	}
	return access, true
}

func (h *WheelHandler) streamTopic(r *http.Request, conn *websocket.Conn, topic string, closeOnCompletion bool) {
	_ = conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(string) error { return conn.SetReadDeadline(time.Now().Add(75 * time.Second)) })
	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), topic)
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
			if conn.WriteJSON(map[string]any{"event": update.Event, "data": json.RawMessage(update.Data)}) != nil {
				return
			}
			if closeOnCompletion && (update.Event == "wheel.session.completed" || update.Event == "wheel.session.expired") {
				return
			}
		case <-ping.C:
			if conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second)) != nil {
				return
			}
		}
	}
}

func (h *WheelHandler) writeError(w http.ResponseWriter, err error) {
	status, code, message := http.StatusInternalServerError, "INTERNAL_ERROR", "Không thể xử lý sự kiện lúc này."
	var notReady *repopg.WheelNotReadyError
	switch {
	case errors.Is(err, service.ErrWheelDisabled):
		status, code, message = http.StatusNotFound, "WHEEL_DISABLED", "Sự kiện hiện chưa mở."
	case errors.Is(err, service.ErrWheelTokenInvalid):
		status, code, message = http.StatusUnauthorized, "WHEEL_TOKEN_INVALID", "Phiên sự kiện đã hết hạn, vui lòng mở lại từ trang chính."
	case errors.Is(err, repopg.ErrWheelInvitationNotFound):
		status, code, message = http.StatusNotFound, "INVITATION_NOT_FOUND", "Không tìm thấy lời mời."
	case errors.Is(err, repopg.ErrWheelInvitationInactive):
		status, code, message = http.StatusGone, "INVITATION_INACTIVE", "Lời mời không còn hiệu lực."
	case errors.Is(err, repopg.ErrWheelSessionNotStarted):
		status, code, message = http.StatusConflict, "SESSION_NOT_STARTED", "Vui lòng bắt đầu phiên trước."
	case errors.Is(err, repopg.ErrWheelSessionExpired):
		status, code, message = http.StatusGone, "SESSION_EXPIRED", "Phiên vòng quay đã hết hạn."
	case errors.Is(err, repopg.ErrWheelSessionCompleted):
		status, code, message = http.StatusConflict, "SESSION_COMPLETED", "Bạn đã hoàn thành vòng quay."
	case errors.Is(err, repopg.ErrWheelRoundOrder):
		status, code, message = http.StatusConflict, "ROUND_INVALID_ORDER", "Vui lòng quay đúng thứ tự."
	case errors.As(err, &notReady):
		writeJSON(w, http.StatusConflict, map[string]any{"message": "Lượt tiếp theo chưa sẵn sàng.", "code": "ROUND_NOT_READY", "available_at": notReady.AvailableAt.UTC().Format(time.RFC3339Nano)})
		return
	case errors.Is(err, service.ErrWheelChatInvalid):
		status, code, message = http.StatusBadRequest, "CHAT_INVALID", "Tin nhắn không hợp lệ hoặc chứa liên kết."
	case errors.Is(err, service.ErrWheelChatRateLimit):
		status, code, message = http.StatusTooManyRequests, "CHAT_RATE_LIMITED", "Bạn gửi tin quá nhanh, vui lòng thử lại."
	case errors.Is(err, repopg.ErrWheelChatBanned):
		status, code, message = http.StatusForbidden, "CHAT_BANNED", "Bạn đang bị khóa chat trong sự kiện."
	case err != nil && err.Error() == "wheel.chat.duplicate":
		status, code, message = http.StatusBadRequest, "CHAT_DUPLICATE", "Không thể gửi hai tin giống nhau liên tiếp."
	}
	writeJSON(w, status, map[string]string{"message": message, "code": code})
}

var wheelWSUpgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 8192, CheckOrigin: func(*http.Request) bool { return true }}
