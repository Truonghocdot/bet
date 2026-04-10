package redis

import (
	"context"
	"fmt"

	"gin/internal/support/message"
	goredis "github.com/redis/go-redis/v9"
)

func Open(ctx context.Context, addr, password string, db int) (*goredis.Client, error) {
	if addr == "" {
		return nil, fmt.Errorf(message.RedisAddressRequired)
	}

	client := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return client, nil
}
