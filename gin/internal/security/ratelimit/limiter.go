package ratelimit

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Result struct {
	Allowed    bool
	Count      int64
	Limit      int64
	RetryAfter time.Duration
}

type Limiter struct {
	client *goredis.Client
}

func New(client *goredis.Client) *Limiter {
	return &Limiter{client: client}
}

func (l *Limiter) HitWindow(ctx context.Context, key string, limit int64, window time.Duration) (Result, error) {
	count, err := l.client.Incr(ctx, key).Result()
	if err != nil {
		return Result{}, err
	}

	if count == 1 {
		if err := l.client.Expire(ctx, key, window).Err(); err != nil {
			return Result{}, err
		}
	}

	ttl, err := l.client.TTL(ctx, key).Result()
	if err != nil {
		return Result{}, err
	}

	return Result{
		Allowed:    count <= limit,
		Count:      count,
		Limit:      limit,
		RetryAfter: ttl,
	}, nil
}

func (l *Limiter) StartCooldown(ctx context.Context, key string, duration time.Duration) (bool, time.Duration, error) {
	started, err := l.client.SetNX(ctx, key, "1", duration).Result()
	if err != nil {
		return false, 0, err
	}

	if started {
		return true, 0, nil
	}

	ttl, err := l.client.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	return false, ttl, nil
}

func (l *Limiter) IsLocked(ctx context.Context, key string) (bool, time.Duration, error) {
	ttl, err := l.client.TTL(ctx, key).Result()
	if err != nil {
		return false, 0, err
	}

	if ttl > 0 {
		return true, ttl, nil
	}

	return false, 0, nil
}

func (l *Limiter) Lock(ctx context.Context, key string, duration time.Duration) error {
	return l.client.Set(ctx, key, "1", duration).Err()
}

func (l *Limiter) Clear(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	return l.client.Del(ctx, keys...).Err()
}
