package http

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/auth"
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

type wsRoomActionPayload struct {
	Action    string         `json:"action"`
	RequestID string         `json:"request_id"`
	PeriodID  string         `json:"period_id"`
	ConnectionID string      `json:"connection_id"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
	Items     []game.BetItem `json:"items"`
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

	var userClaims *auth.TokenClaims
	if claims, ok := authmiddleware.CurrentClaims(r.Context()); ok {
		userClaims = &claims
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

	// Fetch initial room state
	initialResponse, err := h.playRoomService.GetRoomState(r.Context(), roomCode)
	if err != nil {
		_ = conn.WriteJSON(wsRoomEventPayload{Event: "error", Data: map[string]string{"message": message.InternalServerError}})
		return
	}

	// Channel to receive actions from read loop
	actionChan := make(chan wsRoomActionPayload)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var action wsRoomActionPayload
			if readErr := conn.ReadJSON(&action); readErr != nil {
				return
			}
			// Send action to main loop for processing
			select {
			case actionChan <- action:
			case <-done:
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
		case action := <-actionChan:
			// Handle incoming actions from client
			switch action.Action {
			case "request_state":
				// Always fetch latest snapshot instead of reusing stale initialResponse.
				latestState, stateErr := h.playRoomService.GetRoomState(r.Context(), roomCode)
				if stateErr != nil {
					_ = conn.WriteJSON(wsRoomEventPayload{
						Event: "error",
						Data: map[string]any{
							"message": playRoomErrorMessage(stateErr),
						},
					})
					continue
				}
				if err := conn.WriteJSON(wsRoomEventPayload{Event: "room.state", Data: latestState}); err != nil {
					return
				}
			case "request_history":
				page := action.Page
				if page <= 0 {
					page = 1
				}
				pageSize := action.PageSize
				if pageSize <= 0 {
					pageSize = 4
				}

				response, historyErr := h.playRoomService.ListRoomHistory(r.Context(), roomCode, page, pageSize)
				if historyErr != nil {
					_ = conn.WriteJSON(wsRoomEventPayload{
						Event: "history.error",
						Data: map[string]any{
							"request_id": action.RequestID,
							"message":    playRoomErrorMessage(historyErr),
						},
					})
					continue
				}

				_ = conn.WriteJSON(wsRoomEventPayload{
					Event: "history.page",
					Data: map[string]any{
						"request_id": action.RequestID,
						"message":    response.Message,
						"page":       response.Page,
						"page_size":  response.PageSize,
						"total":      response.Total,
						"total_pages": response.TotalPages,
						"items":      response.Items,
					},
				})
			case "place_bet":
				// Handle bet placement
				if userClaims == nil {
					_ = conn.WriteJSON(wsRoomEventPayload{
						Event: "bet.error",
						Data: map[string]any{
							"request_id": action.RequestID,
							"message":    "Unauthorized: please login first",
						},
					})
					continue
				}

				h.handlePlaceBetAction(conn, r.Context(), userClaims, roomCode, action)
			}
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

func (h *PlayRoomHandler) MyBetsWS(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	if roomCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Room code required"})
		return
	}

	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Unauthorized"})
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

	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.PlayRoomBetsTopic(roomCode, claims.UserID))
	if err != nil {
		_ = conn.WriteJSON(wsRoomEventPayload{Event: "error", Data: map[string]string{"message": "Failed to subscribe"}})
		return
	}
	defer unsubscribe()

	clockTicker := time.NewTicker(time.Second)
	pingTicker := time.NewTicker(20 * time.Second)
	defer clockTicker.Stop()
	defer pingTicker.Stop()

	// Send initial state (list of latest bets)
	initialBets, err := h.playRoomService.ListMyRoomBets(r.Context(), claims.UserID, roomCode, 1, 4)
	if err == nil {
		_ = conn.WriteJSON(wsRoomEventPayload{Event: "bets.init", Data: initialBets})
	}

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
			if update.Event != "bets.updated" {
				continue
			}
			if err := conn.WriteJSON(wsRoomEventPayload{Event: update.Event, Data: json.RawMessage(update.Data)}); err != nil {
				return
			}
		case <-clockTicker.C:
			if err := conn.WriteJSON(wsRoomEventPayload{
				Event: "ping",
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

func (h *PlayRoomHandler) handlePlaceBetAction(
	conn *websocket.Conn,
	ctx context.Context,
	userClaims *auth.TokenClaims,
	roomCode string,
	action wsRoomActionPayload,
) {
	// Validate required fields
	if action.RequestID == "" {
		log.Printf("[play.bet.ws.reject] room=%s user_id=%d reason=request_id_required", roomCode, userClaims.UserID)
		_ = conn.WriteJSON(wsRoomEventPayload{
			Event: "bet.error",
			Data: map[string]any{
				"request_id": action.RequestID,
				"message":    "Invalid request: request_id is required",
			},
		})
		return
	}

	if action.PeriodID == "" {
		log.Printf("[play.bet.ws.reject] room=%s user_id=%d request_id=%s reason=period_id_required", roomCode, userClaims.UserID, action.RequestID)
		_ = conn.WriteJSON(wsRoomEventPayload{
			Event: "bet.error",
			Data: map[string]any{
				"request_id": action.RequestID,
				"message":    "Invalid request: period_id is required",
			},
		})
		return
	}

	if len(action.Items) == 0 {
		log.Printf("[play.bet.ws.reject] room=%s user_id=%d request_id=%s reason=items_empty", roomCode, userClaims.UserID, action.RequestID)
		_ = conn.WriteJSON(wsRoomEventPayload{
			Event: "bet.error",
			Data: map[string]any{
				"request_id": action.RequestID,
				"message":    "Invalid request: items cannot be empty",
			},
		})
		return
	}

	connectionID := strings.TrimSpace(action.ConnectionID)
	if connectionID == "" {
		log.Printf("[play.bet.ws.reject] room=%s user_id=%d request_id=%s reason=connection_id_required", roomCode, userClaims.UserID, action.RequestID)
		_ = conn.WriteJSON(wsRoomEventPayload{
			Event: "bet.error",
			Data: map[string]any{
				"request_id": action.RequestID,
				"message":    message.MissingConnectionID,
			},
		})
		return
	}

	// Create RoomBetRequest
	req := game.RoomBetRequest{
		RequestID: action.RequestID,
		PeriodID:  action.PeriodID,
		Items:     action.Items,
	}
	log.Printf(
		"[play.bet.ws.receive] room=%s user_id=%d request_id=%s period_id=%s connection_id=%s items=%d",
		roomCode,
		userClaims.UserID,
		action.RequestID,
		action.PeriodID,
		connectionID,
		len(action.Items),
	)

	// Call service to place bet
	response, err := h.playRoomService.PlaceRoomBet(
		ctx,
		userClaims.UserID,
		roomCode,
		req,
		"",            // X-Forwarded-For (not available from WebSocket)
		"WebSocket",   // User-Agent
		connectionID,
	)

	if err != nil {
		log.Printf(
			"[play.bet.ws.error] room=%s user_id=%d request_id=%s period_id=%s connection_id=%s err=%v",
			roomCode,
			userClaims.UserID,
			action.RequestID,
			action.PeriodID,
			connectionID,
			err,
		)
		_ = conn.WriteJSON(wsRoomEventPayload{
			Event: "bet.error",
			Data: map[string]any{
				"request_id": action.RequestID,
				"message":    playRoomErrorMessage(err),
			},
		})
		return
	}

	// Send success response
	log.Printf(
		"[play.bet.ws.ok] room=%s user_id=%d request_id=%s period_id=%s connection_id=%s status=%s",
		roomCode,
		userClaims.UserID,
		action.RequestID,
		action.PeriodID,
		connectionID,
		response.Status,
	)
	_ = conn.WriteJSON(wsRoomEventPayload{
		Event: "bet.placed",
		Data: map[string]any{
			"request_id":  action.RequestID,
			"message":     response.Message,
			"status":      response.Status,
			"accepted_at": response.AcceptedAt,
		},
	})
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
