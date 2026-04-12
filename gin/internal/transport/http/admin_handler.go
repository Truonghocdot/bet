package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/message"
)

type AdminHandler struct {
	gameRepository *repopg.GameRepository
}

func NewAdminHandler(gameRepo *repopg.GameRepository) *AdminHandler {
	return &AdminHandler{gameRepository: gameRepo}
}

func (h *AdminHandler) ListRoomStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	stats, err := h.gameRepository.ListAllRoomsWithCurrentPeriod(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi khi tải danh sách phòng"})
		return
	}

	type RoomDetail struct {
		Code     string                   `json:"code"`
		GameType int                      `json:"game_type"`
		Period   *repopg.GamePeriodRecord `json:"period"`
		Stats    []repopg.PeriodBetStats  `json:"bet_stats"`
	}

	type Response struct {
		ServerTime time.Time    `json:"server_time"`
		Rooms      []RoomDetail `json:"rooms"`
	}

	rooms := make([]RoomDetail, 0, len(stats))
	for _, s := range stats {
		var betStats []repopg.PeriodBetStats
		if s.Period != nil {
			betStats, _ = h.gameRepository.GetPeriodBetStats(ctx, s.Period.ID)
		}
		rooms = append(rooms, RoomDetail{
			Code:     s.Room.Code,
			GameType: s.Room.GameType,
			Period:   s.Period,
			Stats:    betStats,
		})
	}

	writeJSON(w, http.StatusOK, Response{
		ServerTime: clock.Now(),
		Rooms:      rooms,
	})
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
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Lỗi khi lưu kết quả can thiệp"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Đã lưu kết quả dự kiến"})
}
