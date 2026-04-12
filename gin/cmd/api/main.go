package main

import (
	"log"

	"gin/internal/app"
	"gin/internal/support/logger"
)

func main() {
	logger.Init("storage/logs/gin.log")

	application, err := app.New()
	if err != nil {
		log.Fatalf("Khởi tạo dịch vụ thât bại: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Dịch vụ gin đã dừng: %v", err)
	}
}
