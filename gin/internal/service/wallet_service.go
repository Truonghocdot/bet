package service

import (
	"context"
	"encoding/json"
	"fmt"

	"gin/internal/domain/user"
	"gin/internal/domain/wallet"
	"gin/internal/realtime"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/message"
	"math/big"

	goredis "github.com/redis/go-redis/v9"
)

const (
	ExchangeRateUSDTToVNDDefault = 25000
	ExchangeRateRedisKey         = "shared:exchange-rate:usdt-vnd"
	DefaultWithdrawFeePercent    = "0"
	DefaultWithdrawRequiredBet   = "0"
	DefaultWithdrawMaxTimes      = 3
	DefaultWithdrawMinAmount     = "200000"
	DefaultWithdrawMaxAmount     = "20000000"
)

type WalletService struct {
	repository *repopg.WalletRepository
	broker     *realtime.Broker
	redis      *goredis.Client
}

func NewWalletService(repository *repopg.WalletRepository, broker *realtime.Broker, redis *goredis.Client) *WalletService {
	return &WalletService{repository: repository, broker: broker, redis: redis}
}

func (s *WalletService) Summary(ctx context.Context, userID int64) (wallet.WalletSummaryResponse, error) {
	snapshot := s.getSnapshot(ctx)
	items := make([]wallet.WalletBalance, 0)

	if userID != 0 {
		records, err := s.repository.ListByUserID(ctx, userID)
		if err != nil {
			return wallet.WalletSummaryResponse{}, err
		}

		items = make([]wallet.WalletBalance, 0, len(records))
		for _, record := range records {
			unitCode, unitLabel := walletUnitLabel(record.Unit)
			withdrawCreditLimit, creditErr := s.repository.GetLatestSuccessfulDepositAmount(ctx, userID, record.Unit)
			if creditErr != nil {
				return wallet.WalletSummaryResponse{}, creditErr
			}
			withdrawAvailable, availableErr := repopg.AddNumeric(record.Balance, withdrawCreditLimit)
			if availableErr != nil {
				return wallet.WalletSummaryResponse{}, availableErr
			}
			items = append(items, wallet.WalletBalance{
				ID:                  record.ID,
				Unit:                record.Unit,
				UnitCode:            unitCode,
				UnitLabel:           unitLabel,
				Balance:             record.Balance,
				LockedBalance:       record.LockedBalance,
				WithdrawCreditLimit: withdrawCreditLimit,
				WithdrawAvailable:   withdrawAvailable,
				Status:              record.Status,
				CreatedAt:           record.CreatedAt,
				UpdatedAt:           record.UpdatedAt,
			})
		}
	}

	return wallet.WalletSummaryResponse{
		Message:          message.WalletSummarySuccess,
		ExchangeRate:     snapshot.Rate,
		TelegramCskhLink: snapshot.TelegramCskhLink,
		Marquee: wallet.MarqueeDisplay{
			Enabled:  snapshot.MarqueeEnabled != nil && *snapshot.MarqueeEnabled,
			Messages: snapshot.MarqueeMessages,
		},
		Popup: wallet.PopupDisplay{
			Message:    stringPtrOrNil(snapshot.PopupMessage),
			LatestNews: stringPtrOrNil(snapshot.LatestNewsPopup),
		},
		WithdrawPolicy: wallet.WithdrawPolicyDisplay{
			Enabled:           snapshot.WithdrawPolicyEnabled != nil && *snapshot.WithdrawPolicyEnabled,
			FeePercent:        snapshot.WithdrawFeePercent,
			RequiredBetVolume: snapshot.WithdrawRequiredBet,
			MaxTimesPerDay:    snapshot.WithdrawMaxTimes,
			MinAmount:         snapshot.WithdrawMinAmount,
			MaxAmount:         snapshot.WithdrawMaxAmount,
		},
		Wallets: items,
	}, nil
}

func (s *WalletService) PublishSummary(ctx context.Context, userID int64) error {
	if userID == 0 {
		return nil
	}

	response, err := s.Summary(ctx, userID)
	if err != nil {
		return err
	}

	return s.broker.Publish(ctx, realtime.WalletUserTopic(userID), "wallet.summary", response)
}

