package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gin/internal/domain/withdrawal"
	"gin/internal/repository/postgres"
	"gin/internal/support/clock"

	goredis "github.com/redis/go-redis/v9"
)

type WithdrawalService struct {
	repo       *postgres.WithdrawalRepository
	walletRepo *postgres.WalletRepository
	redis      *goredis.Client
}

const (
	withdrawPolicyRedisKey            = "shared:exchange-rate:usdt-vnd"
	defaultWithdrawFeePercent         = "0"
	defaultWithdrawRequiredBetVolume  = "0"
	defaultWithdrawMaxTimesPerDay     = 3
	defaultWithdrawMinAmount          = "200000"
	defaultWithdrawMaxAmount          = "20000000"
	defaultExchangeRateForPolicy      = "25000"
	defaultWithdrawMinUSDT            = "5"
)

type withdrawalPolicySnapshot struct {
	Rate                  string `json:"rate"`
	WithdrawFeePercent    string `json:"withdraw_fee_percent"`
	RequiredBetVolume     string `json:"withdraw_required_bet_volume"`
	MaxTimesPerDay        int    `json:"withdraw_max_times_per_day"`
	MinAmount             string `json:"withdraw_min_amount"`
	MaxAmount             string `json:"withdraw_max_amount"`
}

func NewWithdrawalService(repo *postgres.WithdrawalRepository, walletRepo *postgres.WalletRepository, redis *goredis.Client) *WithdrawalService {
	return &WithdrawalService{
		repo:       repo,
		walletRepo: walletRepo,
		redis:      redis,
	}
}

func (s *WithdrawalService) ListUserAccounts(ctx context.Context, userID int64) ([]withdrawal.AccountWithdrawalInfo, error) {
	return s.repo.ListAccounts(ctx, userID)
}

func (s *WithdrawalService) AddAccount(ctx context.Context, userID int64, req withdrawal.SetupAccountRequest) (int64, error) {
	return s.repo.CreateAccount(ctx, userID, req)
}

func (s *WithdrawalService) DeleteAccount(ctx context.Context, userID int64, id int64) error {
	return s.repo.DeleteAccount(ctx, userID, id)
}

func (s *WithdrawalService) SubmitWithdrawalRequest(ctx context.Context, userID int64, req withdrawal.SubmitWithdrawalRequest) (int64, error) {
	account, err := s.repo.GetAccount(ctx, userID, req.AccountWithdrawalInfoID)
	if err != nil {
		return 0, fmt.Errorf("invalid account: %w", err)
	}

	amountRat := new(big.Rat)
	if _, ok := amountRat.SetString(strings.TrimSpace(req.Amount)); !ok {
		return 0, errors.New("invalid amount format")
	}

	if amountRat.Sign() <= 0 {
		return 0, errors.New("số tiền rút phải lớn hơn 0")
	}

	policy := s.getWithdrawalPolicy(ctx)

	// Max số lần rút / ngày (giờ VN).
	vnNow := clock.Now().In(time.FixedZone("Asia/Ho_Chi_Minh", 7*3600))
	dayStartVN := time.Date(vnNow.Year(), vnNow.Month(), vnNow.Day(), 0, 0, 0, 0, vnNow.Location())
	dayEndVN := dayStartVN.Add(24 * time.Hour)
	withdrawCountToday, err := s.repo.CountUserWithdrawalRequestsInRange(ctx, userID, dayStartVN.UTC(), dayEndVN.UTC())
	if err != nil {
		return 0, fmt.Errorf("không thể kiểm tra số lần rút trong ngày: %w", err)
	}
	if policy.MaxTimesPerDay > 0 && withdrawCountToday >= policy.MaxTimesPerDay {
		return 0, fmt.Errorf("bạn đã đạt giới hạn %d lần rút trong hôm nay", policy.MaxTimesPerDay)
	}

	// Tổng khối lượng cược tối thiểu.
	requiredVolume, err := parsePositiveOrZero(policy.RequiredBetVolume)
	if err != nil {
		return 0, errors.New("cấu hình tổng tiền cược tối thiểu không hợp lệ")
	}
	if requiredVolume.Sign() > 0 {
		userBetVolumeStr, err := s.repo.SumUserBetVolume(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("không thể kiểm tra tổng khối lượng cược: %w", err)
		}
		userBetVolume, err := parsePositiveOrZero(userBetVolumeStr)
		if err != nil {
			return 0, errors.New("tổng khối lượng cược hiện tại không hợp lệ")
		}
		if userBetVolume.Cmp(requiredVolume) < 0 {
			return 0, fmt.Errorf("khối lượng cược chưa đủ: yêu cầu tối thiểu %s", requiredVolume.FloatString(2))
		}
	}

	minAmountRat, err := parsePositiveOrZero(policy.MinAmount)
	if err != nil {
		return 0, errors.New("cấu hình rút tối thiểu không hợp lệ")
	}
	maxAmountRat, err := parsePositiveOrZero(policy.MaxAmount)
	if err != nil {
		return 0, errors.New("cấu hình rút tối đa không hợp lệ")
	}

	// Chính sách min/max được cấu hình theo VND. Nếu rút USDT thì quy đổi bằng rate hiện tại.
	if account.Unit == 2 {
		rateRat, err := parsePositiveOrZero(policy.Rate)
		if err != nil || rateRat.Sign() <= 0 {
			rateRat, _ = parsePositiveOrZero(defaultExchangeRateForPolicy)
		}
		minAmountRat = new(big.Rat).Quo(minAmountRat, rateRat)
		maxAmountRat = new(big.Rat).Quo(maxAmountRat, rateRat)

		// Giữ baseline kỹ thuật USDT tối thiểu 5.
		minUSDTFallback, _ := parsePositiveOrZero(defaultWithdrawMinUSDT)
		if minAmountRat.Cmp(minUSDTFallback) < 0 {
			minAmountRat = minUSDTFallback
		}
	}
	if minAmountRat.Sign() > 0 && amountRat.Cmp(minAmountRat) < 0 {
		return 0, fmt.Errorf("số tiền rút tối thiểu là %s", minAmountRat.FloatString(2))
	}
	if maxAmountRat.Sign() > 0 && amountRat.Cmp(maxAmountRat) > 0 {
		return 0, fmt.Errorf("số tiền rút tối đa là %s", maxAmountRat.FloatString(2))
	}

	wallet, err := s.walletRepo.FindByUserAndUnit(ctx, userID, account.Unit)
	if err != nil {
		return 0, fmt.Errorf("hệ thống lỗi hoặc chưa có ví cho loại tiền này: %w", err)
	}

	amountStr := amountRat.FloatString(8)
	feePercentRat, err := parsePositiveOrZero(policy.WithdrawFeePercent)
	if err != nil {
		return 0, errors.New("cấu hình phí rút không hợp lệ")
	}
	feeRat := new(big.Rat)
	if feePercentRat.Sign() > 0 {
		feeRat = new(big.Rat).Quo(new(big.Rat).Mul(amountRat, feePercentRat), big.NewRat(100, 1))
	}
	netAmountRat := new(big.Rat).Sub(amountRat, feeRat)
	if netAmountRat.Sign() <= 0 {
		return 0, errors.New("số tiền nhận sau phí không hợp lệ")
	}

	fee := feeRat.FloatString(8)
	netAmount := netAmountRat.FloatString(8)

	return s.repo.CreateWithdrawalRequest(ctx, userID, wallet.ID, account.ID, account.Unit, amountStr, fee, netAmount)
}

