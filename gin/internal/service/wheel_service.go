package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gin/internal/domain/wheel"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"

	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrWheelDisabled      = errors.New("wheel.disabled")
	ErrWheelTokenInvalid  = errors.New("wheel.token.invalid")
	ErrWheelChatInvalid   = errors.New("wheel.chat.invalid")
	ErrWheelChatRateLimit = errors.New("wheel.chat.rate_limited")
)

type WheelConfig struct {
	Enabled             bool
	SiteCode            string
	MicrositeURL        string
	DurationSeconds     int
	SpinDurationSeconds int
}

type WheelService struct {
	repository *repopg.WheelRepository
	wallet     *WalletService
	broker     *realtime.Broker
	redis      *goredis.Client
	config     WheelConfig
}

type launchRecord struct {
	UserID       int64 `json:"user_id"`
	InvitationID int64 `json:"invitation_id"`
}

// Redis GETDEL was introduced in Redis 6.2, while the production baseline on
// Ubuntu 22.04 is Redis 6.0. Lua keeps the read-and-delete operation atomic on
// both versions, which is required for one-time launch and websocket tickets.
var wheelGetDelScript = goredis.NewScript(`
local value = redis.call("GET", KEYS[1])
if value then
  redis.call("DEL", KEYS[1])
end
return value
`)

func NewWheelService(repository *repopg.WheelRepository, wallet *WalletService, broker *realtime.Broker, redis *goredis.Client, config WheelConfig) *WheelService {
	return &WheelService{repository: repository, wallet: wallet, broker: broker, redis: redis, config: config}
}

func (s *WheelService) Enabled() bool { return s != nil && s.config.Enabled }

func (s *WheelService) ListInvitations(ctx context.Context, userID int64) (wheel.InvitationListResponse, error) {
	if !s.Enabled() {
		return wheel.InvitationListResponse{}, ErrWheelDisabled
	}
	items, err := s.repository.ListInvitations(ctx, userID)
	return wheel.InvitationListResponse{Items: items}, err
}

func (s *WheelService) MarkSeen(ctx context.Context, publicID string, userID int64) error {
	if !s.Enabled() {
		return ErrWheelDisabled
	}
	return s.repository.MarkInvitationSeen(ctx, publicID, userID)
}

func (s *WheelService) Launch(ctx context.Context, publicID string, userID int64) (wheel.LaunchResponse, error) {
	if !s.Enabled() {
		return wheel.LaunchResponse{}, ErrWheelDisabled
	}
	record, err := s.repository.FindInvitationForUser(ctx, publicID, userID)
	if err != nil {
		return wheel.LaunchResponse{}, err
	}
	if !wheelInvitationCanOpen(record) {
		return wheel.LaunchResponse{}, repopg.ErrWheelInvitationInactive
	}
	code, err := randomOpaqueToken(32)
	if err != nil {
		return wheel.LaunchResponse{}, err
	}
	raw, _ := json.Marshal(launchRecord{UserID: userID, InvitationID: record.ID})
	if err := s.redis.Set(ctx, "wheel:launch:"+hashOpaqueToken(code), raw, time.Minute).Err(); err != nil {
		return wheel.LaunchResponse{}, err
	}
	destination, err := url.Parse(strings.TrimSpace(s.config.MicrositeURL))
	if err != nil {
		return wheel.LaunchResponse{}, err
	}
	query := destination.Query()
	query.Set("launch_code", code)
	destination.RawQuery = query.Encode()
	return wheel.LaunchResponse{URL: destination.String(), ExpiresIn: 60}, nil
}

func (s *WheelService) Exchange(ctx context.Context, code string) (wheel.ExchangeResponse, error) {
	if !s.Enabled() {
		return wheel.ExchangeResponse{}, ErrWheelDisabled
	}
	raw, err := s.getAndDelete(ctx, "wheel:launch:"+hashOpaqueToken(strings.TrimSpace(code)))
	if errors.Is(err, goredis.Nil) {
		return wheel.ExchangeResponse{}, ErrWheelTokenInvalid
	}
	if err != nil {
		return wheel.ExchangeResponse{}, err
	}
	var launch launchRecord
	if json.Unmarshal(raw, &launch) != nil || launch.UserID == 0 || launch.InvitationID == 0 {
		return wheel.ExchangeResponse{}, ErrWheelTokenInvalid
	}
	record, err := s.repository.FindInvitationByID(ctx, launch.InvitationID, launch.UserID)
	if err != nil || !wheelInvitationCanOpen(record) {
		return wheel.ExchangeResponse{}, ErrWheelTokenInvalid
	}
	token, err := randomOpaqueToken(40)
	if err != nil {
		return wheel.ExchangeResponse{}, err
	}
	access := wheel.Access{UserID: launch.UserID, InvitationID: launch.InvitationID}
	accessRaw, _ := json.Marshal(access)
	const ttl = 15 * time.Minute
	if err := s.redis.Set(ctx, "wheel:access:"+hashOpaqueToken(token), accessRaw, ttl).Err(); err != nil {
		return wheel.ExchangeResponse{}, err
	}
	return wheel.ExchangeResponse{AccessToken: token, ExpiresIn: int64(ttl.Seconds()), Invitation: wheelInvitationDTO(record)}, nil
}

