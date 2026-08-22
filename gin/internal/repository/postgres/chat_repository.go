package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gin/internal/domain/chat"
)

var ErrChatRoomNotFound = errors.New("chat.room_not_found")

type ChatRepository struct{ db *sql.DB }

func NewChatRepository(db *sql.DB) *ChatRepository { return &ChatRepository{db: db} }

func (r *ChatRepository) RoomEnabled(ctx context.Context, code string) (bool, error) {
	var enabled bool
	if err := r.db.QueryRowContext(ctx, `select enabled from chat_rooms where code = $1 limit 1`, code).Scan(&enabled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrChatRoomNotFound
		}
		return false, err
	}
	return enabled, nil
}

func (r *ChatRepository) ListMessages(ctx context.Context, roomCode string, beforeID int64, limit int) ([]chat.Message, error) {
	if limit < 1 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	query := `
		select m.id, m.display_name, m.body, m.actor_type, m.created_at
		from chat_messages m
		join chat_rooms r on r.id = m.room_id
		where r.code = $1 and m.status = 1`
	args := []any{roomCode}
	if beforeID > 0 {
		query += " and m.id < $2"
		args = append(args, beforeID)
	}
	query += fmt.Sprintf(" order by m.id desc limit $%d", len(args)+1)
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]chat.Message, 0, limit)
	for rows.Next() {
		var item chat.Message
		if err := rows.Scan(&item.ID, &item.DisplayName, &item.Body, &item.ActorType, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return items, nil
}

func (r *ChatRepository) ActiveBan(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		select exists(
			select 1 from chat_bans
			where user_id = $1 and revoked_at is null
			  and (expires_at is null or expires_at > now())
		)`, userID).Scan(&exists)
	return exists, err
}

func (r *ChatRepository) DisplayName(ctx context.Context, userID int64) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx, `select display_name from chat_user_profiles where user_id = $1`, userID).Scan(&name)
	if err == nil {
		return name, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	for attempts := 0; attempts < 3; attempts++ {
		err = r.db.QueryRowContext(ctx, `
		insert into chat_user_profiles (user_id, display_name, created_at, updated_at)
		values ($1, concat('Nguoi choi ', substr(md5(random()::text || clock_timestamp()::text), 1, 8)), now(), now())
		on conflict (user_id) do update set updated_at = now()
		returning display_name
		`, userID).Scan(&name)
		if err == nil {
			return name, nil
		}
	}
	return "", err
}

func (r *ChatRepository) InsertMessage(ctx context.Context, roomCode, actorType string, userID, botID int64, displayName, body string) (chat.Message, error) {
	var item chat.Message
	var userArg, botArg any
	if userID > 0 {
		userArg = userID
	}
	if botID > 0 {
		botArg = botID
	}
	createdAtUTC := time.Now().UTC()
	err := r.db.QueryRowContext(ctx, `
		insert into chat_messages (room_id, actor_type, user_id, bot_profile_id, display_name, body, status, created_at, updated_at)
		select r.id, $2, $3, $4, $5, $6, 1, $7, $7
		from chat_rooms r where r.code = $1
		returning id, display_name, body, actor_type, created_at
	`, roomCode, actorType, userArg, botArg, displayName, body, createdAtUTC).Scan(&item.ID, &item.DisplayName, &item.Body, &item.ActorType, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return chat.Message{}, ErrChatRoomNotFound
	}
	if err != nil {
		return chat.Message{}, err
	}
	return item, nil
}

func (r *ChatRepository) DeleteExpired(ctx context.Context, messageHours, auditDays int) error {
	if messageHours < 1 {
		messageHours = 6
	}
	if auditDays < 1 {
		auditDays = 90
	}
	if _, err := r.db.ExecContext(ctx, `delete from chat_messages where created_at < timezone('UTC', now()) - ($1 * interval '1 hour')`, messageHours); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `delete from chat_moderation_actions where created_at < now() - ($1 * interval '1 day')`, auditDays)
	return err
}
