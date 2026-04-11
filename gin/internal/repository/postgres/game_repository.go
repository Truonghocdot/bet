package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	"gin/internal/support/id"
	"gin/internal/support/message"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	periodStatusScheduled = 1
	periodStatusOpen      = 2
	periodStatusLocked    = 3
	periodStatusDrawn     = 4
	periodStatusSettled   = 5

	betStatusPending = 1
	betStatusWon     = 2
	betStatusLost    = 3
)

var (
	ErrBetTicketNotFound       = errors.New("không tìm thấy lệnh cược")
	ErrInsufficientBetBalance  = errors.New(message.InsufficientBalanceBet)
	ErrInsufficientPlayBalance = errors.New(message.InsufficientBalancePlay)
	ErrGameRoomNotFound        = errors.New(message.GameRoomNotFound)
	ErrPeriodNotFound          = errors.New(message.PeriodNotFound)
	ErrPeriodNotOpen           = errors.New(message.PeriodNotOpen)
	ErrPeriodBetLocked         = errors.New(message.PeriodBetLocked)
)

type GameRepository struct {
	db *sql.DB
}

type GameRoomRecord struct {
	Code             string
	GameType         int
	DurationSeconds  int
	BetCutoffSeconds int
	Status           int
	SortOrder        int
}

type GamePeriodRecord struct {
	ID        int64
	RoomCode  string
	GameType  int
	PeriodNo  string
	OpenAt    time.Time
	BetLockAt time.Time
	DrawAt    time.Time
	Status    int
}

type GameRoundRecord struct {
	ID        int64
	GameType  string
	RoomCode  string
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
	GameType     int
	PeriodID     int64
	PeriodNo     string
	RequestID    string
	ConnectionID string
	TotalStake   string
	Status       int
	ItemsJSON    []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CreateBetTicketParams struct {
	UserID       int64
	RoomCode     string
	PeriodID     int64
	RequestID    string
	ConnectionID string
	TotalStake   string
	Items        []BetTicketItemRecord
	PlacedIP     string
	PlacedDevice string
}

type SettleBetTicketRecord struct {
	ID         int64
	UserID     int64
	WalletID   int64
	TotalStake string
	ItemsJSON  []byte
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) ListRooms(ctx context.Context) ([]GameRoomRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		select code, game_type, duration_seconds, bet_cutoff_seconds, status, sort_order
		from game_rooms
		where status = 1
		order by sort_order asc, id asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]GameRoomRecord, 0)
	for rows.Next() {
		var item GameRoomRecord
		if err := rows.Scan(&item.Code, &item.GameType, &item.DurationSeconds, &item.BetCutoffSeconds, &item.Status, &item.SortOrder); err != nil {
			return nil, err
		}
		rooms = append(rooms, item)
	}

	return rooms, rows.Err()
}

func (r *GameRepository) FindRoomByCode(ctx context.Context, roomCode string) (GameRoomRecord, error) {
	var room GameRoomRecord
	err := r.db.QueryRowContext(ctx, `
		select code, game_type, duration_seconds, bet_cutoff_seconds, status, sort_order
		from game_rooms
		where code = $1 and status = 1
		limit 1
	`, roomCode).Scan(&room.Code, &room.GameType, &room.DurationSeconds, &room.BetCutoffSeconds, &room.Status, &room.SortOrder)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GameRoomRecord{}, ErrGameRoomNotFound
		}
		return GameRoomRecord{}, err
	}

	return room, nil
}

func (r *GameRepository) GetCurrentPeriodByRoom(ctx context.Context, roomCode string) (GamePeriodRecord, error) {
	var period GamePeriodRecord
	err := r.db.QueryRowContext(ctx, `
		select id, room_code, game_type, period_no, open_at, bet_lock_at, draw_at, status
		from game_periods
		where room_code = $1 and status in (1, 2, 3, 4)
		order by draw_at asc
		limit 1
	`, roomCode).Scan(
		&period.ID,
		&period.RoomCode,
		&period.GameType,
		&period.PeriodNo,
		&period.OpenAt,
		&period.BetLockAt,
		&period.DrawAt,
		&period.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GamePeriodRecord{}, ErrPeriodNotFound
		}
		return GamePeriodRecord{}, err
	}

	return period, nil
}

func (r *GameRepository) ListRoomRounds(ctx context.Context, roomCode string, page, pageSize int) ([]GameRoundRecord, int, error) {
	page, pageSize = normalizePagination(page, pageSize)

	var total int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from game_round_histories
		where room_code = $1
	`, roomCode).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, game_type, room_code, period_no, result, big_small, color, draw_at, status, created_at, updated_at
		from game_round_histories
		where room_code = $1
		order by draw_at desc, id desc
		limit $2 offset $3
	`, roomCode, pageSize, (page-1)*pageSize)
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
			&record.RoomCode,
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

