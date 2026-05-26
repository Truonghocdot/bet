package service

import (
	"context"
	"time"

	"gin/internal/domain/notification"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/message"
)

type NotificationService struct {
	repository *repopg.NotificationRepository
}

func NewNotificationService(repository *repopg.NotificationRepository) *NotificationService {
	return &NotificationService{repository: repository}
}

func (s *NotificationService) List(ctx context.Context, userID int64, page, pageSize int) (notification.ListResponse, error) {
	if userID == 0 {
		return notification.ListResponse{}, ErrUnauthorized
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	records, total, err := s.repository.ListForUser(ctx, userID, page, pageSize)
	if err != nil {
		return notification.ListResponse{}, err
	}

	items := make([]notification.Item, 0, len(records))
	for _, record := range records {
		items = append(items, notification.Item{
			ID:        record.ID,
			Title:     record.Title,
			Body:      record.Body,
			Status:    record.Status,
			Audience:  record.Audience,
			PublishAt: normalizeNotificationTimePtr(record.PublishAt),
			ExpiresAt: normalizeNotificationTimePtr(record.ExpiresAt),
			CreatedAt: normalizeNotificationTime(record.CreatedAt),
			IsRead:    record.ReadAt != nil,
			ReadAt:    normalizeNotificationTimePtr(record.ReadAt),
		})
	}

	return notification.ListResponse{
		Message:    message.NotificationsListSuccess,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: calcNotificationTotalPages(total, pageSize),
		Items:      items,
	}, nil
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID int64) (notification.MarkReadResponse, error) {
	if userID == 0 {
		return notification.MarkReadResponse{}, ErrUnauthorized
	}

	readAt, err := s.repository.MarkRead(ctx, userID, notificationID)
	if err != nil {
		return notification.MarkReadResponse{}, err
	}

	return notification.MarkReadResponse{
		Message: message.NotificationReadSuccess,
		ID:      notificationID,
		ReadAt:  normalizeNotificationTime(readAt),
	}, nil
}

func calcNotificationTotalPages(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := total / pageSize
	if total%pageSize != 0 {
		pages++
	}
	if pages == 0 {
		pages = 1
	}
	return pages
}

func normalizeNotificationTime(value time.Time) time.Time {
	return value.In(clock.Location())
}

func normalizeNotificationTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	normalized := value.In(clock.Location())
	return &normalized
}
