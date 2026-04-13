package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/game"
	repopg "gin/internal/repository/postgres"
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
		switch r.Method {
		case http.MethodPost:
			h.handlePlaceBet(w, r, gameType)
		case http.MethodGet:
			h.handleMyBets(w, r, gameType)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": message.RouteNotFound})
		}
	case "history":
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": message.RouteNotFound})
			return
		}
		h.handleHistory(w, r, gameType)
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
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": gameErrorMessage(err)})
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
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": gameErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusAccepted, response)
}

func (h *GameHandler) handleHistory(w http.ResponseWriter, r *http.Request, gameType game.GameType) {
	page, pageSize := readPagination(r)

	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	response, err := h.betService.ListGameHistory(r.Context(), gameType, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": gameErrorMessage(err)})
		return
	}

	_ = claims
	writeJSON(w, http.StatusOK, response)
}

func (h *GameHandler) handleMyBets(w http.ResponseWriter, r *http.Request, gameType game.GameType) {
	page, pageSize := readPagination(r)

	claims, ok := authmiddleware.CurrentClaims(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.Unauthorized})
		return
	}

	response, err := h.betService.ListMyBets(r.Context(), claims.UserID, gameType, page, pageSize)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": gameErrorMessage(err)})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func gameErrorMessage(err error) string {
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

func readPagination(r *http.Request) (int, int) {
	page := 1
	pageSize := 5

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
