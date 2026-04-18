package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	authmiddleware "gin/internal/auth/middleware"
	"gin/internal/domain/game"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/clock"
	"gin/internal/support/id"
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
	PeriodIndex  int64   `json:"period_index"`
	DrawAt       string  `json:"draw_at"`
	BetLockAt    string  `json:"bet_lock_at"`
	Status       int     `json:"status"`
	ManualResult *string `json:"manual_result"`
}

type adminRoomPresence struct {
	Count         int     `json:"count"`
	ActiveUserIDs []int64 `json:"active_user_ids"`
}

type adminControlRoomLock struct {
	UserID       int64     `json:"user_id"`
	SessionToken string    `json:"session_token"`
	LockedAt     time.Time `json:"locked_at"`
}

type adminRoomDetail struct {
	Code     string                  `json:"code"`
	GameType int                     `json:"game_type"`
	Period   *adminPeriodDetail      `json:"period"`
	Stats    []repopg.PeriodBetStats `json:"bet_stats"`
	Presence adminRoomPresence       `json:"presence"`
}

type adminStatsResponse struct {
	ServerTime time.Time         `json:"server_time"`
	Rooms      []adminRoomDetail `json:"rooms"`
}

type wsEventPayload struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

type adminManualResultRequest struct {
	Result      string `json:"result"`
	BigSmall    string `json:"big_small"`
	Color       string `json:"color"`
	PayloadJSON any    `json:"payload"`
}

type cachedRoomState struct {
	Payload game.RoomStateResponse `json:"payload"`
}

var errControlRoomSessionMismatch = errors.New("control room session mismatch")

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
		PeriodIndex:  record.PeriodIndex,
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

func parseManualK3Die(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		if typed < 1 || typed > 6 {
			return 0, false
		}
		return typed, true
	case int32:
		return parseManualK3Die(int(typed))
	case int64:
		return parseManualK3Die(int(typed))
	case float32:
		return parseManualK3Die(float64(typed))
	case float64:
		if typed != float64(int(typed)) {
			return 0, false
		}
		return parseManualK3Die(int(typed))
	case json.Number:
		intValue, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return parseManualK3Die(int(intValue))
	case string:
		intValue, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return parseManualK3Die(intValue)
	default:
		return 0, false
	}
}

func parseManualK3Dice(value any) ([]int, bool) {
	switch typed := value.(type) {
	case []int:
		outcome, ok := game.BuildK3Outcome(typed)
		if !ok {
			return nil, false
		}
		return outcome.Dice, true
	case []any:
		if len(typed) != 3 {
			return nil, false
		}
		dice := make([]int, 0, 3)
		for _, item := range typed {
			die, ok := parseManualK3Die(item)
			if !ok {
				return nil, false
			}
			dice = append(dice, die)
		}
		return dice, true
	default:
		return nil, false
	}
}

func parseManualLotteryDigit(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		if typed < 0 || typed > 9 {
			return 0, false
		}
		return typed, true
	case int32:
		return parseManualLotteryDigit(int(typed))
	case int64:
		return parseManualLotteryDigit(int(typed))
	case float32:
		return parseManualLotteryDigit(float64(typed))
	case float64:
		if typed != float64(int(typed)) {
			return 0, false
		}
		return parseManualLotteryDigit(int(typed))
	case json.Number:
		intValue, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return parseManualLotteryDigit(int(intValue))
	case string:
		intValue, err := strconv.Atoi(strings.TrimSpace(typed))
		if err != nil {
			return 0, false
		}
		return parseManualLotteryDigit(intValue)
	default:
		return 0, false
	}
}

