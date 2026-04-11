package notification

import "time"

type Item struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Status    int        `json:"status"`
	Audience  int        `json:"audience"`
	PublishAt *time.Time `json:"publish_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	IsRead    bool       `json:"is_read"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

type ListResponse struct {
	Message    string `json:"message"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Total      int    `json:"total"`
	TotalPages int    `json:"total_pages"`
	Items      []Item `json:"items"`
}

type MarkReadResponse struct {
	Message string    `json:"message"`
	ID      int64     `json:"id"`
	ReadAt  time.Time `json:"read_at"`
}
