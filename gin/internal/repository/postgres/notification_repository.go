package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gin/internal/support/message"
)

const (
	notificationStatusPublished = 2
	notificationAudienceAll     = 1
	notificationAudienceUsers   = 2
)

var (
	ErrNotificationNotFound = errors.New(message.NotificationNotFound)
)

type NotificationRepository struct {
	db *sql.DB
}

type NotificationRecord struct {
	ID        int64
	Title     string
	Body      string
	Status    int
	Audience  int
	PublishAt *time.Time
	ExpiresAt *time.Time
	CreatedAt time.Time
	ReadAt    *time.Time
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) ListForUser(ctx context.Context, userID int64, page, pageSize int) ([]NotificationRecord, int, error) {
	page, pageSize = normalizePagination(page, pageSize)

	var total int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from notifications n
		where n.status = $2
		  and (n.publish_at is null or n.publish_at <= now())
		  and (n.expires_at is null or n.expires_at > now())
		  and (
			n.audience = $3
			or (
				n.audience = $4
				and exists (
					select 1
					from notification_targets nt
					where nt.notification_id = n.id
					  and nt.user_id = $1
				)
			)
		  )
	`, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		select
			n.id,
			n.title,
			n.body,
			n.status,
			n.audience,
			n.publish_at,
			n.expires_at,
			n.created_at,
			nr.read_at
		from notifications n
		left join notification_reads nr
			on nr.notification_id = n.id
		   and nr.user_id = $1
		where n.status = $2
		  and (n.publish_at is null or n.publish_at <= now())
		  and (n.expires_at is null or n.expires_at > now())
		  and (
			n.audience = $3
			or (
				n.audience = $4
				and exists (
					select 1
					from notification_targets nt
					where nt.notification_id = n.id
					  and nt.user_id = $1
				)
			)
		  )
		order by coalesce(n.publish_at, n.created_at) desc, n.id desc
		limit $5 offset $6
	`, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]NotificationRecord, 0)
	for rows.Next() {
		var record NotificationRecord
		if err := rows.Scan(
			&record.ID,
			&record.Title,
			&record.Body,
			&record.Status,
			&record.Audience,
			&record.PublishAt,
			&record.ExpiresAt,
			&record.CreatedAt,
			&record.ReadAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, rows.Err()
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

func (r *NotificationRepository) canAccess(ctx context.Context, userID, notificationID int64) bool {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		select exists (
			select 1
			from notifications n
			where n.id = $1
			  and n.status = $3
			  and (n.publish_at is null or n.publish_at <= now())
			  and (n.expires_at is null or n.expires_at > now())
			  and (
				n.audience = $4
				or (
					n.audience = $5
					and exists (
						select 1
						from notification_targets nt
						where nt.notification_id = n.id
						  and nt.user_id = $2
					)
				)
			  )
		)
	`, notificationID, userID, notificationStatusPublished, notificationAudienceAll, notificationAudienceUsers).Scan(&exists)
	if err != nil {
		return false
	}

	return exists
}
