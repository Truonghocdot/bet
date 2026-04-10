package game

import "time"

type GameType string

const (
	GameWingo   GameType = "wingo"
	GameK3      GameType = "k3"
	GameLottery GameType = "lottery"
)

type JoinRequest struct {
	UserID string `json:"user_id"`
}

type JoinResponse struct {
	ConnectionID string    `json:"connection_id"`
	GameType     GameType  `json:"game_type"`
	JoinedAt     time.Time `json:"joined_at"`
	Message      string    `json:"message"`
}

type BetItem struct {
	OptionType string `json:"option_type"`
	OptionKey  string `json:"option_key"`
	Stake      string `json:"stake"`
}

type PlaceBetRequest struct {
	RequestID string    `json:"request_id"`
	UserID    string    `json:"user_id"`
	GameType  GameType  `json:"game_type"`
	PeriodID  string    `json:"period_id"`
	Items     []BetItem `json:"items"`
}

type PlaceBetResponse struct {
	RequestID    string    `json:"request_id"`
	ConnectionID string    `json:"connection_id"`
	GameType     GameType  `json:"game_type"`
	Status       string    `json:"status"`
	AcceptedAt   time.Time `json:"accepted_at"`
	Message      string    `json:"message"`
}
