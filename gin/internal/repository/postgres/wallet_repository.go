package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type WalletRepository struct {
	db *sql.DB
}

type WalletRecord struct {
	ID            int64
	UserID        int64
	Unit          int
	Balance       string
	LockedBalance string
	Status        int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) FindByUserAndUnit(ctx context.Context, userID int64, unit int) (WalletRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		select id, user_id, unit, balance::text, locked_balance::text, status, created_at, updated_at
		from wallets
		where user_id = $1 and unit = $2
		limit 1
	`, userID, unit)

	var record WalletRecord
	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.Unit,
		&record.Balance,
		&record.LockedBalance,
		&record.Status,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return WalletRecord{}, err
	}

	return record, nil
}

func (r *WalletRepository) ListByUserID(ctx context.Context, userID int64) ([]WalletRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, user_id, unit, balance::text, locked_balance::text, status, created_at, updated_at
		from wallets
		where user_id = $1
		order by unit asc, id asc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []WalletRecord
	for rows.Next() {
		var record WalletRecord
		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.Unit,
			&record.Balance,
			&record.LockedBalance,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *WalletRepository) GetLatestSuccessfulDepositAmount(ctx context.Context, userID int64, unit int) (string, error) {
	var amount string
	err := r.db.QueryRowContext(ctx, `
		select coalesce(amount, net_amount, 0)::text
		from transactions
		where user_id = $1
		  and unit = $2
		  and type = 1
		  and status = 3
		  and deleted_at is null
		order by coalesce(approved_at, updated_at, created_at) desc, id desc
		limit 1
	`, userID, unit).Scan(&amount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "0", nil
		}
		return "", err
	}
	if amount == "" {
		return "0", nil
	}
	return amount, nil
}

func (r *WalletRepository) Exchange(ctx context.Context, userID int64, fromUnit, toUnit int, fromAmount, toAmount string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Lock From Wallet
	var fromWalletID int64
	var fromBalanceBefore string
	err = tx.QueryRowContext(ctx, `
		select id, balance::text from wallets 
		where user_id = $1 and unit = $2 and status = 1 
		for update
	`, userID, fromUnit).Scan(&fromWalletID, &fromBalanceBefore)
	if err != nil {
		return fmt.Errorf("không tìm thấy ví nguồn: %w", err)
	}

	// Check balance
	if compareNumeric(fromBalanceBefore, fromAmount) < 0 {
		return errors.New("số dư không đủ để chuyển đổi")
	}

	// 2. Lock To Wallet
	var toWalletID int64
	var toBalanceBefore string
	err = tx.QueryRowContext(ctx, `
		select id, balance::text from wallets 
		where user_id = $1 and unit = $2 and status = 1 
		for update
	`, userID, toUnit).Scan(&toWalletID, &toBalanceBefore)
	if err != nil {
		return fmt.Errorf("không tìm thấy ví đích: %w", err)
	}

	fromBalanceAfter, _ := subtractNumeric(fromBalanceBefore, fromAmount)
	toBalanceAfter, _ := addNumeric(toBalanceBefore, toAmount)

	// 3. Update From Wallet
	if _, err := tx.ExecContext(ctx, `
		update wallets set balance = balance - $1::numeric(20,8), updated_at = now()
		where id = $2
	`, fromAmount, fromWalletID); err != nil {
		return err
	}

	// 4. Update To Wallet
	if _, err := tx.ExecContext(ctx, `
		update wallets set balance = balance + $1::numeric(20,8), updated_at = now()
		where id = $2
	`, toAmount, toWalletID); err != nil {
		return err
	}

	// 5. Ledger entries
	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (wallet_id, user_id, direction, amount, balance_before, balance_after, reference_type, note, created_at)
		values ($1, $2, 2, $3::numeric(20,8), $4::numeric(20,8), $5::numeric(20,8), 'exchange', $6, now())
	`, fromWalletID, userID, fromAmount, fromBalanceBefore, fromBalanceAfter, fmt.Sprintf("Chuyển đổi sang %d", toUnit)); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (wallet_id, user_id, direction, amount, balance_before, balance_after, reference_type, note, created_at)
		values ($1, $2, 1, $3::numeric(20,8), $4::numeric(20,8), $5::numeric(20,8), 'exchange', $6, now())
	`, toWalletID, userID, toAmount, toBalanceBefore, toBalanceAfter, fmt.Sprintf("Nhận từ chuyển đổi ví %d", fromUnit)); err != nil {
		return err
	}

	return tx.Commit()
}
