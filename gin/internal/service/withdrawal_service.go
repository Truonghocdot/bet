package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"gin/internal/auth/password"
	"gin/internal/domain/withdrawal"
	"gin/internal/repository/postgres"
	"gin/internal/support/message"

	goredis "github.com/redis/go-redis/v9"
)

type WithdrawalService struct {
	repo       *postgres.WithdrawalRepository
	walletRepo *postgres.WalletRepository
	userRepo   *postgres.UserRepository
	redis      *goredis.Client
}

const (
	defaultWithdrawalFee = "0"
)

func NewWithdrawalService(repo *postgres.WithdrawalRepository, walletRepo *postgres.WalletRepository, userRepo *postgres.UserRepository, redis *goredis.Client) *WithdrawalService {
	return &WithdrawalService{
		repo:       repo,
		walletRepo: walletRepo,
		userRepo:   userRepo,
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
	if strings.TrimSpace(req.Password) == "" {
		return 0, errors.New(message.CurrentPasswordRequired)
	}

	passwordHash, err := s.userRepo.FindPasswordHashByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	if err := password.Compare(passwordHash, strings.TrimSpace(req.Password)); err != nil {
		return 0, errors.New(message.CurrentPasswordInvalid)
	}

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

	wallet, err := s.walletRepo.FindByUserAndUnit(ctx, userID, account.Unit)
	if err != nil {
		return 0, fmt.Errorf("hệ thống lỗi hoặc chưa có ví cho loại tiền này: %w", err)
	}

	amountStr := amountRat.FloatString(8)
	feePercentRat, err := parsePositiveOrZero(defaultWithdrawalFee)
	if err != nil {
		return 0, errors.New("cấu hình phí rút mặc định không hợp lệ")
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

func (s *WithdrawalService) ListHistory(ctx context.Context, userID int64, page, pageSize int) (withdrawal.WithdrawalHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	total, err := s.repo.CountUserWithdrawalRequests(ctx, userID)
	if err != nil {
		return withdrawal.WithdrawalHistoryResponse{}, err
	}

	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * pageSize
	items, err := s.repo.ListWithdrawalRequests(ctx, userID, pageSize, offset)
	if err != nil {
		return withdrawal.WithdrawalHistoryResponse{}, err
	}

	return withdrawal.WithdrawalHistoryResponse{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		Data:       items,
	}, nil
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
