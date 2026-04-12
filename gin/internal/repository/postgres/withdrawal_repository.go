package postgres

import (
	"context"
	"database/sql"
	"errors"

	"gin/internal/domain/withdrawal"
	"gin/internal/support/clock"
)

var (
	ErrWithdrawalAccountNotFound = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type WithdrawalRepository struct {
	db *sql.DB
}

func NewWithdrawalRepository(db *sql.DB) *WithdrawalRepository {
	return &WithdrawalRepository{db: db}
}

func (r *WithdrawalRepository) ListAccounts(ctx context.Context, userID int64) ([]withdrawal.AccountWithdrawalInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, unit, provider_code, account_name, account_number, is_default, created_at
		from account_withdrawal_infos
		where user_id = $1 and deleted_at is null
		order by created_at desc
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []withdrawal.AccountWithdrawalInfo
	for rows.Next() {
		var record withdrawal.AccountWithdrawalInfo
		if err := rows.Scan(
			&record.ID,
			&record.Unit,
			&record.ProviderCode,
			&record.AccountName,
			&record.AccountNumber,
			&record.IsDefault,
			&record.CreatedAt,
		); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *WithdrawalRepository) CreateAccount(ctx context.Context, userID int64, req withdrawal.SetupAccountRequest) (int64, error) {
	var id int64
	now := clock.Now()
	
	// If the new account is default, we should probably unset others, but the schema doesn't strictly enforce one default.
	// For simplicity, we just insert.
	err := r.db.QueryRowContext(ctx, `
		insert into account_withdrawal_infos (
			user_id, unit, provider_code, account_name, account_number, is_default, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $7)
		returning id
	`, userID, req.Unit, req.ProviderCode, req.AccountName, req.AccountNumber, req.IsDefault, now).Scan(&id)
	
	return id, err
}

func (r *WithdrawalRepository) DeleteAccount(ctx context.Context, userID, accountID int64) error {
	result, err := r.db.ExecContext(ctx, `
		update account_withdrawal_infos
		set deleted_at = $1
		where id = $2 and user_id = $3 and deleted_at is null
	`, clock.Now(), accountID, userID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrWithdrawalAccountNotFound
	}
	return nil
}

func (r *WithdrawalRepository) GetAccount(ctx context.Context, userID, accountID int64) (withdrawal.AccountWithdrawalInfo, error) {
	row := r.db.QueryRowContext(ctx, `
		select id, unit, provider_code, account_name, account_number, is_default, created_at
		from account_withdrawal_infos
		where id = $1 and user_id = $2 and deleted_at is null
	`, accountID, userID)

	var record withdrawal.AccountWithdrawalInfo
	if err := row.Scan(
		&record.ID,
		&record.Unit,
		&record.ProviderCode,
		&record.AccountName,
		&record.AccountNumber,
		&record.IsDefault,
		&record.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return withdrawal.AccountWithdrawalInfo{}, ErrWithdrawalAccountNotFound
		}
		return withdrawal.AccountWithdrawalInfo{}, err
	}
	return record, nil
}

func (r *WithdrawalRepository) CreateWithdrawalRequest(ctx context.Context, userID, walletID, accountID int64, unit int, amount, fee, netAmount string) (int64, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	now := clock.Now()

	var balanceBefore, lockedBefore string
	if err := tx.QueryRowContext(ctx, `
		select balance::text, locked_balance::text
		from wallets
		where id = $1 and user_id = $2 and unit = $3
		for update
	`, walletID, userID, unit).Scan(&balanceBefore, &lockedBefore); err != nil {
		return 0, err
	}

	balanceAfter, err := subtractNumeric(balanceBefore, amount)
	if err != nil {
		return 0, err
	}
	if compareNumeric(balanceAfter, "0") < 0 {
		return 0, ErrInsufficientBalance
	}

	lockedAfter, err := addNumeric(lockedBefore, amount)
	if err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		update wallets
		set balance = $1::numeric(20,8),
		    locked_balance = $2::numeric(20,8),
		    updated_at = $3
		where id = $4
	`, balanceAfter, lockedAfter, now, walletID); err != nil {
		return 0, err
	}

	var requestID int64
	if err := tx.QueryRowContext(ctx, `
		insert into withdrawal_requests (
			user_id, wallet_id, account_withdrawal_info_id, unit, amount, fee, net_amount, status, created_at, updated_at
		) values (
			$1, $2, $3, $4, $5::numeric(20,8), $6::numeric(20,8), $7::numeric(20,8), 0, $8, $8
		) returning id
	`, userID, walletID, accountID, unit, amount, fee, netAmount, now).Scan(&requestID); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (
			wallet_id, user_id, direction, amount, balance_before, balance_after,
			reference_type, reference_id, note, created_at
		) values (
			$1, $2, 2, $3::numeric(20,8), $4::numeric(20,8), $5::numeric(20,8),
			'App\\Models\\Transaction\\WithdrawalRequest', $6, 'Khóa tiền tạo lệnh rút', $7
		)
	`, walletID, userID, amount, balanceBefore, balanceAfter, requestID, now); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return requestID, nil
}
