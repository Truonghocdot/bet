package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gate/internal/domain/event"
	ginclient "gate/internal/integration/gin"
	telegramclient "gate/internal/integration/telegram"
	"github.com/redis/go-redis/v9"
)

type TelegramNotifierConfig struct {
	Enabled       bool
	SiteCode      string
	WebhookSecret string
	BotToken      string
	Stream        string
	Group         string
	Consumer      string
	MaxRetries    int
}

type TelegramNotifier struct {
	gin           *ginclient.Client
	redis         *redis.Client
	bot           *telegramclient.Client
	botTokenValue string
	enabled       bool
	siteCode      string
	webhookSecret string
	stream        string
	group         string
	consumer      string
	maxRetries    int
}

type depositNotificationEvent struct {
	EventKey  string                          `json:"event_key"`
	SiteCode  string                          `json:"site_code"`
	ClientRef string                          `json:"client_ref,omitempty"`
	Amount    string                          `json:"amount,omitempty"`
	Content   string                          `json:"content,omitempty"`
	PaidAt    time.Time                       `json:"paid_at"`
	Lookup    event.DepositNotificationLookup `json:"lookup"`
}

type telegramUpdate struct {
	MyChatMember *telegramChatMemberUpdated `json:"my_chat_member"`
}

type telegramChatMemberUpdated struct {
	Chat struct {
		ID       int64  `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title"`
		Username string `json:"username"`
	} `json:"chat"`
	NewChatMember struct {
		Status string `json:"status"`
	} `json:"new_chat_member"`
}

func NewTelegramNotifier(gin *ginclient.Client, sharedRedis *redis.Client, config TelegramNotifierConfig) *TelegramNotifier {
	siteCode := strings.TrimSpace(config.SiteCode)
	stream := strings.TrimSpace(config.Stream)
	if stream == "" {
		stream = "telegram:deposit:events:" + siteCode
	}
	group := strings.TrimSpace(config.Group)
	if group == "" {
		group = "telegram-notifier-" + siteCode
	}
	consumer := strings.TrimSpace(config.Consumer)
	if consumer == "" {
		consumer = "gate-" + siteCode
		if hostname, err := os.Hostname(); err == nil && strings.TrimSpace(hostname) != "" {
			consumer += "-" + strings.TrimSpace(hostname)
		}
	}
	maxRetries := config.MaxRetries
	if maxRetries < 1 {
		maxRetries = 3
	}

	return &TelegramNotifier{
		gin:           gin,
		redis:         sharedRedis,
		bot:           telegramclient.NewClient(config.BotToken),
		botTokenValue: strings.TrimSpace(config.BotToken),
		enabled:       config.Enabled,
		siteCode:      siteCode,
		webhookSecret: strings.TrimSpace(config.WebhookSecret),
		stream:        stream,
		group:         group,
		consumer:      consumer,
		maxRetries:    maxRetries,
	}
}

func (n *TelegramNotifier) WebhookSecret() string {
	return n.webhookSecret
}

func (n *TelegramNotifier) Enabled() bool {
	return n.enabled
}

func (n *TelegramNotifier) SiteCode() string {
	return n.siteCode
}

func (n *TelegramNotifier) SendTest(ctx context.Context, chatID int64, message string) error {
	if !n.enabled {
		return fmt.Errorf("telegram notifier is disabled")
	}
	if chatID == 0 {
		return fmt.Errorf("telegram chat id is required")
	}
	return n.bot.SendMessage(ctx, chatID, sanitizeLine(message, 4000))
}

