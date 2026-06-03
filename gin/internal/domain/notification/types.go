package notification

type Item struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Status    int     `json:"status"`
	Audience  int     `json:"audience"`
	PublishAt *string `json:"publish_at,omitempty"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	CreatedAt string  `json:"created_at"`
	IsRead    bool    `json:"is_read"`
	ReadAt    *string `json:"read_at,omitempty"`
}

type ListResponse struct {
	Message     string `json:"message"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
	Total       int    `json:"total"`
	TotalPages  int    `json:"total_pages"`
	UnreadCount int    `json:"unread_count"`
	Items       []Item `json:"items"`
}

type MarkReadResponse struct {
	Message string `json:"message"`
	ID      int64  `json:"id"`
	ReadAt  string `json:"read_at"`
}
