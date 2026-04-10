package main

import (
	"log"

	"gin/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("Khoi tao dich vu gin that bai: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Dich vu gin da dung: %v", err)
	}
}