func (n *TelegramNotifier) HandleWebhookUpdate(ctx context.Context, raw []byte) error {
	if n.gin == nil {
		return fmt.Errorf("gin client is unavailable")
	}
	var update telegramUpdate
	if err := json.Unmarshal(raw, &update); err != nil {
		return fmt.Errorf("invalid telegram update: %w", err)
	}
	if update.MyChatMember == nil {
		return nil
	}

	chat := update.MyChatMember.Chat
	chatType := strings.ToLower(strings.TrimSpace(chat.Type))
	if chat.ID == 0 || (chatType != "group" && chatType != "supergroup") {
		return nil
	}

	status := strings.ToLower(strings.TrimSpace(update.MyChatMember.NewChatMember.Status))
	if status != "member" && status != "administrator" && status != "left" && status != "kicked" {
		return nil
	}

	return n.gin.RecordTelegramGroupEvent(ctx, event.TelegramGroupEvent{
		SiteCode:   n.siteCode,
		ChatID:     chat.ID,
		ChatType:   chatType,
		Title:      sanitizeLine(chat.Title, 160),
		Username:   sanitizeLine(chat.Username, 80),
		BotStatus:  status,
		OccurredAt: time.Now().UTC(),
	})
}

func (n *TelegramNotifier) EnqueueDeposit(ctx context.Context, request event.DepositApplyRequest) error {
	transferType := rawString(request.Raw, "transferType")
	if transferType == "" {
		transferType = rawString(request.Raw, "transfer_type")
	}
	if strings.ToLower(strings.TrimSpace(transferType)) != "in" {
		return nil
	}
	if !n.enabled {
		return nil
	}
	if n.redis == nil {
		return fmt.Errorf("telegram shared redis is unavailable")
	}
	if n.gin == nil {
		return fmt.Errorf("gin client is unavailable")
	}

	lookup, err := n.gin.LookupDepositForNotification(ctx, event.DepositNotificationLookupRequest{
		Provider:      "sepay_vietqr",
		ProviderTxnID: request.ProviderTxnID,
		ClientRef:     request.ClientRef,
	})
	if err != nil {
		return err
	}
	// The provider transaction ID is used only for the internal lookup. Keep it
	// out of the queued Telegram payload and the message shown to operators.
	lookup.ProviderTxnID = ""

	eventKey := n.eventKey(request)
	notification := depositNotificationEvent{
		EventKey:  eventKey,
		SiteCode:  n.siteCode,
		ClientRef: request.ClientRef,
		Amount:    request.Amount,
		Content:   sanitizeLine(rawString(request.Raw, "content"), 240),
		PaidAt:    request.PaidAt,
		Lookup:    lookup,
	}
	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	dedupeKey := "telegram:deposit:dedupe:" + n.siteCode + ":" + eventKey
	created, err := n.redis.SetNX(ctx, dedupeKey, 1, 24*time.Hour).Result()
	if err != nil {
		return err
	}
	if !created {
		return nil
	}

	if _, err := n.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: n.stream,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{"payload": string(payload)},
	}).Result(); err != nil {
		_ = n.redis.Del(ctx, dedupeKey).Err()
		return err
	}

	log.Printf("[telegram][queue] site=%s event=%s matched=%t", n.siteCode, eventKey, lookup.Matched)
	return nil
}

func (n *TelegramNotifier) Start(ctx context.Context) {
	if !n.enabled || n.redis == nil || strings.TrimSpace(n.botToken()) == "" {
		if n.enabled {
			log.Printf("[telegram][worker.disabled] site=%s reason=missing_redis_or_bot_token", n.siteCode)
		}
		return
	}
	go n.run(ctx)
}

func (n *TelegramNotifier) botToken() string {
	return n.botTokenValue
}

func (n *TelegramNotifier) run(ctx context.Context) {
	if err := n.ensureGroup(ctx); err != nil {
		log.Printf("[telegram][worker.error] site=%s stage=create_group err=%v", n.siteCode, err)
		return
	}

	for {
		if ctx.Err() != nil {
			return
		}
		if pending, _, err := n.redis.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   n.stream,
			Group:    n.group,
			Consumer: n.consumer,
			MinIdle:  30 * time.Second,
			Start:    "0-0",
			Count:    10,
		}).Result(); err == nil {
			for _, message := range pending {
				n.processMessage(ctx, message)
			}
		}
		streams, err := n.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    n.group,
			Consumer: n.consumer,
			Streams:  []string{n.stream, ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()
		if err != nil {
			if err == redis.Nil || ctx.Err() != nil {
				continue
			}
			log.Printf("[telegram][worker.error] site=%s stage=read err=%v", n.siteCode, err)
			time.Sleep(time.Second)
			continue
		}
		for _, stream := range streams {
			for _, message := range stream.Messages {
				n.processMessage(ctx, message)
			}
		}
	}
}

