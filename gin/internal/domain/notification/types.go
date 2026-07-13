package notification

const (
	ResponseStatusPending   = 1
	ResponseStatusConfirmed = 2
	ResponseStatusCanceled  = 3

	ResponseActionConfirm = "confirm"
	ResponseActionCancel  = "cancel"
)

type Item struct {
	ID             int64   `json:"id"`
	Title          string  `json:"title"`
	Body           string  `json:"body"`
	ImageURL       *string `json:"image_url,omitempty"`
	Status         int     `json:"status"`
	Audience       int     `json:"audience"`
	PublishAt      *string `json:"publish_at,omitempty"`
	ExpiresAt      *string `json:"expires_at,omitempty"`
	CreatedAt      string  `json:"created_at"`
	IsRead         bool    `json:"is_read"`
	ReadAt         *string `json:"read_at,omitempty"`
	ResponseStatus *int    `json:"response_status,omitempty"`
	RespondedAt    *string `json:"responded_at,omitempty"`
	CanRespond     bool    `json:"can_respond"`
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

type RespondResponse struct {
	Message        string `json:"message"`
	ID             int64  `json:"id"`
	ResponseStatus int    `json:"response_status"`
	RespondedAt    string `json:"responded_at"`
	ReadAt         string `json:"read_at"`
}