func (s *WithdrawalService) ListHistory(ctx context.Context, userID int64, limit, offset int) ([]withdrawal.WithdrawalRequest, error) {
	return s.repo.ListWithdrawalRequests(ctx, userID, limit, offset)
}

func (s *WithdrawalService) getWithdrawalPolicy(ctx context.Context) withdrawalPolicySnapshot {
	policy := withdrawalPolicySnapshot{
		Rate:               defaultExchangeRateForPolicy,
		WithdrawFeePercent: defaultWithdrawFeePercent,
		RequiredBetVolume:  defaultWithdrawRequiredBetVolume,
		MaxTimesPerDay:     defaultWithdrawMaxTimesPerDay,
		MinAmount:          defaultWithdrawMinAmount,
		MaxAmount:          defaultWithdrawMaxAmount,
	}

	if s.redis == nil {
		return policy
	}

	raw, err := s.redis.Get(ctx, withdrawPolicyRedisKey).Result()
	if err != nil {
		return policy
	}

	var snap withdrawalPolicySnapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		return policy
	}

	if strings.TrimSpace(snap.Rate) != "" {
		policy.Rate = strings.TrimSpace(snap.Rate)
	}
	if strings.TrimSpace(snap.WithdrawFeePercent) != "" {
		policy.WithdrawFeePercent = strings.TrimSpace(snap.WithdrawFeePercent)
	}
	if strings.TrimSpace(snap.RequiredBetVolume) != "" {
		policy.RequiredBetVolume = strings.TrimSpace(snap.RequiredBetVolume)
	}
	if snap.MaxTimesPerDay > 0 {
		policy.MaxTimesPerDay = snap.MaxTimesPerDay
	}
	if strings.TrimSpace(snap.MinAmount) != "" {
		policy.MinAmount = strings.TrimSpace(snap.MinAmount)
	}
	if strings.TrimSpace(snap.MaxAmount) != "" {
		policy.MaxAmount = strings.TrimSpace(snap.MaxAmount)
	}

	return policy
}

func parsePositiveOrZero(value string) (*big.Rat, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(strings.TrimSpace(value)); !ok {
		return nil, errors.New("invalid numeric")
	}
	if r.Sign() < 0 {
		return nil, errors.New("negative numeric")
	}
	return r, nil
}
