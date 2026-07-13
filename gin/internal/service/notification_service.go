package service

import (
	"context"
	"strings"
	"time"

	"gin/internal/domain/notification"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/message"

	goredis "github.com/redis/go-redis/v9"
)

type NotificationService struct {
	repository       *repopg.NotificationRepository
	redis            *goredis.Client
	contentAssetBase string
}

func NewNotificationService(repository *repopg.NotificationRepository, redis *goredis.Client, contentAssetBase string) *NotificationService {
	return &NotificationService{
		repository:       repository,
		redis:            redis,
		contentAssetBase: strings.TrimRight(strings.TrimSpace(contentAssetBase), "/"),
	}
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

	records, total, unreadCount, err := s.repository.ListForUser(ctx, userID, page, pageSize)
	if err != nil {
		return notification.ListResponse{}, err
	}

	items := make([]notification.Item, 0, len(records))
	for _, record := range records {
		items = append(items, notification.Item{
			ID:             record.ID,
			Title:          record.Title,
			Body:           record.Body,
			ImageURL:       stringPtrOrNil(buildPublicAssetURL(s.contentAssetBase, firstNonEmptyStringPtr(record.ImagePath))),
			Status:         record.Status,
			Audience:       record.Audience,
			PublishAt:      formatNullableNotificationTime(record.PublishAt),
			ExpiresAt:      formatNullableNotificationTime(record.ExpiresAt),
			CreatedAt:      formatNotificationTime(record.CreatedAt),
			IsRead:         record.ReadAt != nil,
			ReadAt:         formatNullableNotificationTime(record.ReadAt),
			ResponseStatus: record.ResponseStatus,
			RespondedAt:    formatNullableNotificationTime(record.RespondedAt),
			CanRespond:     s.canRespond(record),
		})
	}

	return notification.ListResponse{
		Message:     message.NotificationsListSuccess,
		Page:        page,
		PageSize:    pageSize,
		Total:       total,
		TotalPages:  calcNotificationTotalPages(total, pageSize),
		UnreadCount: unreadCount,
		Items:       items,
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
		ReadAt:  formatNotificationTime(readAt),
	}, nil
}

func (s *NotificationService) Respond(ctx context.Context, userID, notificationID int64, action string) (notification.RespondResponse, error) {
	if userID == 0 {
		return notification.RespondResponse{}, ErrUnauthorized
	}

	requestedStatus := notification.ResponseStatusCanceled
	if action == notification.ResponseActionConfirm {
		requestedStatus = notification.ResponseStatusConfirmed
	}

	snapshot := loadSystemSnapshot(ctx, s.redis)
	forceCancel := snapshot.NotificationImageForceCancelEnabled != nil && *snapshot.NotificationImageForceCancelEnabled

	result, err := s.repository.Respond(ctx, userID, notificationID, requestedStatus, forceCancel)
	if err != nil {
		return notification.RespondResponse{}, err
	}

	return notification.RespondResponse{
		Message:        message.NotificationResponseSuccess,
		ID:             notificationID,
		ResponseStatus: result.ResponseStatus,
		RespondedAt:    formatNotificationTime(result.RespondedAt),
		ReadAt:         formatNotificationTime(result.ReadAt),
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

func formatNotificationTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(dbDateTimeLayout)
}

func formatNullableNotificationTime(value *time.Time) *string {
	if value == nil || value.IsZero() {
		return nil
	}

	formatted := value.Format(dbDateTimeLayout)
	return &formatted
}

func (s *NotificationService) canRespond(record repopg.NotificationRecord) bool {
	if record.Audience != 2 {
		return false
	}
	if strings.TrimSpace(firstNonEmptyStringPtr(record.ImagePath)) == "" {
		return false
	}
	if record.ResponseStatus == nil {
		return true
	}

	return *record.ResponseStatus == notification.ResponseStatusPending
}
