package service

import (
	"context"
	"fmt"
	"log"

	"gate/internal/domain/event"
)

type NotificationService struct{}

func NewNotificationService() *NotificationService {
	return &NotificationService{}
}

func (s *NotificationService) Send(_ context.Context, request event.NotificationRequest) error {
	if request.Channel == "" {
		return fmt.Errorf("channel is required")
	}

	if request.Target == "" {
		return fmt.Errorf("target is required")
	}

	log.Printf("[gate] notify channel=%s target=%s subject=%s", request.Channel, request.Target, request.Subject)
	return nil
}