func parseManualLotteryDigits(value any) ([]int, bool) {
	switch typed := value.(type) {
	case []int:
		outcome, ok := game.BuildLotteryOutcome(typed)
		if !ok {
			return nil, false
		}
		return outcome.Digits, true
	case []any:
		if len(typed) != 5 {
			return nil, false
		}
		digits := make([]int, 0, 5)
		for _, item := range typed {
			digit, ok := parseManualLotteryDigit(item)
			if !ok {
				return nil, false
			}
			digits = append(digits, digit)
		}
		return digits, true
	case string:
		raw := strings.TrimSpace(typed)
		if len(raw) != 5 {
			return nil, false
		}
		digits := make([]int, 0, 5)
		for _, ch := range raw {
			if ch < '0' || ch > '9' {
				return nil, false
			}
			digits = append(digits, int(ch-'0'))
		}
		return digits, true
	default:
		return nil, false
	}
}

func normalizeManualResultRequest(request *adminManualResultRequest) any {
	payloadMap, ok := request.PayloadJSON.(map[string]any)
	if !ok {
		return request.PayloadJSON
	}

	gameType := strings.ToLower(strings.TrimSpace(fmt.Sprint(payloadMap["game_type"])))
	switch gameType {
	case string(game.GameK3):
		dice, ok := parseManualK3Dice(payloadMap["dice"])
		if !ok {
			return request.PayloadJSON
		}

		outcome, ok := game.BuildK3Outcome(dice)
		if !ok {
			return request.PayloadJSON
		}

		payloadMap["game_type"] = string(game.GameK3)
		payloadMap["dice"] = outcome.Dice
		payloadMap["sum"] = outcome.Sum
		payloadMap["result"] = outcome.Result
		payloadMap["big_small"] = outcome.BigSmall
		payloadMap["odd_even"] = outcome.OddEven
		payloadMap["is_triple"] = outcome.IsTriple
		payloadMap["tags"] = outcome.Tags
		if _, exists := payloadMap["generated_at"]; !exists {
			payloadMap["generated_at"] = clock.Now()
		}

		request.Result = outcome.Result
		request.BigSmall = outcome.BigSmall
		request.Color = "-"
		return payloadMap
	case string(game.GameLottery):
		digits, ok := parseManualLotteryDigits(payloadMap["digits"])
		if !ok {
			digits, ok = parseManualLotteryDigits(request.Result)
			if !ok {
				return request.PayloadJSON
			}
		}

		outcome, ok := game.BuildLotteryOutcome(digits)
		if !ok {
			return request.PayloadJSON
		}

		payloadMap["game_type"] = string(game.GameLottery)
		payloadMap["digits"] = outcome.Digits
		payloadMap["positions"] = outcome.Positions
		payloadMap["sum"] = outcome.Sum
		payloadMap["sum_big_small"] = outcome.SumBigSmall
		payloadMap["sum_odd_even"] = outcome.SumOddEven
		payloadMap["last_digit"] = outcome.LastDigit
		payloadMap["result"] = outcome.Result
		payloadMap["big_small"] = outcome.BigSmall
		payloadMap["odd_even"] = outcome.OddEven
		payloadMap["tags"] = outcome.Tags
		if _, exists := payloadMap["generated_at"]; !exists {
			payloadMap["generated_at"] = clock.Now()
		}

		request.Result = outcome.Result
		request.BigSmall = outcome.BigSmall
		request.Color = "-"
		return payloadMap
	default:
		return request.PayloadJSON
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
		PeriodIndex:  cached.Payload.CurrentPeriod.PeriodIndex,
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

	requestedRoomCode := strings.TrimSpace(r.URL.Query().Get("room_code"))
	if requestedRoomCode != "" {
		filtered := make([]struct {
			Room   repopg.GameRoomRecord
			Period *repopg.GamePeriodRecord
		}, 0, 1)
		for _, item := range stats {
			if strings.EqualFold(strings.TrimSpace(item.Room.Code), requestedRoomCode) {
				filtered = append(filtered, item)
			}
		}
		stats = filtered
	}

	roomCodes := make([]string, 0, len(stats))
	for _, item := range stats {
		roomCodes = append(roomCodes, item.Room.Code)
	}
	presenceMap := h.listControlRoomPresence(ctx, roomCodes)

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
			Presence: presenceMap[strings.TrimSpace(s.Room.Code)],
		})
	}

	return adminStatsResponse{
		ServerTime: clock.Now(),
		Rooms:      rooms,
	}, nil
}