func (s *WheelService) Authenticate(ctx context.Context, token string) (wheel.Access, error) {
	if !s.Enabled() {
		return wheel.Access{}, ErrWheelDisabled
	}
	raw, err := s.redis.Get(ctx, "wheel:access:"+hashOpaqueToken(strings.TrimSpace(token))).Bytes()
	if errors.Is(err, goredis.Nil) {
		return wheel.Access{}, ErrWheelTokenInvalid
	}
	if err != nil {
		return wheel.Access{}, err
	}
	var access wheel.Access
	if json.Unmarshal(raw, &access) != nil || access.UserID == 0 || access.InvitationID == 0 {
		return wheel.Access{}, ErrWheelTokenInvalid
	}
	return access, nil
}

func (s *WheelService) CreateUserSocketTicket(ctx context.Context, userID int64) (wheel.SocketTicketResponse, error) {
	if !s.Enabled() {
		return wheel.SocketTicketResponse{}, ErrWheelDisabled
	}
	ticket, err := randomOpaqueToken(28)
	if err != nil {
		return wheel.SocketTicketResponse{}, err
	}
	if err := s.redis.Set(ctx, "wheel:socket:user:"+hashOpaqueToken(ticket), userID, time.Minute).Err(); err != nil {
		return wheel.SocketTicketResponse{}, err
	}
	return wheel.SocketTicketResponse{Ticket: ticket, ExpiresIn: 60}, nil
}

func (s *WheelService) ConsumeUserSocketTicket(ctx context.Context, ticket string) (int64, error) {
	raw, err := s.getAndDelete(ctx, "wheel:socket:user:"+hashOpaqueToken(strings.TrimSpace(ticket)))
	if errors.Is(err, goredis.Nil) {
		return 0, ErrWheelTokenInvalid
	}
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(string(raw), 10, 64)
	if err != nil || value <= 0 {
		return 0, ErrWheelTokenInvalid
	}
	return value, nil
}

func (s *WheelService) CreateSessionSocketTicket(ctx context.Context, access wheel.Access) (wheel.SocketTicketResponse, error) {
	record, err := s.repository.FindInvitationByID(ctx, access.InvitationID, access.UserID)
	if err != nil || record.SessionID == nil {
		return wheel.SocketTicketResponse{}, repopg.ErrWheelSessionNotStarted
	}
	ticket, err := randomOpaqueToken(28)
	if err != nil {
		return wheel.SocketTicketResponse{}, err
	}
	raw, _ := json.Marshal(access)
	if err := s.redis.Set(ctx, "wheel:socket:session:"+hashOpaqueToken(ticket), raw, time.Minute).Err(); err != nil {
		return wheel.SocketTicketResponse{}, err
	}
	return wheel.SocketTicketResponse{Ticket: ticket, ExpiresIn: 60}, nil
}

func (s *WheelService) ConsumeSessionSocketTicket(ctx context.Context, ticket string) (wheel.Access, int64, error) {
	raw, err := s.getAndDelete(ctx, "wheel:socket:session:"+hashOpaqueToken(strings.TrimSpace(ticket)))
	if errors.Is(err, goredis.Nil) {
		return wheel.Access{}, 0, ErrWheelTokenInvalid
	}
	if err != nil {
		return wheel.Access{}, 0, err
	}
	var access wheel.Access
	if json.Unmarshal(raw, &access) != nil {
		return wheel.Access{}, 0, ErrWheelTokenInvalid
	}
	record, err := s.repository.FindInvitationByID(ctx, access.InvitationID, access.UserID)
	if err != nil || record.SessionID == nil {
		return wheel.Access{}, 0, ErrWheelTokenInvalid
	}
	return access, *record.SessionID, nil
}

func (s *WheelService) getAndDelete(ctx context.Context, key string) ([]byte, error) {
	value, err := wheelGetDelScript.Run(ctx, s.redis, []string{key}).Result()
	if err != nil {
		return nil, err
	}
	switch typed := value.(type) {
	case string:
		return []byte(typed), nil
	case []byte:
		return typed, nil
	case nil:
		return nil, goredis.Nil
	default:
		return nil, fmt.Errorf("wheel token returned unexpected redis type %T", value)
	}
}

func (s *WheelService) Start(ctx context.Context, access wheel.Access) (wheel.State, error) {
	result, err := s.repository.StartSession(ctx, access.InvitationID, access.UserID, s.config.DurationSeconds)
	if err != nil {
		return wheel.State{}, err
	}
	if len(result.OutboxIDs) > 0 {
		if publishErr := s.broker.Publish(ctx, realtime.WheelSessionTopic(result.SessionID), "wheel.session.started", result.State); publishErr == nil {
			_ = s.repository.MarkOutboxPublished(ctx, result.OutboxIDs)
		}
	}
	return result.State, nil
}

func (s *WheelService) State(ctx context.Context, access wheel.Access) (wheel.State, error) {
	return s.repository.State(ctx, access.InvitationID, access.UserID, s.config.SpinDurationSeconds)
}

