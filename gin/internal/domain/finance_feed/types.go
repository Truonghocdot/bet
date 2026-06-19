package finance_feed

type Item struct {
	ID          int64   `json:"id"`
	MaskedCode  string  `json:"masked_code"`
	MaskedPhone string  `json:"masked_phone"`
	StatusLabel string  `json:"status_label"`
	CreatedAt   string  `json:"created_at"`
	Channel     *string `json:"channel,omitempty"`
}

type FeedResponse struct {
	Message string `json:"message"`
	Items   []Item `json:"items"`
}