const (
	adminControlLockKey      = "admin:lock:control_center"
	adminControlPresenceTTL  = 45 * time.Second
	adminControlPresenceBase = "admin:lock:control_room:"
)

func adminControlRoomLockKey(roomCode string) string {
	return fmt.Sprintf("%s%s", adminControlPresenceBase, strings.TrimSpace(roomCode))
}

func (h *AdminHandler) loadControlRoomLock(ctx context.Context, roomCode string) (*adminControlRoomLock, error) {
	if h.redis == nil {
		return nil, redis.Nil
	}

	raw, err := h.redis.Get(ctx, adminControlRoomLockKey(roomCode)).Result()
	if err != nil {
		return nil, err
	}

	var lock adminControlRoomLock
	if err := json.Unmarshal([]byte(raw), &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

func (h *AdminHandler) touchControlRoomLock(ctx context.Context, roomCode string, userID int64, sessionToken string) error {
	if h.redis == nil {
		return nil
	}

	currentLock, err := h.loadControlRoomLock(ctx, roomCode)
	if err != nil {
		return err
	}
	if currentLock == nil {
		return redis.Nil
	}
	if currentLock.UserID != userID || strings.TrimSpace(currentLock.SessionToken) != strings.TrimSpace(sessionToken) {
		return errControlRoomSessionMismatch
	}

	return h.redis.Expire(ctx, adminControlRoomLockKey(roomCode), adminControlPresenceTTL).Err()
}

func (h *AdminHandler) listControlRoomPresence(ctx context.Context, roomCodes []string) map[string]adminRoomPresence {
	result := make(map[string]adminRoomPresence, len(roomCodes))
	for _, roomCode := range roomCodes {
		normalized := strings.TrimSpace(roomCode)
		result[normalized] = adminRoomPresence{Count: 0, ActiveUserIDs: []int64{}}
	}

	if h.redis == nil {
		return result
	}

	for _, roomCode := range roomCodes {
		normalized := strings.TrimSpace(roomCode)
		lock, err := h.loadControlRoomLock(ctx, normalized)
		if err != nil {
			if !errors.Is(err, redis.Nil) {
				continue
			}
			result[normalized] = adminRoomPresence{Count: 0, ActiveUserIDs: []int64{}}
			continue
		}

		result[normalized] = adminRoomPresence{
			Count:         1,
			ActiveUserIDs: []int64{lock.UserID},
		}
	}

	return result
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
	if claims.Role != 0 && claims.Role != 1 {
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "Quyền truy cập bị từ chối"})
		return
	}

	wsRoomCode := strings.TrimSpace(r.URL.Query().Get("room_code"))
	wsControlSession := strings.TrimSpace(r.URL.Query().Get("control_session"))

	conn, err := adminWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(75 * time.Second))
	conn.SetPongHandler(func(_ string) error {
		if err := conn.SetReadDeadline(time.Now().Add(75 * time.Second)); err != nil {
			return err
		}
		if wsRoomCode != "" && wsControlSession != "" {
			if err := h.touchControlRoomLock(r.Context(), wsRoomCode, claims.UserID, wsControlSession); err != nil && !errors.Is(err, redis.Nil) && !errors.Is(err, errControlRoomSessionMismatch) {
				log.Printf("[admin.ws][control_room.touch.error] room=%s user_id=%d err=%v", wsRoomCode, claims.UserID, err)
			}
		}
		return nil
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

	if wsRoomCode != "" && wsControlSession != "" {
		if err := h.touchControlRoomLock(r.Context(), wsRoomCode, claims.UserID, wsControlSession); err != nil && !errors.Is(err, redis.Nil) && !errors.Is(err, errControlRoomSessionMismatch) {
			log.Printf("[admin.ws][control_room.touch.initial.error] room=%s user_id=%d err=%v", wsRoomCode, claims.UserID, err)
		}
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

	var request adminManualResultRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Dữ liệu không hợp lệ"})
		return
	}

	payloadJSON := normalizeManualResultRequest(&request)
	payload, _ := json.Marshal(payloadJSON)
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

func (h *AdminHandler) EnterControlRoom(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmiddleware.CurrentClaims(r.Context())
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	if roomCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.GameRoomCodeRequired})
		return
	}

	if _, err := h.gameRepository.FindRoomByCode(r.Context(), roomCode); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"message": message.GameRoomNotFound})
		return
	}

	sessionToken := id.Long()
	lockPayload := adminControlRoomLock{
		UserID:       claims.UserID,
		SessionToken: sessionToken,
		LockedAt:     clock.Now(),
	}
	lockJSON, _ := json.Marshal(lockPayload)

	if h.redis != nil {
		success, err := h.redis.SetNX(r.Context(), adminControlRoomLockKey(roomCode), lockJSON, adminControlPresenceTTL).Result()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Không thể ghi nhận slot control room"})
			return
		}

		if !success {
			currentLock, loadErr := h.loadControlRoomLock(r.Context(), roomCode)
			if loadErr == nil && currentLock != nil && currentLock.UserID == claims.UserID {
				if err := h.redis.Set(r.Context(), adminControlRoomLockKey(roomCode), lockJSON, adminControlPresenceTTL).Err(); err != nil {
					writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Không thể làm mới session điều khiển phòng"})
					return
				}
			} else {
				holder := "Admin khác"
				if currentLock != nil && currentLock.UserID > 0 {
					holder = "Admin #" + strconv.FormatInt(currentLock.UserID, 10)
				}
				writeJSON(w, http.StatusConflict, map[string]string{
					"message": "Phòng này đang được điều khiển bởi admin khác",
					"holder":  holder,
				})
				return
			}
		}
	}

	_ = h.broker.Publish(r.Context(), realtime.AdminRoomsTopic(), "admin.rooms.changed", map[string]any{
		"room_code": roomCode,
		"user_id":   claims.UserID,
		"source":    "control_room.enter",
		"at":        clock.Now(),
	})

	presence := h.listControlRoomPresence(r.Context(), []string{roomCode})[roomCode]
	writeJSON(w, http.StatusOK, map[string]any{
		"message":       "Đã vào control room",
		"room_code":     roomCode,
		"presence":      presence,
		"session_token": sessionToken,
	})
}

