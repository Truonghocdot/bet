package game

import "time"

type RoomItem struct {
	Code             string `json:"code"`
	GameType         string `json:"game_type"`
	DurationSeconds  int    `json:"duration_seconds"`
	BetCutoffSeconds int    `json:"bet_cutoff_seconds"`
	Status           string `json:"status"`
	SortOrder        int    `json:"sort_order"`
}

type RoomListResponse struct {
	Message string     `json:"message"`
	Items   []RoomItem `json:"items"`
}

type RoomPeriod struct {
	ID        int64     `json:"id"`
	PeriodNo  string    `json:"period_no"`
	PeriodIndex int64   `json:"period_index"`
	Status    string    `json:"status"`
	OpenAt    time.Time `json:"open_at"`
	BetLockAt time.Time `json:"bet_lock_at"`
	DrawAt    time.Time `json:"draw_at"`
}

type RoomStateResponse struct {
	Message       string            `json:"message"`
	ServerTime    time.Time         `json:"server_time"`
	Room          RoomItem          `json:"room"`
	CurrentPeriod RoomPeriod        `json:"current_period"`
	RecentResults []HistoryListItem `json:"recent_results"`
}

type RoomBetRequest struct {
	RequestID string    `json:"request_id"`
	PeriodID  string    `json:"period_id"`
	Items     []BetItem `json:"items"`
}

type RoomBetResponse struct {
	RequestID  string    `json:"request_id"`
	RoomCode   string    `json:"room_code"`
	Status     string    `json:"status"`
	AcceptedAt time.Time `json:"accepted_at"`
	Message    string    `json:"message"`
}

var defaultRoomByGame = map[GameType]string{
	GameWingo:   "wingo_1m",
	GameK3:      "k3_1m",
	GameLottery: "lottery_1m",
}

func DefaultRoomCode(gameType GameType) (string, bool) {
	roomCode, ok := defaultRoomByGame[gameType]
	return roomCode, ok
}