func (r *GameRepository) ListRoomRecentRounds(ctx context.Context, roomCode string, limit int) ([]GameRoundRecord, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, game_type, room_code, period_no, result, big_small, color, draw_at, status, created_at, updated_at
		from game_round_histories
		where room_code = $1
		order by draw_at desc, id desc
		limit $2
	`, roomCode, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]GameRoundRecord, 0)
	for rows.Next() {
		var record GameRoundRecord
		if err := rows.Scan(
			&record.ID,
			&record.GameType,
			&record.RoomCode,
			&record.PeriodNo,
			&record.Result,
			&record.BigSmall,
			&record.Color,
			&record.DrawAt,
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

func (r *GameRepository) ListGameRounds(ctx context.Context, gameType string, page, pageSize int) ([]GameRoundRecord, int, error) {
	roomCode, ok := game.DefaultRoomCode(game.GameType(gameType))
	if !ok {
		return nil, 0, ErrGameRoomNotFound
	}
	return r.ListRoomRounds(ctx, roomCode, page, pageSize)
}

func (r *GameRepository) CreateBetTicket(ctx context.Context, params CreateBetTicketParams) (BetTicketRecord, error) {
	if params.RoomCode == "" {
		return BetTicketRecord{}, ErrGameRoomNotFound
	}

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

	if params.RequestID != "" {
		existing, found, err := r.findTicketByRequestIDTx(ctx, tx, params.RequestID)
		if err != nil {
			return BetTicketRecord{}, err
		}
		if found {
			return existing, nil
		}
	}

	period, err := r.lockPeriodForBet(ctx, tx, params.RoomCode, params.PeriodID)
	if err != nil {
		return BetTicketRecord{}, err
	}

	if period.Status != periodStatusOpen {
		return BetTicketRecord{}, ErrPeriodNotOpen
	}

	now := time.Now()
	if !now.Before(period.BetLockAt) {
		return BetTicketRecord{}, ErrPeriodBetLocked
	}

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

	ticketNo := buildTicketNo()
	betType := 1
	if len(params.Items) > 1 {
		betType = 2
	}

	placedIP := strings.TrimSpace(params.PlacedIP)
	placedDevice := strings.TrimSpace(params.PlacedDevice)

	row := tx.QueryRowContext(ctx, `
		insert into bet_tickets (
			ticket_no, user_id, wallet_id, request_id, connection_id, unit, game_type,
			period_id, bet_type, stake, total_stake, total_odds, potential_payout,
			status, placed_ip, placed_device, items, created_at, updated_at
		)
		values (
			$1, $2, $3, $4, $5, 1, $6,
			$7, $8, $9::numeric(20,8), $9::numeric(20,8), 1::numeric(14,6), $9::numeric(20,8),
			$10, nullif($11, ''), nullif($12, ''), $13::jsonb, $14, $14
		)
		returning id, user_id, wallet_id, game_type, period_id, request_id, connection_id,
		          total_stake::text, status, items, created_at, updated_at
	`, ticketNo, params.UserID, walletID, nullIfEmpty(params.RequestID), nullIfEmpty(params.ConnectionID), period.GameType, period.ID, betType, stakeAmount.String(), betStatusPending, placedIP, placedDevice, itemsJSON, now)

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
		if params.RequestID != "" && isUniqueViolation(err) {
			existing, found, findErr := r.findTicketByRequestIDTx(ctx, tx, params.RequestID)
			if findErr != nil {
				return BetTicketRecord{}, findErr
			}
			if found {
				return existing, nil
			}
		}
		return BetTicketRecord{}, err
	}
	record.PeriodNo = period.PeriodNo

	for _, item := range params.Items {
		optionTypeValue := mapOptionType(item.OptionType)
		stakeValue, err := parseNumeric(item.Stake)
		if err != nil {
			return BetTicketRecord{}, err
		}
		if _, err := tx.ExecContext(ctx, `
			insert into bet_items (
				ticket_id, period_id, option_type, option_key, option_label,
				odds_at_placement, stake, result, created_at, updated_at
			)
			values (
				$1, $2, $3, $4, $5,
				1::numeric(12,4), $6::numeric(20,8), 1, $7, $7
			)
		`, record.ID, period.ID, optionTypeValue, strings.TrimSpace(item.OptionKey), strings.TrimSpace(item.OptionKey), stakeValue.String(), now); err != nil {
			return BetTicketRecord{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (
			wallet_id, user_id, direction, amount, balance_before, balance_after,
			reference_type, reference_id, note, created_at
		)
		values ($1, $2, $3, $4::numeric(20,8), $5::numeric(20,8), $6::numeric(20,8),
		        $7, $8, $9, now())
	`, walletID, params.UserID, 2, stakeAmount.String(), availableBefore, availableAfter, "App\\Models\\Bet\\BetTicket", record.ID, "Khóa tiền khi đặt cược"); err != nil {
		return BetTicketRecord{}, err
	}

	if err := tx.Commit(); err != nil {
		return BetTicketRecord{}, err
	}

	return record, nil
}

