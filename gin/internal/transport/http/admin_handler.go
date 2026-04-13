package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/clock"
	"gin/internal/support/message"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type AdminHandler struct {
	gameRepository *repopg.GameRepository
	broker         *realtime.Broker
	redis          *redis.Client
	authService    *service.AuthService
}

func NewAdminHandler(gameRepo *repopg.GameRepository, broker *realtime.Broker, redisClient *redis.Client, authService *service.AuthService) *AdminHandler {
	return &AdminHandler{gameRepository: gameRepo, broker: broker, redis: redisClient, authService: authService}
}

type adminPeriodDetail struct {
	ID           int64   `json:"id"`
	PeriodNo     string  `json:"period_no"`
	DrawAt       string  `json:"draw_at"`
	BetLockAt    string  `json:"bet_lock_at"`
	Status       int     `json:"status"`
	ManualResult *string `json:"manual_result"`
}

type adminRoomDetail struct {
	Code     string                  `json:"code"`
	GameType int                     `json:"game_type"`
	Period   *adminPeriodDetail      `json:"period"`
	Stats    []repopg.PeriodBetStats `json:"bet_stats"`
}

type adminStatsResponse struct {
	ServerTime time.Time         `json:"server_time"`
	Rooms      []adminRoomDetail `json:"rooms"`
}

type wsEventPayload struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type cachedRoomState struct {
	Payload game.RoomStateResponse `json:"payload"`
}

func toAdminPeriod(record *repopg.GamePeriodRecord) *adminPeriodDetail {
	if record == nil {
		return nil
	}
	var manual *string
	raw := strings.TrimSpace(string(record.ManualResultJSON))
	if raw != "" && raw != "null" {
		manual = &raw
	}
	return &adminPeriodDetail{
		ID:           record.ID,
		PeriodNo:     record.PeriodNo,
		DrawAt:       record.DrawAt.Format(time.RFC3339),
		BetLockAt:    record.BetLockAt.Format(time.RFC3339),
		Status:       record.Status,
		ManualResult: manual,
	}
}

func toPeriodStatusCode(status string) int {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "SCHEDULED":
		return 1
	case "OPEN":
		return 2
	case "LOCKED":
		return 3
	case "DRAWN":
		return 4
	case "SETTLED":
		return 5
	case "CANCELED":
		return 6
	default:
		return 0
	}
}

func (h *AdminHandler) roomStateCacheKey(roomCode string) string {
	return "cache:play:room_state:" + strings.TrimSpace(roomCode)
}

func (h *AdminHandler) loadPeriodFromRoomStateCache(r *http.Request, roomCode string) *adminPeriodDetail {
	if h.redis == nil {
		return nil
	}
	raw, err := h.redis.Get(r.Context(), h.roomStateCacheKey(roomCode)).Bytes()
	if err != nil || len(raw) == 0 {
		return nil
	}
	var cached cachedRoomState
	if err := json.Unmarshal(raw, &cached); err != nil {
		return nil
	}
	if cached.Payload.CurrentPeriod.ID == 0 {
		return nil
	}
	return &adminPeriodDetail{
		ID:           cached.Payload.CurrentPeriod.ID,
		PeriodNo:     cached.Payload.CurrentPeriod.PeriodNo,
		DrawAt:       cached.Payload.CurrentPeriod.DrawAt.Format(time.RFC3339),
		BetLockAt:    cached.Payload.CurrentPeriod.BetLockAt.Format(time.RFC3339),
		Status:       toPeriodStatusCode(cached.Payload.CurrentPeriod.Status),
		ManualResult: nil,
	}
}

func (h *AdminHandler) buildStatsSnapshot(r *http.Request) (adminStatsResponse, error) {
	ctx := r.Context()
	stats, err := h.gameRepository.ListAllRoomsWithCurrentPeriod(ctx)
	if err != nil {
		return adminStatsResponse{}, err
	}

	rooms := make([]adminRoomDetail, 0, len(stats))
	for _, s := range stats {
		period := toAdminPeriod(s.Period)
		if period == nil {
			period = h.loadPeriodFromRoomStateCache(r, s.Room.Code)
		}

		var betStats []repopg.PeriodBetStats
		if period != nil {
			betStats, _ = h.gameRepository.GetPeriodBetStats(ctx, period.ID)
		}
		rooms = append(rooms, adminRoomDetail{
			Code:     s.Room.Code,
			GameType: s.Room.GameType,
			Period:   period,
			Stats:    betStats,
		})
	}

	return adminStatsResponse{
		ServerTime: clock.Now(),
		Rooms:      rooms,
	}, nil
}

