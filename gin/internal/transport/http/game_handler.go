package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/game"
	"gin/internal/service"
	"gin/internal/support/message"
)

type GameHandler struct {
	sessionService *service.GameSessionService
	betService     *service.BetService
}

func NewGameHandler(sessionService *service.GameSessionService, betService *service.BetService) *GameHandler {
	return &GameHandler{
		sessionService: sessionService,
		betService:     betService,
	}
}

func (h *GameHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/games/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) != 2 {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
		return
	}

	gameType := game.GameType(parts[0])
	action := parts[1]

	switch action {
	case "join":
		h.handleJoin(w, r, gameType)
	case "bets":
		h.handlePlaceBet(w, r, gameType)
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.RouteNotFound})
	}
}

func (h *GameHandler) handleJoin(w http.ResponseWriter, r *http.Request, gameType game.GameType) {
	var request game.JoinRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&request)
	}

	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}
	request.UserID = strconv.FormatInt(claims.UserID, 10)

	response, err := h.sessionService.JoinGame(r.Context(), gameType, request.UserID)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *GameHandler) handlePlaceBet(w http.ResponseWriter, r *http.Request, gameType game.GameType) {
	connectionID := r.Header.Get("X-Connection-ID")
	if connectionID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.MissingConnectionID})
		return
	}

	var request game.PlaceBetRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.InvalidBetPayload})
		return
	}

	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	request.GameType = gameType
	request.UserID = strconv.FormatInt(claims.UserID, 10)

	response, err := h.betService.PlaceBet(r.Context(), connectionID, request)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, response)
}
