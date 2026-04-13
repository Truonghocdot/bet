package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/message"
)

type AdminHandler struct {
	gameRepository *repopg.GameRepository
	broker         *realtime.Broker
}

func NewAdminHandler(gameRepo *repopg.GameRepository, broker *realtime.Broker) *AdminHandler {
	return &AdminHandler{gameRepository: gameRepo, broker: broker}
}

type adminRoomDetail struct {
	Code     string                   `json:"code"`
	GameType int                      `json:"game_type"`
	Period   *repopg.GamePeriodRecord `json:"period"`
	Stats    []repopg.PeriodBetStats  `json:"bet_stats"`
}

type adminStatsResponse struct {
	ServerTime time.Time         `json:"server_time"`
	Rooms      []adminRoomDetail `json:"rooms"`
}

func (h *AdminHandler) buildStatsSnapshot(r *http.Request) (adminStatsResponse, error) {
	ctx := r.Context()
	stats, err := h.gameRepository.ListAllRoomsWithCurrentPeriod(ctx)
	if err != nil {
		return adminStatsResponse{}, err
	}

	rooms := make([]adminRoomDetail, 0, len(stats))
	for _, s := range stats {
		var betStats []repopg.PeriodBetStats
		if s.Period != nil {
			betStats, _ = h.gameRepository.GetPeriodBetStats(ctx, s.Period.ID)
		}
		rooms = append(rooms, adminRoomDetail{
			Code:     s.Room.Code,
			GameType: s.Room.GameType,
			Period:   s.Period,
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
