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
	Message string          `json:"message"`
	Wallets []WalletBalance `json:"wallets"`
}
