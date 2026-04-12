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
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/service"
	"gin/internal/support/logger"
)

func main() {
	logger.Init("storage/logs/gin.log")
	config := app.LoadConfig()

	db, err := platformpg.Open(config.DatabaseURL)
	if err != nil {
		log.Fatalf("Khởi tạo engine thất bại (db): %v", err)
	}
	platformpg.ConfigurePool(db, platformpg.PoolConfig{
		MaxOpenConns:    config.DBMaxOpenConns,
		MaxIdleConns:    config.DBMaxIdleConns,
		ConnMaxLifetime: config.DBConnMaxLifetime,
		ConnMaxIdleTime: config.DBConnMaxIdleTime,
	})
	log.Printf("[db.pool] max_open=%d max_idle=%d conn_max_lifetime=%s conn_max_idle_time=%s", config.DBMaxOpenConns, config.DBMaxIdleConns, config.DBConnMaxLifetime, config.DBConnMaxIdleTime)
	defer db.Close()

	redisClient, err := platformredis.Open(context.Background(), config.RedisAddr, config.RedisPassword, config.RedisDB)
	if err != nil {
		log.Fatalf("Khởi tạo engine thất bại (redis): %v", err)
	}
	defer redisClient.Close()

	gameRepository := repopg.NewGameRepository(db)
	walletRepository := repopg.NewWalletRepository(db)
	broker := realtime.NewBroker(redisClient)
	walletService := service.NewWalletService(walletRepository, broker)
	playRoomService := service.NewPlayRoomService(gameRepository, walletRepository, walletService, redisClient, broker)
	engineService := service.NewRoomEngineService(gameRepository, redisClient, playRoomService, walletService, time.Second)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("[room-engine] đang chạy 24/7")
	if err := engineService.Run(ctx); err != nil {
		log.Fatalf("[room-engine] dừng do lỗi: %v", err)
	}
}
