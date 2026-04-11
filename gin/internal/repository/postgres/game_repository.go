package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gin/internal/support/message"
)

var (
	ErrBetTicketNotFound        = errors.New("không tìm thấy lệnh cược")
	ErrInsufficientBetBalance   = errors.New(message.InsufficientBalanceBet)
	ErrInsufficientPlayBalance  = errors.New(message.InsufficientBalancePlay)
)

type GameRepository struct {
	db *sql.DB
}

type GameRoundRecord struct {
	ID        int64
	GameType  string
	PeriodNo  string
	Result    string
	BigSmall  string
	Color     string
	DrawAt    time.Time
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BetTicketItemRecord struct {
	OptionType string `json:"option_type"`
	OptionKey  string `json:"option_key"`
	Stake      string `json:"stake"`
}

type BetTicketRecord struct {
	ID           int64
	UserID       int64
	WalletID     int64
	GameType     string
	PeriodID     string
	RequestID    string
	ConnectionID string
	TotalStake   string
	Status       string
	ItemsJSON    []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) ListGameRounds(ctx context.Context, gameType string, page, pageSize int) ([]GameRoundRecord, int, error) {
	page, pageSize = normalizePagination(page, pageSize)
	var total int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from game_round_histories
		where game_type = $1
	`, gameType).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, game_type, period_no, result, big_small, color, draw_at, status, created_at, updated_at
		from game_round_histories
		where game_type = $1
		order by draw_at desc, id desc
		limit $2 offset $3
	`, gameType, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]GameRoundRecord, 0)
	for rows.Next() {
		var record GameRoundRecord
		if err := rows.Scan(
			&record.ID,
			&record.GameType,
			&record.PeriodNo,
			&record.Result,
			&record.BigSmall,
			&record.Color,
			&record.DrawAt,
			&record.Status,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, rows.Err()
}

func (r *GameRepository) CreateBetTicket(ctx context.Context, params CreateBetTicketParams) (BetTicketRecord, error) {
	itemsJSON, err := json.Marshal(params.Items)
	if err != nil {
		return BetTicketRecord{}, err
	}

	stakeAmount, err := parseNumeric(params.TotalStake)
	if err != nil {
		return BetTicketRecord{}, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return BetTicketRecord{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	walletID, balanceBefore, lockedBefore, err := r.lockWalletForBet(ctx, tx, params.UserID)
	if err != nil {
		return BetTicketRecord{}, err
	}

	availableBefore, err := subtractNumeric(balanceBefore, lockedBefore)
	if err != nil {
		return BetTicketRecord{}, err
	}

	if compareNumeric(availableBefore, params.TotalStake) < 0 {
		return BetTicketRecord{}, ErrInsufficientBetBalance
	}

	lockedAfter, err := addNumeric(lockedBefore, params.TotalStake)
	if err != nil {
		return BetTicketRecord{}, err
	}

	availableAfter, err := subtractNumeric(balanceBefore, lockedAfter)
	if err != nil {
		return BetTicketRecord{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		update wallets
		set locked_balance = $1::numeric(20,8),
		    updated_at = now()
		where id = $2
	`, lockedAfter, walletID); err != nil {
		return BetTicketRecord{}, err
	}

	now := time.Now()
	row := tx.QueryRowContext(ctx, `
		insert into bet_tickets (
			user_id, wallet_id, game_type, period_id, request_id, connection_id,
			total_stake, status, items, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7::numeric(20,8), $8, $9::jsonb, $10, $10)
		returning id, user_id, wallet_id, game_type, period_id, request_id, connection_id,
		          total_stake::text, status, items, created_at, updated_at
	`, params.UserID, walletID, params.GameType, params.PeriodID, params.RequestID, params.ConnectionID, stakeAmount.String(), "LOCKED", itemsJSON, now)

	var record BetTicketRecord
	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.WalletID,
		&record.GameType,
		&record.PeriodID,
		&record.RequestID,
		&record.ConnectionID,
		&record.TotalStake,
		&record.Status,
		&record.ItemsJSON,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return BetTicketRecord{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (
			wallet_id, user_id, direction, amount, balance_before, balance_after,
			reference_type, reference_id, note, created_at
		)
		values ($1, $2, $3, $4::numeric(20,8), $5::numeric(20,8), $6::numeric(20,8),
		        $7, $8, $9, now())
	`, walletID, params.UserID, 2, stakeAmount.String(), availableBefore, availableAfter, "App\\Models\\Game\\BetTicket", record.ID, "Khóa tiền khi đặt cược"); err != nil {
		return BetTicketRecord{}, err
	}

	if err := tx.Commit(); err != nil {
		return BetTicketRecord{}, err
	}

	return record, nil
}

type CreateBetTicketParams struct {
	UserID       int64
	GameType     string
	PeriodID     string
	RequestID    string
	ConnectionID string
	TotalStake   string
	Items        []BetTicketItemRecord
}

func (r *GameRepository) ListBetTickets(ctx context.Context, userID int64, gameType string, page, pageSize int) ([]BetTicketRecord, int, error) {
	page, pageSize = normalizePagination(page, pageSize)

	var total int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from bet_tickets
		where user_id = $1 and game_type = $2
	`, userID, gameType).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, user_id, wallet_id, game_type, period_id, request_id, connection_id,
		       total_stake::text, status, items, created_at, updated_at
		from bet_tickets
		where user_id = $1 and game_type = $2
		order by created_at desc, id desc
		limit $3 offset $4
	`, userID, gameType, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	records := make([]BetTicketRecord, 0)
	for rows.Next() {
		var record BetTicketRecord
		if err := rows.Scan(
			&record.ID,
			&record.UserID,
			&record.WalletID,
			&record.GameType,
			&record.PeriodID,
			&record.RequestID,
			&record.ConnectionID,
			&record.TotalStake,
			&record.Status,
			&record.ItemsJSON,
			&record.CreatedAt,
			&record.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		records = append(records, record)
	}

	return records, total, rows.Err()
}

func (r *GameRepository) lockWalletForBet(ctx context.Context, tx *sql.Tx, userID int64) (walletID int64, balanceBefore string, lockedBefore string, err error) {
	row := tx.QueryRowContext(ctx, `
		select id, balance::text, locked_balance::text
		from wallets
		where user_id = $1 and unit = $2 and status = $3
		limit 1
		for update
	`, userID, 1, 1)

	if err := row.Scan(&walletID, &balanceBefore, &lockedBefore); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", "", ErrInsufficientBetBalance
		}

		return 0, "", "", err
	}

	return walletID, balanceBefore, lockedBefore, nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}

func parseNumeric(value string) (*big.Rat, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(strings.TrimSpace(value)); !ok {
		return nil, fmt.Errorf("invalid numeric value: %s", value)
	}
	return rat, nil
}

func addNumeric(left, right string) (string, error) {
	lv, err := parseNumeric(left)
	if err != nil {
		return "", err
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Add(lv, rv).FloatString(8), nil
}

func subtractNumeric(left, right string) (string, error) {
	lv, err := parseNumeric(left)
	if err != nil {
		return "", err
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return "", err
	}
	return new(big.Rat).Sub(lv, rv).FloatString(8), nil
}

func compareNumeric(left, right string) int {
	lv, err := parseNumeric(left)
	if err != nil {
		return -1
	}
	rv, err := parseNumeric(right)
	if err != nil {
		return -1
	}
	return lv.Cmp(rv)
}
