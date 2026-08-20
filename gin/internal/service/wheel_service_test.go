package service

import (
	"errors"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestWheelGetAndDeleteIsAtomicAndRedis6Compatible(t *testing.T) {
	client := testRoomEngineRedisClient(t)
	service := &WheelService{redis: client}
	key := "wheel:test:one-time-token"
	if err := client.Set(t.Context(), key, "payload", 0).Err(); err != nil {
		t.Fatalf("set token: %v", err)
	}

	value, err := service.getAndDelete(t.Context(), key)
	if err != nil {
		t.Fatalf("consume token: %v", err)
	}
	if string(value) != "payload" {
		t.Fatalf("consume token = %q, want payload", value)
	}
	if _, err := service.getAndDelete(t.Context(), key); !errors.Is(err, goredis.Nil) {
		t.Fatalf("second consume error = %v, want redis.Nil", err)
	}
}
