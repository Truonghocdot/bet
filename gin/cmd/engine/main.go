package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"gin/internal/app"
	platformpg "gin/internal/platform/postgres"
	platformredis "gin/internal/platform/redis"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
)

func main() {
	config := app.LoadConfig()

	db, err := platformpg.Open(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Khởi tạo engine thất bại (db): %v", err)
	}
	defer db.Close()

	redisClient, err := platformredis.Open(context.Background(), config.RedisAddr, config.RedisPassword, config.RedisDB)
	if err != nil {
		log.Fatalf("Khởi tạo engine thất bại (redis): %v", err)
	}
	defer redisClient.Close()

	gameRepository := repopg.NewGameRepository(db)
	engineService := service.NewRoomEngineService(gameRepository, redisClient, time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("[room-engine] đang chạy 24/7")
	if err := engineService.Run(ctx); err != nil {
		log.Fatalf("[room-engine] dừng do lỗi: %v", err)
	}
}
