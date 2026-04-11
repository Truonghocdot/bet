package game

import "time"

type HistoryListItem struct {
	PeriodNo   string    `json:"period_no"`
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
	ID         int64     `json:"id"`
	PeriodNo   string    `json:"period_no"`
	Result     string    `json:"result"`
	BigSmall   string    `json:"big_small"`
	Color      string    `json:"color"`
	Stake      string    `json:"stake"`
	Status     string    `json:"status"`
	ItemsCount int       `json:"items_count"`
	CreatedAt  time.Time `json:"created_at"`
}

type BetTicketHistoryResponse struct {
	Message    string                `json:"message"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	Total      int                   `json:"total"`
	TotalPages int                   `json:"total_pages"`
	Items      []BetTicketHistoryItem `json:"items"`
}