func (h *AdminHandler) HeartbeatControlRoom(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmiddleware.CurrentClaims(r.Context())
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	sessionToken := strings.TrimSpace(r.Header.Get("X-Control-Session"))
	if roomCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.GameRoomCodeRequired})
		return
	}
	if sessionToken == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Thiếu session điều khiển phòng"})
		return
	}

	if h.redis != nil {
		currentLock, err := h.loadControlRoomLock(r.Context(), roomCode)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				writeJSON(w, http.StatusGone, map[string]string{"message": "Session điều khiển phòng đã hết hạn"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Không thể duy trì slot control room"})
			return
		}
		if currentLock.UserID != claims.UserID || currentLock.SessionToken != sessionToken {
			writeJSON(w, http.StatusForbidden, map[string]string{"message": "Session điều khiển phòng không còn hợp lệ"})
			return
		}
		if err := h.redis.Expire(r.Context(), adminControlRoomLockKey(roomCode), adminControlPresenceTTL).Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Không thể duy trì slot control room"})
			return
		}
	}

	_ = h.broker.Publish(r.Context(), realtime.AdminRoomsTopic(), "admin.rooms.changed", map[string]any{
		"room_code": roomCode,
		"user_id":   claims.UserID,
		"source":    "control_room.heartbeat",
		"at":        clock.Now(),
	})

	presence := h.listControlRoomPresence(r.Context(), []string{roomCode})[roomCode]
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "Duy trì slot control room thành công",
		"room_code": roomCode,
		"presence":  presence,
	})
}

