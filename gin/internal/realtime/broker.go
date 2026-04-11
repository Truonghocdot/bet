package realtime

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Message struct {
	Event string
	Data  json.RawMessage
}

type Broker struct {
	redis *goredis.Client

	mu     sync.Mutex
	topics map[string]*topicStream
}

type topicStream struct {
	pubsub      *goredis.PubSub
	subscribers map[uint64]chan Message
	nextID      uint64
}

type envelope struct {
	Event       string          `json:"event"`
	Data        json.RawMessage `json:"data"`
	PublishedAt time.Time       `json:"published_at"`
}

func NewBroker(redis *goredis.Client) *Broker {
	return &Broker{
		redis:  redis,
		topics: make(map[string]*topicStream),
	}
}

func (b *Broker) Publish(ctx context.Context, topic string, event string, payload any) error {
	if b == nil || b.redis == nil || topic == "" || event == "" {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(envelope{
		Event:       event,
		Data:        data,
		PublishedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	return b.redis.Publish(ctx, topic, raw).Err()
}

func (b *Broker) Subscribe(ctx context.Context, topic string) (<-chan Message, func(), error) {
	if b == nil || b.redis == nil || topic == "" {
		ch := make(chan Message)
		close(ch)
		return ch, func() {}, nil
	}

	b.mu.Lock()
	stream := b.topics[topic]
	if stream == nil {
		pubsub := b.redis.Subscribe(ctx, topic)
		stream = &topicStream{
			pubsub:      pubsub,
			subscribers: make(map[uint64]chan Message),
		}
		b.topics[topic] = stream
		go b.runTopic(topic, stream)
	}

	stream.nextID++
	id := stream.nextID
	ch := make(chan Message, 16)
	stream.subscribers[id] = ch
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		current := b.topics[topic]
		if current == nil {
			close(ch)
			return
		}
		if current.subscribers[id] != nil {
			delete(current.subscribers, id)
			close(ch)
		}
		if len(current.subscribers) == 0 {
			_ = current.pubsub.Close()
			delete(b.topics, topic)
		}
	}

	return ch, unsubscribe, nil
}

func (b *Broker) runTopic(topic string, stream *topicStream) {
	redisCh := stream.pubsub.Channel()
	for msg := range redisCh {
		var payload envelope
		if err := json.Unmarshal([]byte(msg.Payload), &payload); err != nil {
			continue
		}

		b.mu.Lock()
		subscribers := make([]chan Message, 0, len(stream.subscribers))
		for _, subscriber := range stream.subscribers {
			subscribers = append(subscribers, subscriber)
		}
		b.mu.Unlock()

		message := Message{
			Event: payload.Event,
			Data:  payload.Data,
		}
		for _, subscriber := range subscribers {
			select {
			case subscriber <- message:
			default:
			}
		}
	}
}
