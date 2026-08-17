package chat

import "time"

type Message struct {
	ID          int64     `json:"id"`
	DisplayName string    `json:"display_name"`
	Body        string    `json:"body"`
	ActorType   string    `json:"actor_type"`
	CreatedAt   time.Time `json:"created_at"`
}

type ListResponse struct {
	Items      []Message `json:"items"`
	NextCursor *int64    `json:"next_cursor,omitempty"`
}

type CreateRequest struct {
	Body string `json:"body"`
}

type CreateResponse struct {
	Message Message `json:"message"`
}
