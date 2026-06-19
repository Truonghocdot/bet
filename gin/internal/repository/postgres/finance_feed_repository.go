package postgres

import (
	"context"
	"database/sql"
	"time"
)

type FinanceFeedRepository struct {
	db *sql.DB
}

type FakeFinanceFeedRecord struct {
	ID          int64
	MaskedCode  string
	MaskedPhone string
	StatusLabel string
	CreatedAt   time.Time
}

func NewFinanceFeedRepository(db *sql.DB) *FinanceFeedRepository {
	return &FinanceFeedRepository{db: db}
}

func (r *FinanceFeedRepository) ListFakeDeposits(ctx context.Context, limit int) ([]FakeFinanceFeedRecord, error) {
	return r.list(ctx, "fake_deposit_transactions", limit)
}

func (r *FinanceFeedRepository) ListFakeWithdraws(ctx context.Context, limit int) ([]FakeFinanceFeedRecord, error) {
	return r.list(ctx, "fake_withdraw_transactions", limit)
}

func (r *FinanceFeedRepository) list(ctx context.Context, table string, limit int) ([]FakeFinanceFeedRecord, error) {
	if limit <= 0 {
		limit = 12
	}
	if limit > 50 {
		limit = 50
	}

	switch table {
	case "fake_deposit_transactions", "fake_withdraw_transactions":
	default:
		return nil, sql.ErrNoRows
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, masked_code, masked_phone, status_label, created_at
		from `+table+`
		order by created_at desc, id desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]FakeFinanceFeedRecord, 0, limit)
	for rows.Next() {
		var item FakeFinanceFeedRecord
		if err := rows.Scan(
			&item.ID,
			&item.MaskedCode,
			&item.MaskedPhone,
			&item.StatusLabel,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