func (n *TelegramNotifier) ensureGroup(ctx context.Context) error {
	err := n.redis.XGroupCreateMkStream(ctx, n.stream, n.group, "0").Err()
	if err != nil && !strings.Contains(strings.ToUpper(err.Error()), "BUSYGROUP") {
		return err
	}
	return nil
}

func (n *TelegramNotifier) processMessage(ctx context.Context, message redis.XMessage) {
	defer func() { _ = n.redis.XAck(ctx, n.stream, n.group, message.ID).Err() }()
	payload, ok := message.Values["payload"].(string)
	if !ok {
		log.Printf("[telegram][dead-letter] site=%s stream_id=%s reason=missing_payload", n.siteCode, message.ID)
		return
	}
	var notification depositNotificationEvent
	if err := json.Unmarshal([]byte(payload), &notification); err != nil {
		log.Printf("[telegram][dead-letter] site=%s stream_id=%s reason=invalid_payload err=%v", n.siteCode, message.ID, err)
		return
	}

	var err error
	for attempt := 0; attempt < n.maxRetries; attempt++ {
		err = n.deliver(ctx, notification)
		if err == nil {
			log.Printf("[telegram][sent] site=%s event=%s", n.siteCode, notification.EventKey)
			return
		}
		if attempt+1 < n.maxRetries {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
	}

	_, _ = n.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: n.stream + ":dead-letter",
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{"payload": payload, "error": err.Error()},
	}).Result()
	log.Printf("[telegram][dead-letter] site=%s event=%s err=%v", n.siteCode, notification.EventKey, err)
}

