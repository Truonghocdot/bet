package wheel

type Invitation struct {
	ID            string  `json:"id"`
	CampaignName  string  `json:"campaign_name"`
	Status        string  `json:"status"`
	ExpiresAt     *string `json:"expires_at,omitempty"`
	SeenAt        *string `json:"seen_at,omitempty"`
	SessionID     *string `json:"session_id,omitempty"`
	SessionStatus *string `json:"session_status,omitempty"`
}

type InvitationListResponse struct {
	Items []Invitation `json:"items"`
}

type LaunchResponse struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in"`
}

type ExchangeRequest struct {
	LaunchCode string `json:"launch_code"`
}

type ExchangeResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiresIn   int64      `json:"expires_in"`
	Invitation  Invitation `json:"invitation"`
}

type Round struct {
	RoundNo     int     `json:"round_no"`
	Status      string  `json:"status"`
	SegmentKey  *string `json:"segment_key,omitempty"`
	ResultLabel *string `json:"result_label,omitempty"`
	PrizeAmount *string `json:"prize_amount,omitempty"`
	SpunAt      *string `json:"spun_at,omitempty"`
}

type Reward struct {
	RoundNo int     `json:"round_no"`
	Amount  string  `json:"amount"`
	Status  string  `json:"status"`
	PaidAt  *string `json:"paid_at,omitempty"`
}

type State struct {
	ServerNow            string   `json:"server_now"`
	InvitationID         string   `json:"invitation_id"`
	CampaignName         string   `json:"campaign_name"`
	SessionID            *string  `json:"session_id,omitempty"`
	SessionStatus        string   `json:"session_status"`
	StartedAt            *string  `json:"started_at,omitempty"`
	EndsAt               *string  `json:"ends_at,omitempty"`
	CurrentRound         int      `json:"current_round"`
	NextRoundAvailableAt *string  `json:"next_round_available_at,omitempty"`
	SpinDurationSeconds  int      `json:"spin_duration_seconds"`
	Rounds               []Round  `json:"rounds"`
	PaidRewards          []Reward `json:"paid_rewards"`
	TotalReward          string   `json:"total_reward"`
}

type SpinResponse struct {
	State  State `json:"state"`
	Result Round `json:"result"`
}

type SocketTicketResponse struct {
	Ticket    string `json:"ticket"`
	ExpiresIn int64  `json:"expires_in"`
}

type ChatMessage struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
	Body        string `json:"body"`
	ActorType   string `json:"actor_type"`
	CreatedAt   string `json:"created_at"`
}

type ChatListResponse struct {
	Items      []ChatMessage `json:"items"`
	NextCursor *int64        `json:"next_cursor,omitempty"`
}

type ChatCreateRequest struct {
	Body string `json:"body"`
}

type ChatCreateResponse struct {
	Message ChatMessage `json:"message"`
}

type Access struct {
	UserID       int64 `json:"user_id"`
	InvitationID int64 `json:"invitation_id"`
}