func (h *AdminHandler) ListRoomStats(w http.ResponseWriter, r *http.Request) {
	snapshot, err := h.buildStatsSnapshot(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi khi tải danh sách phòng"})
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (h *AdminHandler) StreamRoomStats(w http.ResponseWriter, r *http.Request) {
	stream, err := newSSEStream(w)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	initial, err := h.buildStatsSnapshot(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi khi tải danh sách phòng"})
		return
	}
	if err := stream.Event("admin.rooms.stats", initial); err != nil {
		return
	}

	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.AdminRoomsTopic())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}
	defer unsubscribe()

	heartbeatTicker := time.NewTicker(20 * time.Second)
	fullRefreshTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()
	defer fullRefreshTicker.Stop()

	emitSnapshot := func() bool {
		snapshot, buildErr := h.buildStatsSnapshot(r)
		if buildErr != nil {
			return false
		}
		if writeErr := stream.Event("admin.rooms.stats", snapshot); writeErr != nil {
			return false
		}
		return true
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case _, ok := <-updates:
			if !ok {
				return
			}
			if !emitSnapshot() {
				return
			}
		case <-fullRefreshTicker.C:
			if !emitSnapshot() {
				return
			}
		case <-heartbeatTicker.C:
			if err := stream.KeepAlive(); err != nil {
				return
			}
		}
	}
}

var adminWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *AdminHandler) StreamRoomStatsWS(w http.ResponseWriter, r *http.Request) {
	if h.authService == nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": message.InternalServerError})
		return
	}

	tokenValue := strings.TrimSpace(r.URL.Query().Get("token"))
	if tokenValue == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.MissingBearerToken})
		return
	}

	claims, err := h.authService.VerifyAccessToken(tokenValue)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": message.InvalidAccessToken})
		return
	}
	if claims.Role != 1 {
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "Quyền truy cập bị từ chối"})
		return
	}

	conn, err := adminWSUpgrader.Upgrade(w, r, nil)
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

	initial, err := h.buildStatsSnapshot(r)
	if err != nil {
		_ = conn.WriteJSON(wsEventPayload{Event: "error", Data: map[string]string{"message": "Lỗi khi tải danh sách phòng"}})
		return
	}
	if err := conn.WriteJSON(wsEventPayload{Event: "admin.rooms.stats", Data: initial}); err != nil {
		return
	}

	updates, unsubscribe, err := h.broker.Subscribe(r.Context(), realtime.AdminRoomsTopic())
	if err != nil {
		_ = conn.WriteJSON(wsEventPayload{Event: "error", Data: map[string]string{"message": message.InternalServerError}})
		return
	}
	defer unsubscribe()

	pingTicker := time.NewTicker(20 * time.Second)
	fullRefreshTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()
	defer fullRefreshTicker.Stop()

	emitSnapshot := func() bool {
		snapshot, buildErr := h.buildStatsSnapshot(r)
		if buildErr != nil {
			log.Printf("[admin.ws] build snapshot error: %v", buildErr)
			return false
		}
		writeErr := conn.WriteJSON(wsEventPayload{Event: "admin.rooms.stats", Data: snapshot})
		return writeErr == nil
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-done:
			return
		case <-pingTicker.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(10*time.Second)); err != nil {
				return
			}
		case <-fullRefreshTicker.C:
			if !emitSnapshot() {
				return
			}
		case _, ok := <-updates:
			if !ok {
				return
			}
			if !emitSnapshot() {
				return
			}
		}
	}
}

func (h *AdminHandler) SetManualResult(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	periodID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.PeriodNotFound})
		return
	}

	var request struct {
		Result      string `json:"result"`
		BigSmall    string `json:"big_small"`
		Color       string `json:"color"`
		PayloadJSON any    `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Dữ liệu không hợp lệ"})
		return
	}

	payload, _ := json.Marshal(request.PayloadJSON)
	draw := repopg.DrawResult{
		Result:      request.Result,
		BigSmall:    request.BigSmall,
		Color:       request.Color,
		PayloadJSON: payload,
	}

	drawJSON, _ := json.Marshal(draw)
	if err := h.gameRepository.SetPeriodManualResult(r.Context(), periodID, drawJSON); err != nil {
		if errors.Is(err, repopg.ErrPeriodBetLocked) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": message.PeriodBetLocked})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi khi lưu kết quả can thiệp"})
		return
	}
	_ = h.broker.Publish(r.Context(), realtime.AdminRoomsTopic(), "admin.rooms.changed", map[string]any{
		"period_id": periodID,
		"at":        clock.Now(),
		"source":    "manual_result.updated",
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "Đã lưu kết quả dự kiến"})
}
