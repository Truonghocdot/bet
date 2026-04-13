package wallet

import "time"

type WalletBalance struct {
	ID            int64      `json:"id"`
	Unit          int        `json:"unit"`
	UnitCode      string     `json:"unit_code"`
	UnitLabel     string     `json:"unit_label"`
	Balance       string     `json:"balance"`
	LockedBalance string     `json:"locked_balance"`
	Status        int        `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type WalletSummaryResponse struct {
	Message      string          `json:"message"`
	ExchangeRate string          `json:"exchange_rate"`
	Wallets      []WalletBalance `json:"wallets"`
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
