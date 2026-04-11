package http

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/game"
	"gin/internal/service"
	"gin/internal/support/clock"
	"gin/internal/support/message"
)

type PlayRoomHandler struct {
	playRoomService *service.PlayRoomService
}

func NewPlayRoomHandler(playRoomService *service.PlayRoomService) *PlayRoomHandler {
	return &PlayRoomHandler{playRoomService: playRoomService}
}

func (h *PlayRoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	response, err := h.playRoomService.ListRooms(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *PlayRoomHandler) RoomState(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	response, err := h.playRoomService.GetRoomState(r.Context(), roomCode)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
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
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	stream, err := newSSEStream(w)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	stateTicker := time.NewTicker(2 * time.Second)
	clockTicker := time.NewTicker(time.Second)
	heartbeatTicker := time.NewTicker(20 * time.Second)
	defer stateTicker.Stop()
	defer clockTicker.Stop()
	defer heartbeatTicker.Stop()

	lastPayload := ""
	emitState := func(response any) error {
		payload, err := json.Marshal(response)
		if err != nil {
			return err
		}
		payloadKey := string(payload)
		if payloadKey == lastPayload {
			return nil
		}

		lastPayload = payloadKey
		return stream.Event("room.state", response)
	}

	if err := emitState(initialResponse); err != nil {
		return
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-stateTicker.C:
			response, err := h.playRoomService.GetRoomState(r.Context(), roomCode)
			if err != nil {
				return
			}
			if err := emitState(response); err != nil {
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

func (h *PlayRoomHandler) RoomHistory(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	page, pageSize := readPagination(r)

	response, err := h.playRoomService.ListRoomHistory(r.Context(), roomCode, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
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
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
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
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, response)
}
