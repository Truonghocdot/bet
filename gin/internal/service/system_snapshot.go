package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	goredis "github.com/redis/go-redis/v9"
)

type systemSnapshot struct {
	Rate                                string   `json:"rate"`
	TelegramCskhLink                    string   `json:"telegram_cskh_link"`
	AppHeaderLogoPath                   string   `json:"app_header_logo_path"`
	AppHeaderLogoWebpPath               string   `json:"app_header_logo_webp_path"`
	MarqueeEnabled                      *bool    `json:"marquee_enabled"`
	FakeFinanceFeedEnabled              *bool    `json:"fake_finance_feed_enabled"`
	NotificationImageForceCancelEnabled *bool    `json:"notification_image_force_cancel_enabled"`
	RegisterSecurityNoticeEnabled       *bool    `json:"register_security_notice_enabled"`
	RegisterSecurityNoticeText          string   `json:"register_security_notice_text"`
	MarqueeMessages                     []string `json:"marquee_messages_list"`
	PopupMessage                        string   `json:"popup_message"`
	LatestNewsPopup                     string   `json:"latest_news_popup"`
	WithdrawPolicyEnabled               *bool    `json:"withdraw_policy_enabled"`
	WithdrawFeePercent                  string   `json:"withdraw_fee_percent"`
	WithdrawRequiredBet                 string   `json:"withdraw_required_bet_volume"`
	WithdrawMaxTimes                    int      `json:"withdraw_max_times_per_day"`
	WithdrawMinAmount                   string   `json:"withdraw_min_amount"`
	WithdrawMaxAmount                   string   `json:"withdraw_max_amount"`
}

func loadSystemSnapshot(ctx context.Context, redis *goredis.Client) systemSnapshot {
	defaultEnabled := true
	defaultDisabled := false
	defaultSnap := systemSnapshot{
		Rate:                                fmt.Sprintf("%d", ExchangeRateUSDTToVNDDefault),
		MarqueeEnabled:                      &defaultEnabled,
		FakeFinanceFeedEnabled:              &defaultEnabled,
		NotificationImageForceCancelEnabled: &defaultDisabled,
		RegisterSecurityNoticeEnabled:       &defaultDisabled,
		MarqueeMessages: []string{
			"Quý khách thân mến vui lòng thay đổi cổng nạp tiền nếu không thể tạo lệnh nạp.",
			"Khi nạp tiền bằng cổng CHUYỂN KHOẢN sẽ được nhận thêm ưu đãi đặc biệt!",
			"fh88u - Đăng ký hôm nay nhận ngay thưởng chào mừng 100%.",
		},
		WithdrawPolicyEnabled: &defaultEnabled,
		WithdrawFeePercent:    DefaultWithdrawFeePercent,
		WithdrawRequiredBet:   DefaultWithdrawRequiredBet,
		WithdrawMaxTimes:      DefaultWithdrawMaxTimes,
		WithdrawMinAmount:     DefaultWithdrawMinAmount,
		WithdrawMaxAmount:     DefaultWithdrawMaxAmount,
	}

	if redis == nil {
		return defaultSnap
	}

	val, err := redis.Get(ctx, ExchangeRateRedisKey).Result()
	if err != nil {
		return defaultSnap
	}

	var snapshot systemSnapshot
	if err := json.Unmarshal([]byte(val), &snapshot); err != nil {
		return defaultSnap
	}

	if snapshot.Rate == "" {
		snapshot.Rate = defaultSnap.Rate
	}
	if snapshot.MarqueeEnabled == nil {
		snapshot.MarqueeEnabled = defaultSnap.MarqueeEnabled
	}
	if snapshot.FakeFinanceFeedEnabled == nil {
		snapshot.FakeFinanceFeedEnabled = defaultSnap.FakeFinanceFeedEnabled
	}
	if snapshot.NotificationImageForceCancelEnabled == nil {
		snapshot.NotificationImageForceCancelEnabled = defaultSnap.NotificationImageForceCancelEnabled
	}
	if snapshot.RegisterSecurityNoticeEnabled == nil {
		snapshot.RegisterSecurityNoticeEnabled = defaultSnap.RegisterSecurityNoticeEnabled
	}
	if len(snapshot.MarqueeMessages) == 0 {
		snapshot.MarqueeMessages = defaultSnap.MarqueeMessages
	}
	if snapshot.WithdrawPolicyEnabled == nil {
		snapshot.WithdrawPolicyEnabled = defaultSnap.WithdrawPolicyEnabled
	}
	if snapshot.WithdrawFeePercent == "" {
		snapshot.WithdrawFeePercent = defaultSnap.WithdrawFeePercent
	}
	if snapshot.WithdrawRequiredBet == "" {
		snapshot.WithdrawRequiredBet = defaultSnap.WithdrawRequiredBet
	}
	if snapshot.WithdrawMaxTimes <= 0 {
		snapshot.WithdrawMaxTimes = defaultSnap.WithdrawMaxTimes
	}
	if snapshot.WithdrawMinAmount == "" {
		snapshot.WithdrawMinAmount = defaultSnap.WithdrawMinAmount
	}
	if snapshot.WithdrawMaxAmount == "" {
		snapshot.WithdrawMaxAmount = defaultSnap.WithdrawMaxAmount
	}

	return snapshot
}

func buildPublicAssetURL(base, path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "/storage/") || strings.HasPrefix(trimmed, "storage/") {
		cleanPath := strings.TrimPrefix(trimmed, "/")
		if strings.TrimSpace(base) == "" {
			return "/" + cleanPath
		}
		return strings.TrimRight(strings.TrimSpace(base), "/") + "/" + cleanPath
	}
	if strings.TrimSpace(base) == "" {
		return "/storage/" + strings.TrimPrefix(trimmed, "/")
	}
	return strings.TrimRight(strings.TrimSpace(base), "/") + "/storage/" + strings.TrimPrefix(trimmed, "/")
}