func (h *AdminHandler) LeaveControlRoom(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmiddleware.CurrentClaims(r.Context())
	roomCode := strings.TrimSpace(r.PathValue("room_code"))
	sessionToken := strings.TrimSpace(r.Header.Get("X-Control-Session"))
	if roomCode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": message.GameRoomCodeRequired})
		return
	}

	if h.redis != nil {
		currentLock, err := h.loadControlRoomLock(r.Context(), roomCode)
		if err == nil && currentLock != nil && currentLock.UserID == claims.UserID && currentLock.SessionToken == sessionToken {
			_ = h.redis.Del(r.Context(), adminControlRoomLockKey(roomCode)).Err()
		}
	}

	_ = h.broker.Publish(r.Context(), realtime.AdminRoomsTopic(), "admin.rooms.changed", map[string]any{
		"room_code": roomCode,
		"user_id":   claims.UserID,
		"source":    "control_room.leave",
		"at":        clock.Now(),
	})

	presence := h.listControlRoomPresence(r.Context(), []string{roomCode})[roomCode]
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "Đã rời control room",
		"room_code": roomCode,
		"presence":  presence,
	})
}

func (h *AdminHandler) AcquireLock(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmiddleware.CurrentClaims(r.Context())
	userID := claims.UserID

	ctx := r.Context()
	lockValue := strconv.FormatInt(userID, 10)

	// NX: Set if not exists, GET: Return old value if exists
	success, err := h.redis.SetNX(ctx, adminControlLockKey, lockValue, 45*time.Second).Result()
	if err != nil {
		log.Printf("[admin][lock.acquire.error] user_id=%d err=%v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi hệ thống khi kiểm tra khóa"})
		return
	}

	if !success {
		current, _ := h.redis.Get(ctx, adminControlLockKey).Result()
		if current != lockValue {
			log.Printf("[admin][lock.acquire.denied] user_id=%d holder=%s", userID, current)
			writeJSON(w, http.StatusConflict, map[string]string{
				"message": "Trang này đang được quản lý bởi một Admin khác",
				"holder":  "Admin #" + current,
			})
			return
		}
	}

	log.Printf("[admin][lock.acquire.ok] user_id=%d ttl_seconds=45", userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Đã giữ quyền kiểm soát"})
}

func (h *AdminHandler) HeartbeatLock(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmiddleware.CurrentClaims(r.Context())
	userID := claims.UserID
	lockValue := strconv.FormatInt(userID, 10)

	ctx := r.Context()
	current, err := h.redis.Get(ctx, adminControlLockKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			log.Printf("[admin][lock.heartbeat.gone] user_id=%d", userID)
			writeJSON(w, http.StatusGone, map[string]string{"message": "Mất phiên quản lý"})
			return
		}
		log.Printf("[admin][lock.heartbeat.error] user_id=%d err=%v", userID, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi hệ thống"})
		return
	}

	if current != lockValue {
		log.Printf("[admin][lock.heartbeat.denied] user_id=%d holder=%s", userID, current)
		writeJSON(w, http.StatusForbidden, map[string]string{"message": "Quyền quản lý đã bị chiếm bởi người khác"})
		return
	}

	h.redis.Expire(ctx, adminControlLockKey, 45*time.Second)
	log.Printf("[admin][lock.heartbeat.ok] user_id=%d ttl_seconds=45", userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Duy trì thành công"})
}

func (h *AdminHandler) ReleaseLock(w http.ResponseWriter, r *http.Request) {
	claims, _ := authmiddleware.CurrentClaims(r.Context())
	userID := claims.UserID
	lockValue := strconv.FormatInt(userID, 10)

	ctx := r.Context()
	current, err := h.redis.Get(ctx, adminControlLockKey).Result()
	if err == nil && current == lockValue {
		h.redis.Del(ctx, adminControlLockKey)
		log.Printf("[admin][lock.release.ok] user_id=%d", userID)
	} else {
		log.Printf("[admin][lock.release.skip] user_id=%d holder=%s err=%v", userID, current, err)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Đã thoát chế độ quản lý"})
}
