package service

import (
	"context"
	"fmt"

	"gin/internal/domain/user"
	"gin/internal/domain/wallet"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/message"
)

type WalletService struct {
	repository *repopg.WalletRepository
}

func NewWalletService(repository *repopg.WalletRepository) *WalletService {
	return &WalletService{repository: repository}
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
		Message: message.WalletSummarySuccess,
		Wallets: items,
	}, nil
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
