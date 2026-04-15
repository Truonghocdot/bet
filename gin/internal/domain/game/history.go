package game

import "time"

type HistoryListItem struct {
	PeriodNo   string    `json:"period_no"`
	PeriodIndex int64    `json:"period_index"`
	Result     string    `json:"result"`
	BigSmall   string    `json:"big_small"`
	Color      string    `json:"color"`
	DrawAt     time.Time `json:"draw_at"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type HistoryListResponse struct {
	Message    string           `json:"message"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
	Items      []HistoryListItem `json:"items"`
}

type BetTicketHistoryItem struct {
	ID             int64      `json:"id"`
	PeriodNo       string     `json:"period_no"`
	PeriodIndex    int64      `json:"period_index"`
	Result         string     `json:"result"`
	BigSmall       string     `json:"big_small"`
	Color          string     `json:"color"`
	Stake          string     `json:"stake"`
	OriginalAmount string     `json:"original_amount,omitempty"`
	TaxAmount      string     `json:"tax_amount,omitempty"`
	NetAmount      string     `json:"net_amount,omitempty"`
	ActualPayout   string     `json:"actual_payout"`
	ProfitLoss     string     `json:"profit_loss"`
	SettledAt      *time.Time `json:"settled_at,omitempty"`
	Status         string     `json:"status"`
	ItemsCount     int        `json:"items_count"`
	CreatedAt      time.Time  `json:"created_at"`
}

type BetTicketHistoryResponse struct {
	Message    string                `json:"message"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	Total      int                   `json:"total"`
	TotalPages int                   `json:"total_pages"`
	Items      []BetTicketHistoryItem `json:"items"`
}
