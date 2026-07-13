package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gin/internal/domain/notification"
	"gin/internal/support/message"
)

const (
	notificationStatusPublished = 2
	notificationAudienceAll     = 1
	notificationAudienceUsers   = 2
)

var (
	ErrNotificationNotFound           = errors.New(message.NotificationNotFound)
	ErrNotificationResponseNotAllowed = errors.New("notification.response.not_allowed")
)

type NotificationRepository struct {
	db *sql.DB
}

type NotificationRecord struct {
	ID             int64
	Title          string
	Body           string
	ImagePath      *string
	Status         int
	Audience       int
	PublishAt      *time.Time
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	ReadAt         *time.Time
	ResponseStatus *int
	RespondedAt    *time.Time
}

type NotificationRespondResult struct {
	ResponseStatus int
	RespondedAt    time.Time
	ReadAt         time.Time
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) ListForUser(ctx context.Context, userID int64, page, pageSize int) ([]NotificationRecord, int, int, error) {
	page, pageSize = normalizePagination(page, pageSize)

	var total int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from notifications n
		left join notification_targets nt
			on nt.notification_id = n.id
		   and nt.user_id = $1
		where n.status = $2
		  and (n.publish_at is null or n.publish_at <= now())
		  and (n.expires_at is null or n.expires_at > now())
		  and (
			n.audience = $3
			or (n.audience = $4 and nt.user_id is not null)
		  )
	`, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers).Scan(&total); err != nil {
		return nil, 0, 0, err
	}

	var unreadCount int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from notifications n
		left join notification_targets nt
			on nt.notification_id = n.id
		   and nt.user_id = $1
		left join notification_reads nr
			on nr.notification_id = n.id
		   and nr.user_id = $1
		where n.status = $2
		  and nr.read_at is null
		  and (n.publish_at is null or n.publish_at <= now())
		  and (n.expires_at is null or n.expires_at > now())
		  and (
			n.audience = $3
			or (n.audience = $4 and nt.user_id is not null)
		  )
	`, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers).Scan(&unreadCount); err != nil {
		return nil, 0, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
			select
				n.id,
				n.title,
				coalesce(n.body, ''),
				n.image_path,
				n.status,
				n.audience,
				n.publish_at,
				n.expires_at,
				n.created_at,
				nr.read_at,
				nt.response_status,
				nt.responded_at
			from notifications n
			left join notification_targets nt
				on nt.notification_id = n.id
			   and nt.user_id = $1
			left join notification_reads nr
				on nr.notification_id = n.id
			   and nr.user_id = $1
		where n.status = $2
		  and (n.publish_at is null or n.publish_at <= now())
		  and (n.expires_at is null or n.expires_at > now())
			  and (
				n.audience = $3
				or (n.audience = $4 and nt.user_id is not null)
			  )
			order by coalesce(n.publish_at, n.created_at) desc, n.id desc
			limit $5 offset $6
	`, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	records := make([]NotificationRecord, 0)
	for rows.Next() {
		var record NotificationRecord
		var imagePath sql.NullString
		var responseStatus sql.NullInt64
		if err := rows.Scan(
			&record.ID,
			&record.Title,
			&record.Body,
			&imagePath,
			&record.Status,
			&record.Audience,
			&record.PublishAt,
			&record.ExpiresAt,
			&record.CreatedAt,
			&record.ReadAt,
			&responseStatus,
			&record.RespondedAt,
		); err != nil {
			return nil, 0, 0, err
		}
		if imagePath.Valid {
			record.ImagePath = &imagePath.String
		}
		if responseStatus.Valid {
			value := int(responseStatus.Int64)
			record.ResponseStatus = &value
		}
		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	return records, total, unreadCount, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, notificationID int64) (time.Time, error) {
	if !r.canAccess(ctx, userID, notificationID) {
		return time.Time{}, ErrNotificationNotFound
	}

	var readAt time.Time
	if err := r.db.QueryRowContext(ctx, `
		insert into notification_reads (notification_id, user_id, read_at)
		values ($1, $2, now())
		on conflict (notification_id, user_id)
		do update set read_at = excluded.read_at
		returning read_at
	`, notificationID, userID).Scan(&readAt); err != nil {
		return time.Time{}, err
	}

	return readAt, nil
}

func (r *NotificationRepository) Respond(ctx context.Context, userID, notificationID int64, requestedStatus int, forceCancel bool) (NotificationRespondResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return NotificationRespondResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	record, err := r.findResponseTargetForUpdate(ctx, tx, userID, notificationID)
	if err != nil {
		return NotificationRespondResult{}, err
	}

	if record.ImagePath == nil || *record.ImagePath == "" || record.Audience != notificationAudienceUsers {
		return NotificationRespondResult{}, ErrNotificationResponseNotAllowed
	}

	if record.ResponseStatus != nil && *record.ResponseStatus != notification.ResponseStatusPending {
		readAt := derefTime(record.ReadAt)
		if readAt.IsZero() {
			readAt, err = r.upsertReadTx(ctx, tx, notificationID, userID)
			if err != nil {
				return NotificationRespondResult{}, err
			}
		}

		if err := tx.Commit(); err != nil {
			return NotificationRespondResult{}, err
		}

		return NotificationRespondResult{
			ResponseStatus: *record.ResponseStatus,
			RespondedAt:    derefTime(record.RespondedAt),
			ReadAt:         readAt,
		}, nil
	}

	effectiveStatus := requestedStatus
	if forceCancel {
		effectiveStatus = notification.ResponseStatusCanceled
	}

	readAt, err := r.upsertReadTx(ctx, tx, notificationID, userID)
	if err != nil {
		return NotificationRespondResult{}, err
	}

	var respondedAt time.Time
	if err := tx.QueryRowContext(ctx, `
		update notification_targets
		set response_status = $3,
		    responded_at = now()
		where notification_id = $1
		  and user_id = $2
		returning responded_at
	`, notificationID, userID, effectiveStatus).Scan(&respondedAt); err != nil {
		return NotificationRespondResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return NotificationRespondResult{}, err
	}

	return NotificationRespondResult{
		ResponseStatus: effectiveStatus,
		RespondedAt:    respondedAt,
		ReadAt:         readAt,
	}, nil
}

func (r *NotificationRepository) canAccess(ctx context.Context, userID, notificationID int64) bool {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		select exists (
			select 1
			from notifications n
			left join notification_targets nt
				on nt.notification_id = n.id
			   and nt.user_id = $2
			where n.id = $1
			  and n.status = $3
			  and (n.publish_at is null or n.publish_at <= now())
			  and (n.expires_at is null or n.expires_at > now())
			  and (
				n.audience = $4
				or (n.audience = $5 and nt.user_id is not null)
			  )
		)
	`, notificationID, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}

