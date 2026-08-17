package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/chat"
	"gin/internal/realtime"
	"gin/internal/repository/postgres"
	"gin/internal/security/ratelimit"

	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrChatDisabled    = errors.New("chat.disabled")
	ErrChatBanned      = errors.New("chat.banned")
	ErrChatRateLimited = errors.New("chat.rate_limited")
	ErrChatInvalid     = errors.New("chat.invalid_message")
)

var chatURLPattern = regexp.MustCompile(`(?i)(https?://|www\.|(?:^|[\s(])(?:[a-z0-9-]+\.)+[a-z]{2,}(?:[/?#:)]|\s|$))`)

type ChatService struct {
	repo     *postgres.ChatRepository
	broker   *realtime.Broker
	limiter  *ratelimit.Limiter
	redis    *goredis.Client
	enabled  bool
	roomCode string
}

func NewChatService(repo *postgres.ChatRepository, broker *realtime.Broker, limiter *ratelimit.Limiter, redis *goredis.Client, enabled bool, roomCode string) *ChatService {
	if strings.TrimSpace(roomCode) == "" {
		roomCode = "global"
	}
	return &ChatService{repo: repo, broker: broker, limiter: limiter, redis: redis, enabled: enabled, roomCode: roomCode}
}

func (s *ChatService) enabledRoom(ctx context.Context) error {
	if !s.enabled {
		return ErrChatDisabled
	}
	enabled, err := s.repo.RoomEnabled(ctx, s.roomCode)
	if err != nil {
		return err
	}
	if !enabled {
		return ErrChatDisabled
	}
	return nil
}

func (s *ChatService) List(ctx context.Context, beforeID, limit int64) (chat.ListResponse, error) {
	if err := s.enabledRoom(ctx); err != nil {
		return chat.ListResponse{}, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	items, err := s.repo.ListMessages(ctx, s.roomCode, beforeID, int(limit))
	if err != nil {
		return chat.ListResponse{}, err
	}
	response := chat.ListResponse{Items: items}
	if len(items) == int(limit) && len(items) > 0 {
		cursor := items[0].ID
		response.NextCursor = &cursor
	}
	return response, nil
}

func (s *ChatService) Create(ctx context.Context, userID int64, ip, body string) (chat.Message, error) {
	if err := s.enabledRoom(ctx); err != nil {
		return chat.Message{}, err
	}
	ban, err := s.repo.ActiveBan(ctx, userID)
	if err != nil {
		return chat.Message{}, err
	}
	if ban {
		return chat.Message{}, ErrChatBanned
	}
	body = normalizeChatBody(body)
	if !validChatBody(body) {
		return chat.Message{}, ErrChatInvalid
	}
	if s.limiter != nil {
		ok, _, err := s.limiter.StartCooldown(ctx, "chat:cooldown:"+s.roomCode+":user:"+itoa(userID), 10*time.Second)
		if err != nil {
			return chat.Message{}, err
		}
		if !ok {
			return chat.Message{}, ErrChatRateLimited
		}
		window, err := s.limiter.HitWindow(ctx, "chat:window:"+s.roomCode+":user:"+itoa(userID), 20, time.Minute)
		if err != nil {
			return chat.Message{}, err
		}
		if !window.Allowed {
			return chat.Message{}, ErrChatRateLimited
		}
		ipWindow, err := s.limiter.HitWindow(ctx, "chat:window:"+s.roomCode+":ip:"+ip, 20, time.Minute)
		if err != nil {
			return chat.Message{}, err
		}
		if !ipWindow.Allowed {
			return chat.Message{}, ErrChatRateLimited
		}
	}
	if s.redis != nil {
		hash := sha256.Sum256([]byte(body))
		key := "chat:duplicate:" + s.roomCode + ":user:" + itoa(userID) + ":" + hex.EncodeToString(hash[:])
		duplicate, err := s.redis.SetNX(ctx, key, "1", time.Minute).Result()
		if err != nil {
			return chat.Message{}, err
		}
		if !duplicate {
			return chat.Message{}, ErrChatInvalid
		}
	}
	displayName, err := s.repo.DisplayName(ctx, userID)
	if err != nil {
		return chat.Message{}, err
	}
	item, err := s.repo.InsertMessage(ctx, s.roomCode, "user", userID, 0, displayName, body)
	if err != nil {
		return chat.Message{}, err
	}
	if s.broker != nil {
		if err := s.broker.Publish(ctx, realtime.ChatGlobalTopic(s.roomCode), "chat.message.created", item); err != nil { /* REST remains source of truth. */
		}
	}
	return item, nil
}

func (s *ChatService) RoomCode() string { return s.roomCode }

func normalizeChatBody(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}
func validChatBody(value string) bool {
	value = normalizeChatBody(value)
	return value != "" && len([]rune(value)) <= 280 && !chatURLPattern.MatchString(value)
}
func itoa(value int64) string { return strconv.FormatInt(value, 10) }
