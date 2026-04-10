package ws

import (
	"sync"
	"time"
)

type Session struct {
	ConnectionID string
	UserID       string
	GameType     string
	JoinedAt     time.Time
	LastSeenAt   time.Time
}

type Hub struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[string]Session),
	}
}

func (h *Hub) Upsert(session Session) Session {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sessions[session.ConnectionID] = session
	return session
}

func (h *Hub) Get(connectionID string) (Session, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	session, ok := h.sessions[connectionID]
	return session, ok
}