func (s *WalletService) Exchange(ctx context.Context, userID int64, req wallet.ExchangeRequest) (wallet.ExchangeResponse, error) {
	if userID == 0 {
		return wallet.ExchangeResponse{}, ErrUnauthorized
	}

	if req.FromUnit == req.ToUnit {
		return wallet.ExchangeResponse{}, fmt.Errorf("không thể chuyển đổi cùng một loại ví")
	}

	amountRat := new(big.Rat)
	if _, ok := amountRat.SetString(req.Amount); !ok {
		return wallet.ExchangeResponse{}, fmt.Errorf("số tiền không hợp lệ")
	}

	if amountRat.Cmp(new(big.Rat)) <= 0 {
		return wallet.ExchangeResponse{}, fmt.Errorf("số tiền phải lớn hơn 0")
	}

	rateStr := s.GetExchangeRate(ctx)
	rateRat := new(big.Rat)
	if _, ok := rateRat.SetString(rateStr); !ok {
		rateRat.SetInt64(ExchangeRateUSDTToVNDDefault)
	}

	var toAmount string

	if req.FromUnit == user.WalletUnitUSDT && req.ToUnit == user.WalletUnitVND {
		// USDT -> VND
		toAmount = new(big.Rat).Mul(amountRat, rateRat).FloatString(0) // VND no decimal
	} else if req.FromUnit == user.WalletUnitVND && req.ToUnit == user.WalletUnitUSDT {
		// VND -> USDT
		toAmount = new(big.Rat).Quo(amountRat, rateRat).FloatString(8)
	} else {
		return wallet.ExchangeResponse{}, fmt.Errorf("cặp tiền tệ chưa được hỗ trợ")
	}

	err := s.repository.Exchange(ctx, userID, req.FromUnit, req.ToUnit, req.Amount, toAmount)
	if err != nil {
		return wallet.ExchangeResponse{}, err
	}

	// Publish updated summary
	_ = s.PublishSummary(ctx, userID)

	return wallet.ExchangeResponse{
		Message:      "Chuyển đổi thành công",
		FromUnit:     req.FromUnit,
		ToUnit:       req.ToUnit,
		FromAmount:   req.Amount,
		ToAmount:     toAmount,
		ExchangeRate: rateRat.FloatString(0),
	}, nil
}

func (s *WalletService) GetExchangeRate(ctx context.Context) string {
	return s.getSnapshot(ctx).Rate
}

type systemSnapshot struct {
	Rate                  string   `json:"rate"`
	TelegramCskhLink      string   `json:"telegram_cskh_link"`
	MarqueeEnabled        *bool    `json:"marquee_enabled"`
	MarqueeMessages       []string `json:"marquee_messages_list"`
	PopupMessage          string   `json:"popup_message"`
	LatestNewsPopup       string   `json:"latest_news_popup"`
	WithdrawPolicyEnabled *bool    `json:"withdraw_policy_enabled"`
	WithdrawFeePercent    string   `json:"withdraw_fee_percent"`
	WithdrawRequiredBet   string   `json:"withdraw_required_bet_volume"`
	WithdrawMaxTimes      int      `json:"withdraw_max_times_per_day"`
	WithdrawMinAmount     string   `json:"withdraw_min_amount"`
	WithdrawMaxAmount     string   `json:"withdraw_max_amount"`
}

func (s *WalletService) getSnapshot(ctx context.Context) systemSnapshot {
	defaultEnabled := true
	defaultSnap := systemSnapshot{
		Rate:           fmt.Sprintf("%d", ExchangeRateUSDTToVNDDefault),
		MarqueeEnabled: &defaultEnabled,
		MarqueeMessages: []string{
			"Quý khách thân mến vui lòng thay đổi cổng nạp tiền nếu không thể tạo lệnh nạp.",
			"Khi nạp tiền bằng cổng CHUYỂN KHOẢN sẽ được nhận thêm ưu đãi đặc biệt!",
			"FF789 - Đăng ký hôm nay nhận ngay thưởng chào mừng 100%.",
		},
		WithdrawPolicyEnabled: &defaultEnabled,
		WithdrawFeePercent:    DefaultWithdrawFeePercent,
		WithdrawRequiredBet:   DefaultWithdrawRequiredBet,
		WithdrawMaxTimes:      DefaultWithdrawMaxTimes,
		WithdrawMinAmount:     DefaultWithdrawMinAmount,
		WithdrawMaxAmount:     DefaultWithdrawMaxAmount,
	}

	val, err := s.redis.Get(ctx, ExchangeRateRedisKey).Result()
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

func walletUnitLabel(unit int) (string, string) {
	switch unit {
	case user.WalletUnitVND:
		return "VND", "Ví VND"
	case user.WalletUnitUSDT:
		return "USDT", "Ví USDT"
	default:
		return fmt.Sprintf("UNIT_%d", unit), fmt.Sprintf("Ví %d", unit)
	}
}