func (r *GameRepository) ListRoomBetTickets(ctx context.Context, userID int64, roomCode string, page, pageSize int) ([]BetTicketRecord, int, error) {
	page, pageSize = normalizePagination(page, pageSize)

	var total int
	if err := r.db.QueryRowContext(ctx, `
		select count(*)
		from bet_tickets t
		inner join game_periods p on p.id = t.period_id
		where t.user_id = $1 and p.room_code = $2
	`, userID, roomCode).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		select t.id, t.user_id, t.wallet_id, t.game_type, t.period_id, p.period_no,
		       t.request_id, t.connection_id, coalesce(t.total_stake, t.stake)::text, t.status,
		       t.items, t.created_at, t.updated_at
		from bet_tickets t
		inner join game_periods p on p.id = t.period_id
		where t.user_id = $1 and p.room_code = $2
		order by t.created_at desc, t.id desc
		limit $3 offset $4
	`, userID, roomCode, pageSize, (page-1)*pageSize)
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
			&record.PeriodNo,
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

func (r *GameRepository) ListBetTickets(ctx context.Context, userID int64, gameType string, page, pageSize int) ([]BetTicketRecord, int, error) {
	roomCode, ok := game.DefaultRoomCode(game.GameType(gameType))
	if !ok {
		return nil, 0, ErrGameRoomNotFound
	}
	return r.ListRoomBetTickets(ctx, userID, roomCode, page, pageSize)
}

func (r *GameRepository) lockPeriodForBet(ctx context.Context, tx *sql.Tx, roomCode string, periodID int64) (GamePeriodRecord, error) {
	var period GamePeriodRecord
	err := tx.QueryRowContext(ctx, `
		select id, room_code, game_type, period_no, open_at, bet_lock_at, draw_at, status
		from game_periods
		where id = $1 and room_code = $2
		limit 1
		for update
	`, periodID, roomCode).Scan(
		&period.ID,
		&period.RoomCode,
		&period.GameType,
		&period.PeriodNo,
		&period.OpenAt,
		&period.BetLockAt,
		&period.DrawAt,
		&period.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GamePeriodRecord{}, ErrPeriodNotFound
		}
		return GamePeriodRecord{}, err
	}

	return period, nil
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

func (r *GameRepository) findTicketByRequestIDTx(ctx context.Context, tx *sql.Tx, requestID string) (BetTicketRecord, bool, error) {
	row := tx.QueryRowContext(ctx, `
		select t.id, t.user_id, t.wallet_id, t.game_type, t.period_id, p.period_no,
		       t.request_id, t.connection_id, coalesce(t.total_stake, t.stake)::text, t.status,
		       t.items, t.created_at, t.updated_at
		from bet_tickets t
		left join game_periods p on p.id = t.period_id
		where t.request_id = $1
		limit 1
	`, requestID)

	var record BetTicketRecord
	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.WalletID,
		&record.GameType,
		&record.PeriodID,
		&record.PeriodNo,
		&record.RequestID,
		&record.ConnectionID,
		&record.TotalStake,
		&record.Status,
		&record.ItemsJSON,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return BetTicketRecord{}, false, nil
		}
		return BetTicketRecord{}, false, err
	}

	return record, true, nil
}

func mapOptionType(optionType string) int {
	switch strings.ToUpper(strings.TrimSpace(optionType)) {
	case "NUMBER":
		return 1
	case "BIG_SMALL":
		return 2
	case "ODD_EVEN":
		return 3
	case "COLOR":
		return 4
	case "SUM":
		return 5
	case "COMBINATION":
		return 6
	default:
		return 1
	}
}

func buildTicketNo() string {
	now := time.Now().Format("20060102150405")
	suffix := strings.ToUpper(id.New())
	if len(suffix) > 20 {
		suffix = suffix[:20]
	}
	return fmt.Sprintf("BT%s%s", now, suffix)
}

func nullIfEmpty(value string) any {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
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

func ParsePeriodID(value string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, ErrPeriodNotFound
	}
	return parsed, nil
}
