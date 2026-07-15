package service

import (
	"context"
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
	DefaultWithdrawMaxAmount     = "30000000"
)

type WalletService struct {
	repository       *repopg.WalletRepository
	broker           *realtime.Broker
	redis            *goredis.Client
	contentAssetBase string
}

func NewWalletService(repository *repopg.WalletRepository, broker *realtime.Broker, redis *goredis.Client, contentAssetBase string) *WalletService {
	return &WalletService{
		repository:       repository,
		broker:           broker,
		redis:            redis,
		contentAssetBase: contentAssetBase,
	}
}

func (s *WalletService) Summary(ctx context.Context, userID int64) (wallet.WalletSummaryResponse, error) {
	snapshot := loadSystemSnapshot(ctx, s.redis)
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
		Message:                  message.WalletSummarySuccess,
		ExchangeRate:             snapshot.Rate,
		TelegramCskhLink:         snapshot.TelegramCskhLink,
		AppHeaderLogoURL:         buildPublicAssetURL(s.contentAssetBase, snapshot.AppHeaderLogoPath),
		AppHeaderLogoFallbackURL: buildPublicAssetURL(s.contentAssetBase, snapshot.AppHeaderLogoWebpPath),
		Marquee: wallet.MarqueeDisplay{
			Enabled:  snapshot.MarqueeEnabled != nil && *snapshot.MarqueeEnabled,
			Messages: snapshot.MarqueeMessages,
		},
		FakeFinanceFeed: wallet.FakeFinanceFeedDisplay{
			Enabled: snapshot.FakeFinanceFeedEnabled != nil && *snapshot.FakeFinanceFeedEnabled,
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
	return loadSystemSnapshot(ctx, s.redis).Rate
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
