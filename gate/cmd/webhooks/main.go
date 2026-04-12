package main

import (
	"log"

	"gate/internal/app"
	"gate/internal/support/logger"
)

func main() {
	logger.Init("storage/logs/gate.log")
	application, err := app.New()
	if err != nil {
		log.Fatalf("bootstrap gate failed: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("gate stopped: %v", err)
	}
}