func (s *WheelService) Spin(ctx context.Context, access wheel.Access, roundNo int) (wheel.SpinResponse, error) {
	result, err := s.repository.Spin(ctx, access.InvitationID, access.UserID, roundNo, s.config.SpinDurationSeconds, s.config.SiteCode)
	if err != nil {
		return wheel.SpinResponse{}, err
	}
	publishErr := s.broker.Publish(ctx, realtime.WheelSessionTopic(result.SessionID), "wheel.round.revealed", result.Round)
	if result.Round.PrizeAmount != nil && positiveWheelAmount(*result.Round.PrizeAmount) {
		if err := s.broker.Publish(ctx, realtime.WheelSessionTopic(result.SessionID), "wheel.reward.paid", result.State.PaidRewards); publishErr == nil {
			publishErr = err
		}
		_ = s.wallet.PublishSummary(ctx, access.UserID)
	}
	if result.State.SessionStatus == "completed" {
		if err := s.broker.Publish(ctx, realtime.WheelSessionTopic(result.SessionID), "wheel.session.completed", result.State); publishErr == nil {
			publishErr = err
		}
	}
	if publishErr == nil {
		_ = s.repository.MarkOutboxPublished(ctx, result.OutboxIDs)
	}
	return wheel.SpinResponse{State: result.State, Result: result.Round}, nil
}

func (s *WheelService) ListChat(ctx context.Context, access wheel.Access, before, limit int64) (wheel.ChatListResponse, error) {
	items, next, err := s.repository.ListChat(ctx, access.InvitationID, access.UserID, before, limit)
	return wheel.ChatListResponse{Items: items, NextCursor: next}, err
}

func (s *WheelService) CreateChat(ctx context.Context, access wheel.Access, ip, body string) (wheel.ChatCreateResponse, error) {
	body = strings.TrimSpace(body)
	if utf8.RuneCountInString(body) < 1 || utf8.RuneCountInString(body) > 280 || wheelURLPattern.MatchString(body) {
		return wheel.ChatCreateResponse{}, ErrWheelChatInvalid
	}
	allowed, err := s.allowChat(ctx, access.UserID, ip)
	if err != nil {
		return wheel.ChatCreateResponse{}, err
	}
	if !allowed {
		return wheel.ChatCreateResponse{}, ErrWheelChatRateLimit
	}
	item, sessionID, err := s.repository.CreateChat(ctx, access.InvitationID, access.UserID, body)
	if err != nil {
		return wheel.ChatCreateResponse{}, err
	}
	_ = s.broker.Publish(ctx, realtime.WheelSessionTopic(sessionID), "chat.message.created", item)
	return wheel.ChatCreateResponse{Message: item}, nil
}

func (s *WheelService) allowChat(ctx context.Context, userID int64, ip string) (bool, error) {
	cooldown := fmt.Sprintf("wheel:chat:cooldown:%d", userID)
	ok, err := s.redis.SetNX(ctx, cooldown, 1, 10*time.Second).Result()
	if err != nil || !ok {
		return ok, err
	}
	bucket := time.Now().UTC().Format("200601021504")
	keys := []string{fmt.Sprintf("wheel:chat:minute:user:%d:%s", userID, bucket), fmt.Sprintf("wheel:chat:minute:ip:%s:%s", hashOpaqueToken(ip), bucket)}
	for _, key := range keys {
		count, err := s.redis.Incr(ctx, key).Result()
		if err != nil {
			return false, err
		}
		if count == 1 {
			_ = s.redis.Expire(ctx, key, 2*time.Minute).Err()
		}
		if count > 20 {
			return false, nil
		}
	}
	return true, nil
}

func wheelInvitationCanOpen(record repopg.WheelInvitationAccess) bool {
	if !map[string]bool{"pending": true, "started": true, "completed": true}[record.Status] {
		return false
	}
	return record.Status != "pending" || record.ExpiresAt == nil || time.Now().Before(*record.ExpiresAt)
}

func wheelInvitationDTO(record repopg.WheelInvitationAccess) wheel.Invitation {
	item := wheel.Invitation{ID: record.PublicID, CampaignName: record.CampaignName, Status: record.Status}
	if record.ExpiresAt != nil {
		value := record.ExpiresAt.UTC().Format(time.RFC3339Nano)
		item.ExpiresAt = &value
	}
	if record.PopupSeenAt != nil {
		value := record.PopupSeenAt.UTC().Format(time.RFC3339Nano)
		item.SeenAt = &value
	}
	item.SessionID, item.SessionStatus = record.SessionPublicID, record.SessionStatus
	return item
}

func randomOpaqueToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func hashOpaqueToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func positiveWheelAmount(value string) bool {
	amount, ok := new(big.Rat).SetString(strings.TrimSpace(value))
	return ok && amount.Sign() > 0
}

var wheelURLPattern = regexp.MustCompile(`(?i)(https?://|www\.|[a-z0-9-]+\.(com|net|org|vn|io|me|xyz|site|win)\b)`)
