package postgres

import (
	"context"
	"database/sql"
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
