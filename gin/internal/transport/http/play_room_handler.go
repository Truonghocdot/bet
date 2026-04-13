package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/game"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/clock"
	"gin/internal/support/message"
	"github.com/gorilla/websocket"
)

type PlayRoomHandler struct {
	playRoomService *service.PlayRoomService
	broker          *realtime.Broker
}

func NewPlayRoomHandler(playRoomService *service.PlayRoomService, broker *realtime.Broker) *PlayRoomHandler {
	return &PlayRoomHandler{playRoomService: playRoomService, broker: broker}
}

func (h *PlayRoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	response, err := h.playRoomService.ListRooms(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *PlayRoomHandler) RoomState(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	response, err := h.playRoomService.GetRoomState(r.Context(), roomCode)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *PlayRoomHandler) RoomStateStream(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	if roomCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.GameRoomCodeRequired})
		return
	}

	initialResponse, err := h.playRoomService.GetRoomState(r.Context(), roomCode)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	stream, err := newSSEStream(w)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	if err := stream.Event("room.state", initialResponse); err != nil {
		return
	}

	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.PlayRoomTopic(roomCode))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}
	defer unsubscribe()

	clockTicker := time.NewTicker(time.Second)
	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer clockTicker.Stop()
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case update, ok := <-updates:
			if !ok {
				return
			}
			if update.Event != "room.state" {
				continue
			}
			if err := stream.EventRaw(update.Event, update.Data); err != nil {
				return
			}
		case <-clockTicker.C:
			if err := stream.Event("room.clock", map[string]any{
				"server_time": clock.Now(),
			}); err != nil {
				return
			}
		case <-heartbeatTicker.C:
			if err := stream.KeepAlive(); err != nil {
				return
			}
		}
	}
}

type wsRoomEventPayload struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

var playWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *PlayRoomHandler) RoomStateWS(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	if roomCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.GameRoomCodeRequired})
		return
	}

	initialResponse, err := h.playRoomService.GetRoomState(r.Context(), roomCode)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	conn, err := playWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(_ string) error {
		return conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, readErr := conn.ReadMessage(); readErr != nil {
				return
			}
		}
	}()

	if err := conn.WriteJSON(wsRoomEventPayload{Event: "room.state", Data: initialResponse}); err != nil {
		return
	}

	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.PlayRoomTopic(roomCode))
	if err != nil {
		_ = conn.WriteJSON(wsRoomEventPayload{Event: "error", Data: map[string]string{"message": message.InternalServerError}})
		return
	}
	defer unsubscribe()

	clockTicker := time.NewTicker(time.Second)
	pingTicker := time.NewTicker(20 * time.Second)
	defer clockTicker.Stop()
	defer pingTicker.Stop()

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
			if update.Event != "room.state" {
				continue
			}
			if err := conn.WriteJSON(wsRoomEventPayload{Event: update.Event, Data: json.RawMessage(update.Data)}); err != nil {
				return
			}
		case <-clockTicker.C:
			if err := conn.WriteJSON(wsRoomEventPayload{
				Event: "room.clock",
				Data: map[string]any{
					"server_time": clock.Now(),
				},
			}); err != nil {
				return
			}
		case <-pingTicker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second)); err != nil {
				return
			}
		}
	}
}

func (h *PlayRoomHandler) RoomHistory(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	page, pageSize := readPagination(r)

	response, err := h.playRoomService.ListRoomHistory(r.Context(), roomCode, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *PlayRoomHandler) MyRoomBets(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	page, pageSize := readPagination(r)

	response, err := h.playRoomService.ListMyRoomBets(r.Context(), claims.UserID, roomCode, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *PlayRoomHandler) PlaceRoomBet(w http.ResponseWriter, r *http.Request) {
	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	var req game.RoomBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidBetPayload})
		return
	}

	response, err := h.playRoomService.PlaceRoomBet(
		r.Context(),
		claims.UserID,
		roomCode,
		req,
		strings.TrimSpace(r.Header.Get("X-Forwarded-For")),
		strings.TrimSpace(r.UserAgent()),
		strings.TrimSpace(r.Header.Get("X-Connection-ID")),
	)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": playRoomErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusAccepted, response)
}

func playRoomErrorMessage(err error) string {
	switch {
	case errors.Is(err, repopg.ErrGameRoomNotFound):
		return message.GameRoomNotFound
	case errors.Is(err, repopg.ErrPeriodNotFound):
		return message.PeriodNotFound
	case errors.Is(err, repopg.ErrPeriodNotOpen):
		return message.PeriodNotOpen
	case errors.Is(err, repopg.ErrPeriodBetLocked):
		return message.PeriodBetLocked
	case errors.Is(err, repopg.ErrInsufficientBetBalance):
		return message.InsufficientBalanceBet
	case errors.Is(err, repopg.ErrInsufficientPlayBalance):
		return message.InsufficientBalancePlay
	default:
		return message.InternalServerError
	}
}
