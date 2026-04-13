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
	if userID == 0 {
		return wallet.WalletSummaryResponse{}, ErrUnauthorized
	}

	records, err := s.repository.ListByUserID(ctx, userID)
	if err != nil {
		return wallet.WalletSummaryResponse{}, err
	}

	items := make([]wallet.WalletBalance, 0, len(records))
	for _, record := range records {
		unitCode, unitLabel := walletUnitLabel(record.Unit)
		items = append(items, wallet.WalletBalance{
			ID:            record.ID,
			Unit:          record.Unit,
			UnitCode:      unitCode,
			UnitLabel:     unitLabel,
			Balance:       record.Balance,
			LockedBalance: record.LockedBalance,
			Status:        record.Status,
			CreatedAt:     record.CreatedAt,
			UpdatedAt:     record.UpdatedAt,
		})
	}

	return wallet.WalletSummaryResponse{
		Message:      message.WalletSummarySuccess,
		ExchangeRate: s.GetExchangeRate(ctx),
		Wallets:      items,
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
	val, err := s.redis.Get(ctx, ExchangeRateRedisKey).Result()
	if err != nil {
		return fmt.Sprintf("%d", ExchangeRateUSDTToVNDDefault)
	}

	var snapshot struct {
		Rate string `json:"rate"`
	}
	if err := json.Unmarshal([]byte(val), &snapshot); err != nil {
		return fmt.Sprintf("%d", ExchangeRateUSDTToVNDDefault)
	}

	if snapshot.Rate == "" {
		return fmt.Sprintf("%d", ExchangeRateUSDTToVNDDefault)
	}

	return snapshot.Rate
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
