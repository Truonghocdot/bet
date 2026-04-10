package outbox

import (
	"context"
	"log"
	"time"
)

type Event struct {
	ID         string
	Name       string
	OccurredAt time.Time
	Payload    map[string]any
}

type Publisher interface {
	Publish(ctx context.Context, event Event) error
}

type NoopPublisher struct{}

func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

func (p *NoopPublisher) Publish(_ context.Context, event Event) error {
	log.Printf("[outbox] da phat su kien=%s id=%s payload=%v", event.Name, event.ID, event.Payload)
	return nil
}