func (n *TelegramNotifier) deliver(ctx context.Context, notification depositNotificationEvent) error {
	targets, err := n.gin.ListTelegramTargets(ctx, n.siteCode)
	if err != nil {
		return err
	}
	for _, target := range targets {
		deliveryKey := "telegram:deposit:sent:" + n.siteCode + ":" + notification.EventKey + ":" + strconv.FormatInt(target.ChatID, 10)
		sent, err := n.redis.Exists(ctx, deliveryKey).Result()
		if err != nil {
			return err
		}
		if sent > 0 {
			continue
		}
		if err := n.sendToTarget(ctx, target, formatDepositMessage(notification)); err != nil {
			return err
		}
		if err := n.redis.Set(ctx, deliveryKey, 1, 30*24*time.Hour).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (n *TelegramNotifier) sendToTarget(ctx context.Context, target event.TelegramTarget, message string) error {
	for attempt := 0; attempt < n.maxRetries; attempt++ {
		err := n.bot.SendMessage(ctx, target.ChatID, message)
		if err == nil {
			return nil
		}
		var apiErr *telegramclient.APIError
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode == 401 {
				log.Printf("[telegram][config-error] site=%s reason=unauthorized_bot_token", n.siteCode)
				return err
			}
			if apiErr.StatusCode == 403 {
				_ = n.gin.MarkTelegramTargetError(ctx, event.TelegramTargetError{SiteCode: n.siteCode, ChatID: target.ChatID, Error: apiErr.Error()})
				return err
			}
			if apiErr.RetryAfter > 0 {
				time.Sleep(apiErr.RetryAfter)
			} else if attempt+1 < n.maxRetries {
				time.Sleep(time.Duration(1<<attempt) * time.Second)
			}
		} else if attempt+1 < n.maxRetries {
			time.Sleep(time.Duration(1<<attempt) * time.Second)
		}
	}
	return fmt.Errorf("telegram delivery failed for chat %d", target.ChatID)
}

func (n *TelegramNotifier) eventKey(request event.DepositApplyRequest) string {
	value := strings.TrimSpace(request.ProviderTxnID)
	if value == "" {
		value = strings.TrimSpace(request.ClientRef)
	}
	if value == "" {
		raw, _ := json.Marshal(request.Raw)
		sum := sha256.Sum256(raw)
		value = hex.EncodeToString(sum[:])
	}
	return value
}

func formatDepositMessage(notification depositNotificationEvent) string {
	lookup := notification.Lookup
	amount := formatVND(notification.Amount)
	if amount == "0" && lookup.Amount != "" {
		amount = formatVND(lookup.Amount)
	}
	when := notification.PaidAt.Local().Format("02/01/2006 15:04:05")
	if !lookup.Matched {
		return strings.Join([]string{
			"[CẢNH BÁO TIỀN VÀO CHƯA KHỚP]",
			"Số tiền: " + amount + " VND",
			"Thời gian: " + when,
			"Nội dung CK: " + firstNonEmpty(notification.Content, "—"),
			"Lý do: Không tìm thấy lệnh nạp tương ứng",
		}, "\n")
	}

	status := "CHỜ DUYỆT THỦ CÔNG"
	if lookup.Status == 3 {
		status = "ĐÃ ĐƯỢC DUYỆT"
	}
	user := fmt.Sprintf("#%d - %s - %s", lookup.UserID, firstNonEmpty(lookup.UserName, "—"), firstNonEmpty(lookup.UserPhone, "—"))
	accountParts := []string{lookup.ReceivingBank, lookup.ReceivingAccountName, lookup.ReceivingAccount}
	accountValues := make([]string, 0, len(accountParts))
	for _, value := range accountParts {
		if strings.TrimSpace(value) != "" {
			accountValues = append(accountValues, strings.TrimSpace(value))
		}
	}
	if len(accountValues) == 0 {
		accountValues = append(accountValues, "—")
	}
	account := strings.Join(accountValues, " - ")
	lines := []string{
		"[NẠP TIỀN]",
		"Số tiền: " + amount + " VND",
		"Thời gian: " + when,
		"Mã lệnh nạp: " + firstNonEmpty(lookup.ClientRef, notification.ClientRef, "—"),
		"User: " + user,
		"Tài khoản nhận: " + account,
		"Trạng thái: " + status,
	}
	if lookup.Amount != "" && formatVND(notification.Amount) != formatVND(lookup.Amount) {
		lines = append(lines, "Số tiền theo lệnh: "+formatVND(lookup.Amount)+" VND")
	}
	if lookup.TransactionID > 0 {
		lines = append(lines, fmt.Sprintf("ID giao dịch hệ thống: #%d", lookup.TransactionID))
	}
	return strings.Join(lines, "\n")
}

func rawString(raw map[string]any, key string) string {
	if value, ok := raw[key]; ok && value != nil {
		return strings.TrimSpace(fmt.Sprint(value))
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func sanitizeLine(value string, limit int) string {
	value = strings.NewReplacer("\r", " ", "\n", " ", "\t", " ").Replace(strings.TrimSpace(value))
	if len([]rune(value)) > limit {
		return string([]rune(value)[:limit])
	}
	return value
}

func formatVND(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0"
	}
	negative := strings.HasPrefix(value, "-")
	value = strings.TrimPrefix(value, "-")
	parts := strings.SplitN(value, ".", 2)
	integer := parts[0]
	groups := make([]string, 0, 4)
	for len(integer) > 3 {
		idx := len(integer) - 3
		groups = append([]string{integer[idx:]}, groups...)
		integer = integer[:idx]
	}
	if len(groups) > 0 {
		integer += "." + strings.Join(groups, ".")
	}
	if negative {
		integer = "-" + integer
	}
	if len(parts) == 2 && strings.Trim(parts[1], "0") != "" {
		return integer + "," + parts[1]
	}
	return integer
}