func (r *NotificationRepository) findResponseTargetForUpdate(ctx context.Context, tx *sql.Tx, userID, notificationID int64) (NotificationRecord, error) {
	var record NotificationRecord
	var imagePath sql.NullString
	var responseStatus sql.NullInt64

	err := tx.QueryRowContext(ctx, `
		select
			n.id,
			n.title,
			coalesce(n.body, ''),
			n.image_path,
			n.status,
			n.audience,
			n.publish_at,
			n.expires_at,
			n.created_at,
			nr.read_at,
			nt.response_status,
			nt.responded_at
		from notifications n
		join notification_targets nt
			on nt.notification_id = n.id
		   and nt.user_id = $2
		left join notification_reads nr
			on nr.notification_id = n.id
		   and nr.user_id = $2
		where n.id = $1
		  and n.status = $3
		  and (n.publish_at is null or n.publish_at <= now())
		  and (n.expires_at is null or n.expires_at > now())
		for update of nt
	`, notificationID, userID, notificationStatusPublished).Scan(
		&record.ID,
		&record.Title,
		&record.Body,
		&imagePath,
		&record.Status,
		&record.Audience,
		&record.PublishAt,
		&record.ExpiresAt,
		&record.CreatedAt,
		&record.ReadAt,
		&responseStatus,
		&record.RespondedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return NotificationRecord{}, ErrNotificationNotFound
		}

		return NotificationRecord{}, err
	}

	if imagePath.Valid {
		record.ImagePath = &imagePath.String
	}
	if responseStatus.Valid {
		value := int(responseStatus.Int64)
		record.ResponseStatus = &value
	}

	return record, nil
}

func (r *NotificationRepository) upsertReadTx(ctx context.Context, tx *sql.Tx, notificationID, userID int64) (time.Time, error) {
	var readAt time.Time
	if err := tx.QueryRowContext(ctx, `
		insert into notification_reads (notification_id, user_id, read_at)
		values ($1, $2, now())
		on conflict (notification_id, user_id)
		do update set read_at = excluded.read_at
		returning read_at
	`, notificationID, userID).Scan(&readAt); err != nil {
		return time.Time{}, err
	}

	return readAt, nil
}

func derefTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}

	return *value
}
