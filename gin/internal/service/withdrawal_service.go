package service

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"gin/internal/domain/withdrawal"
	"gin/internal/repository/postgres"
)

type WithdrawalService struct {
	repo       *postgres.WithdrawalRepository
	walletRepo *postgres.WalletRepository
}

func NewWithdrawalService(repo *postgres.WithdrawalRepository, walletRepo *postgres.WalletRepository) *WithdrawalService {
	return &WithdrawalService{
		repo:       repo,
		walletRepo: walletRepo,
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
		return 0, errors.New("đóng tối thiểu số lượng không hợp lệ")
	}

	// Validate minimum withdrawal amounts based on unit
	if account.Unit == 1 {
		// VND: min 50,000
		minVND := big.NewRat(50000, 1)
		if amountRat.Cmp(minVND) < 0 {
			return 0, errors.New("số tiền rút tối thiểu là 50,000")
		}
	} else if account.Unit == 2 {
		// USDT: min 5
		minUSDT := big.NewRat(5, 1)
		if amountRat.Cmp(minUSDT) < 0 {
			return 0, errors.New("số tiền rút tối thiểu là 5")
		}
	}

	wallet, err := s.walletRepo.FindByUserAndUnit(ctx, userID, account.Unit)
	if err != nil {
		return 0, fmt.Errorf("hệ thống lỗi hoặc chưa có ví cho loại tiền này: %w", err)
	}

	amountStr := amountRat.FloatString(8)
	fee := "0" // Future expansion: Calculate fee if necessary
	netAmount := amountStr

	return s.repo.CreateWithdrawalRequest(ctx, userID, wallet.ID, account.ID, account.Unit, amountStr, fee, netAmount)
}

func (s *WithdrawalService) ListHistory(ctx context.Context, userID int64, limit, offset int) ([]withdrawal.WithdrawalRequest, error) {
	return s.repo.ListWithdrawalRequests(ctx, userID, limit, offset)
}
