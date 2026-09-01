package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gin/internal/domain/telegram"
)

type TelegramRepository struct {
	db *sql.DB
}

func NewTelegramRepository(db *sql.DB) *TelegramRepository {
	return &TelegramRepository{db: db}
}

func (r *TelegramRepository) UpsertGroupEvent(ctx context.Context, event telegram.GroupEvent) error {
	status := strings.ToLower(strings.TrimSpace(event.BotStatus))
	if event.SiteCode == "" || event.ChatID == 0 || event.ChatType == "" || status == "" {
		return sql.ErrNoRows
	}
	when := event.OccurredAt
	if when.IsZero() {
		when = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
		insert into telegram_chat_destinations (
			site_code, telegram_chat_id, chat_type, title, username, bot_status,
			is_active, discovered_at, last_seen_at, removed_at, created_at, updated_at
		) values ($1, $2, $3, $4, nullif($5, ''), $6,
			false, $7, $7, case when $6 in ('left', 'kicked') then $7 else null end, $7, $7)
		on conflict (site_code, telegram_chat_id) do update set
			chat_type = excluded.chat_type,
			title = excluded.title,
			username = excluded.username,
			bot_status = excluded.bot_status,
			last_seen_at = excluded.last_seen_at,
			removed_at = case when excluded.bot_status in ('left', 'kicked') then excluded.last_seen_at else null end,
			last_error = case when excluded.bot_status in ('left', 'kicked') then telegram_chat_destinations.last_error else null end,
			last_error_at = case when excluded.bot_status in ('left', 'kicked') then telegram_chat_destinations.last_error_at else null end,
			is_active = case when excluded.bot_status in ('left', 'kicked') then false else telegram_chat_destinations.is_active end,
			updated_at = excluded.updated_at
	`, event.SiteCode, event.ChatID, strings.ToLower(strings.TrimSpace(event.ChatType)), event.Title, event.Username, status, when)
	return err
}

func (r *TelegramRepository) ListActiveTargets(ctx context.Context, siteCode string) ([]telegram.Target, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, telegram_chat_id, chat_type, title, coalesce(username, '')
		from telegram_chat_destinations
		where site_code = $1 and is_active = true and bot_status in ('member', 'administrator')
		order by id
	`, strings.TrimSpace(siteCode))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := make([]telegram.Target, 0)
	for rows.Next() {
		var target telegram.Target
		if err := rows.Scan(&target.ID, &target.ChatID, &target.ChatType, &target.Title, &target.Username); err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, rows.Err()
}

func (r *TelegramRepository) MarkTargetError(ctx context.Context, request telegram.TargetError) error {
	_, err := r.db.ExecContext(ctx, `
		update telegram_chat_destinations
		set is_active = false,
		    bot_status = 'error',
		    last_error = nullif($1, ''),
		    last_error_at = now(),
		    deactivated_at = now(),
		    updated_at = now()
		where site_code = $2 and telegram_chat_id = $3
	`, request.Error, strings.TrimSpace(request.SiteCode), request.ChatID)
	return err
}
