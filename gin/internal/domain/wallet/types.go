package wallet

import "time"

type WalletBalance struct {
	ID                  int64     `json:"id"`
	Unit                int       `json:"unit"`
	UnitCode            string    `json:"unit_code"`
	UnitLabel           string    `json:"unit_label"`
	Balance             string    `json:"balance"`
	LockedBalance       string    `json:"locked_balance"`
	WithdrawCreditLimit string    `json:"withdraw_credit_limit"`
	WithdrawAvailable   string    `json:"withdraw_available"`
	Status              int       `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type WithdrawPolicyDisplay struct {
	Enabled           bool   `json:"enabled"`
	FeePercent        string `json:"fee_percent"`
	RequiredBetVolume string `json:"required_bet_volume"`
	MaxTimesPerDay    int    `json:"max_times_per_day"`
	MinAmount         string `json:"min_amount"`
	MaxAmount         string `json:"max_amount"`
}

type MarqueeDisplay struct {
	Enabled  bool     `json:"enabled"`
	Messages []string `json:"messages"`
}

type PopupDisplay struct {
	Message    *string `json:"message,omitempty"`
	LatestNews *string `json:"latest_news,omitempty"`
}

type WalletSummaryResponse struct {
	Message          string                `json:"message"`
	ExchangeRate     string                `json:"exchange_rate"`
	TelegramCskhLink string                `json:"telegram_cskh_link,omitempty"`
	Marquee          MarqueeDisplay        `json:"marquee"`
	Popup            PopupDisplay          `json:"popup"`
	WithdrawPolicy   WithdrawPolicyDisplay `json:"withdraw_policy"`
	Wallets          []WalletBalance       `json:"wallets"`
}

type ExchangeRequest struct {
	FromUnit int    `json:"from_unit" binding:"required,oneof=1 2"`
	ToUnit   int    `json:"to_unit" binding:"required,oneof=1 2"`
	Amount   string `json:"amount" binding:"required"`
}

type ExchangeResponse struct {
	Message      string `json:"message"`
	FromUnit     int    `json:"from_unit"`
	ToUnit       int    `json:"to_unit"`
	FromAmount   string `json:"from_amount"`
	ToAmount     string `json:"to_amount"`
	ExchangeRate string `json:"exchange_rate"`
}
