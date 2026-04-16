package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gin/internal/domain/game"
	"gin/internal/support/clock"
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
	ID               int64
	RoomCode         string
	GameType         int
	PeriodNo         string
	PeriodIndex      int64
	OpenAt           time.Time
	BetLockAt        time.Time
	DrawAt           time.Time
	Status           int
	ManualResultJSON []byte
}

type GameRoundRecord struct {
	ID        int64
	GameType  string
	RoomCode  string
	PeriodNo  string
	PeriodIndex int64
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
	ID             int64
	UserID         int64
	WalletID       int64
	GameType       int
	PeriodID       int64
	PeriodNo       string
	PeriodIndex    int64
	RequestID      string
	ConnectionID   string
	TotalStake     string
	OriginalAmount string
	TaxAmount      string
	NetAmount      string
	ActualPayout   string
	ProfitLoss     string
	Status         int
	ItemsJSON      []byte
	CreatedAt      time.Time
	SettledAt      sql.NullTime
	UpdatedAt      time.Time
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
	ID             int64
	UserID         int64
	WalletID       int64
	OriginalAmount string
	TaxAmount      string
	NetAmount      string
	TotalStake     string
	ItemsJSON      []byte
}

type PeriodBetStats struct {
	OptionKey   string `json:"option_key"`
	OptionType  int    `json:"option_type"`
	PlayerCount int    `json:"player_count"`
	TotalStake  string `json:"total_stake"`
}

func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) GetPeriodBetStats(ctx context.Context, periodID int64) ([]PeriodBetStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		select option_key, option_type, count(distinct user_id) as player_count, sum(stake)::text as total_stake
		from bet_items i
		inner join bet_tickets t on t.id = i.ticket_id
		where i.period_id = $1
		group by option_key, option_type
		order by total_stake desc
	`, periodID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make([]PeriodBetStats, 0)
	for rows.Next() {
		var item PeriodBetStats
		if err := rows.Scan(&item.OptionKey, &item.OptionType, &item.PlayerCount, &item.TotalStake); err != nil {
			return nil, err
		}
		stats = append(stats, item)
	}
	return stats, rows.Err()
}

func (r *GameRepository) SetPeriodManualResult(ctx context.Context, periodID int64, resultJSON []byte) error {
	result, err := r.db.ExecContext(ctx, `
		update game_periods
		set manual_result = $1,
		    updated_at = now()
		where id = $2
		  and status in (1, 2)
		  and now() < bet_lock_at
	`, resultJSON, periodID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrPeriodBetLocked
	}
	return nil
}

func (r *GameRepository) ListAllRoomsWithCurrentPeriod(ctx context.Context) ([]struct {
	Room   GameRoomRecord
	Period *GamePeriodRecord
}, error) {
	rooms, err := r.ListRooms(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]struct {
		Room   GameRoomRecord
		Period *GamePeriodRecord
	}, 0, len(rooms))

	for _, room := range rooms {
		period, err := r.GetCurrentPeriodByRoom(ctx, room.Code)
		if errors.Is(err, ErrPeriodNotFound) {
			period, err = r.GetNearestUpcomingPeriodByRoom(ctx, room.Code)
		}
		if err != nil && !errors.Is(err, ErrPeriodNotFound) {
			return nil, err
		}
		var pPtr *GamePeriodRecord
		if err == nil {
			pPtr = &period
		}
		result = append(result, struct {
			Room   GameRoomRecord
			Period *GamePeriodRecord
		}{Room: room, Period: pPtr})
	}

	return result, nil
}

func (r *GameRepository) GetNearestUpcomingPeriodByRoom(ctx context.Context, roomCode string) (GamePeriodRecord, error) {
	now := clock.Now()
	var period GamePeriodRecord
	err := r.db.QueryRowContext(ctx, `
		select id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
		from game_periods
		where room_code = $1
		  and status in ($2, $3, $4)
		  and draw_at > $5
		order by draw_at asc, open_at asc, id asc
		limit 1
	`, roomCode, periodStatusScheduled, periodStatusOpen, periodStatusLocked, now).Scan(
		&period.ID,
		&period.RoomCode,
		&period.GameType,
		&period.PeriodNo,
		&period.PeriodIndex,
		&period.OpenAt,
		&period.BetLockAt,
		&period.DrawAt,
		&period.Status,
		&period.ManualResultJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GamePeriodRecord{}, ErrPeriodNotFound
		}
		return GamePeriodRecord{}, err
	}

	return period, nil
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
	now := clock.Now()
	roomDurationSeconds := 60
	if err := r.db.QueryRowContext(ctx, `
		select duration_seconds
		from game_rooms
		where code = $1
		limit 1
	`, roomCode).Scan(&roomDurationSeconds); err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[engine][period.current.lookup.error] room_code=%s stage=load_room_duration err=%v", roomCode, err)
		return GamePeriodRecord{}, err
	}
	if roomDurationSeconds <= 0 {
		roomDurationSeconds = 60
	}
	maxFuture := now.Add(time.Duration(roomDurationSeconds*3) * time.Second)

	scanPeriod := func(query string, args ...any) (GamePeriodRecord, error) {
		var period GamePeriodRecord
		err := r.db.QueryRowContext(ctx, query, args...).Scan(
			&period.ID,
			&period.RoomCode,
			&period.GameType,
			&period.PeriodNo,
			&period.PeriodIndex,
			&period.OpenAt,
			&period.BetLockAt,
			&period.DrawAt,
			&period.Status,
			&period.ManualResultJSON,
		)
		if err != nil {
			return GamePeriodRecord{}, err
		}
		return period, nil
	}

	current, err := scanPeriod(`
		select id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
		from game_periods
		where room_code = $1
		  and status in ($2, $3, $4)
		  and open_at <= $5
		  and draw_at > $5
		  and draw_at <= $6
		order by draw_at asc, open_at desc, id asc
		limit 1
	`, roomCode, periodStatusOpen, periodStatusLocked, periodStatusScheduled, now, maxFuture)
	if err == nil {
		return current, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[engine][period.current.lookup.error] room_code=%s stage=current err=%v", roomCode, err)
		return GamePeriodRecord{}, err
	}

	return GamePeriodRecord{}, ErrPeriodNotFound
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
		select h.id, h.game_type, h.room_code, h.period_no, coalesce(p.period_index, 0) as period_index,
		       h.result, h.big_small, h.color, h.draw_at, h.status, h.created_at, h.updated_at
		from game_round_histories h
		left join game_periods p
			on p.room_code = h.room_code
		   and p.period_no = h.period_no
		where h.room_code = $1
		order by h.draw_at desc, h.id desc
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
			&record.PeriodIndex,
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
		select h.id, h.game_type, h.room_code, h.period_no, coalesce(p.period_index, 0) as period_index,
		       h.result, h.big_small, h.color, h.draw_at, h.status, h.created_at, h.updated_at
		from game_round_histories h
		left join game_periods p
			on p.room_code = h.room_code
		   and p.period_no = h.period_no
		where h.room_code = $1
		order by h.draw_at desc, h.id desc
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
			&record.PeriodIndex,
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
	originalAmountDB := stakeAmount.FloatString(8)
	taxAmountDB, netAmountDB, err := calculateBetTaxAndNet(originalAmountDB)
	if err != nil {
		return BetTicketRecord{}, err
	}
	if compareNumeric(netAmountDB, "0") <= 0 {
		return BetTicketRecord{}, fmt.Errorf("số tiền cược sau thuế không hợp lệ")
	}
	potentialPayoutDB, err := estimateBetPotentialPayout(params.Items, taxAmountDB)
	if err != nil {
		return BetTicketRecord{}, err
	}
	log.Printf("[engine][bet.tax] room_code=%s original_amount=%s tax_amount=%s net_amount=%s potential_payout=%s", params.RoomCode, originalAmountDB, taxAmountDB, netAmountDB, potentialPayoutDB)

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

	now := clock.Now()

	period, err := r.lockPeriodForBet(ctx, tx, params.RoomCode, params.PeriodID)
	if err != nil {
		if !errors.Is(err, ErrPeriodNotFound) {
			return BetTicketRecord{}, err
		}

		// Production-safe fallback:
		// the client may still hold the previous period_id for a brief moment while
		// the room has already rolled to the next OPEN period. In that case, accept
		// the latest valid OPEN period instead of failing hard with PeriodNotFound.
		fallbackPeriod, fallbackErr := r.lockCurrentOpenPeriodForBet(ctx, tx, params.RoomCode, now)
		if fallbackErr != nil {
			return BetTicketRecord{}, err
		}
		log.Printf(
			"[play.bet.period.fallback] room_code=%s requested_period_id=%d actual_period_id=%d actual_period_no=%s reason=period_not_found",
			params.RoomCode,
			params.PeriodID,
			fallbackPeriod.ID,
			fallbackPeriod.PeriodNo,
		)
		period = fallbackPeriod
	}

	if period.Status != periodStatusOpen || !now.Before(period.BetLockAt) {
		fallbackPeriod, fallbackErr := r.lockCurrentOpenPeriodForBet(ctx, tx, params.RoomCode, now)
		if fallbackErr == nil {
			if fallbackPeriod.ID != period.ID {
				log.Printf(
					"[play.bet.period.fallback] room_code=%s requested_period_id=%d actual_period_id=%d actual_period_no=%s reason=requested_period_unavailable",
					params.RoomCode,
					params.PeriodID,
					fallbackPeriod.ID,
					fallbackPeriod.PeriodNo,
				)
			}
			period = fallbackPeriod
		}
	}

	if period.Status != periodStatusOpen {
		return BetTicketRecord{}, ErrPeriodNotOpen
	}
	if !now.Before(period.BetLockAt) {
		return BetTicketRecord{}, ErrPeriodBetLocked
	}

	walletID, balanceBefore, _, err := r.lockWalletForBet(ctx, tx, params.UserID)
	if err != nil {
		return BetTicketRecord{}, err
	}

	if compareNumeric(balanceBefore, originalAmountDB) < 0 {
		return BetTicketRecord{}, ErrInsufficientBetBalance
	}

	balanceAfter, err := subtractNumeric(balanceBefore, originalAmountDB)
	if err != nil {
		return BetTicketRecord{}, err
	}


	ticketNo := buildTicketNo()
	betType := 1
	if len(params.Items) > 1 {
		betType = 2
	}

	placedIP := trimToVarchar(strings.TrimSpace(params.PlacedIP), 45)
	placedDevice := trimToVarchar(strings.TrimSpace(params.PlacedDevice), 100)

	row := tx.QueryRowContext(ctx, `
		insert into bet_tickets (
			ticket_no, user_id, wallet_id, request_id, connection_id, unit, game_type,
			period_id, bet_type, stake, total_stake, original_amount, tax_amount, net_amount, total_odds, potential_payout,
			status, placed_ip, placed_device, items, created_at, updated_at
		)
		values (
			$1, $2, $3, $4, $5, 1, $6,
			$7, $8, $9::numeric(20,8), $9::numeric(20,8), $9::numeric(20,8), $10::numeric(20,8), $11::numeric(20,8), 1::numeric(14,6), $12::numeric(20,8),
			$13, nullif($14, ''), nullif($15, ''), $16::jsonb, $17, $17
		)
		returning id, user_id, wallet_id, game_type, period_id, request_id, connection_id,
		          total_stake::text, original_amount::text, tax_amount::text, net_amount::text,
		          coalesce(actual_payout, 0)::text, status, items, created_at, updated_at
	`, ticketNo, params.UserID, walletID, nullIfEmpty(params.RequestID), nullIfEmpty(params.ConnectionID), period.GameType, period.ID, betType, originalAmountDB, taxAmountDB, netAmountDB, potentialPayoutDB, betStatusPending, placedIP, placedDevice, itemsJSON, now)

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
		&record.OriginalAmount,
		&record.TaxAmount,
		&record.NetAmount,
		&record.ActualPayout,
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
	record.PeriodIndex = period.PeriodIndex
	if record.OriginalAmount == "" {
		record.OriginalAmount = originalAmountDB
	}
	if record.TaxAmount == "" {
		record.TaxAmount = taxAmountDB
	}
	if record.NetAmount == "" {
		record.NetAmount = netAmountDB
	}
	if record.TotalStake == "" {
		record.TotalStake = originalAmountDB
	}

	if _, err := tx.ExecContext(ctx, `
		update wallets
		set balance = balance - $1::numeric(20,8),
		    locked_balance = locked_balance + $1::numeric(20,8),
		    updated_at = now()
		where id = $2
	`, params.TotalStake, walletID); err != nil {
		return BetTicketRecord{}, err
	}

	for _, item := range params.Items {
		optionTypeValue := mapOptionType(item.OptionType)
		stakeValue, err := parseNumeric(item.Stake)
		if err != nil {
			return BetTicketRecord{}, err
		}
		stakeValueDB := stakeValue.FloatString(8)
		optionKey := trimToVarchar(strings.TrimSpace(item.OptionKey), 100)
		optionLabel := trimToVarchar(strings.TrimSpace(item.OptionKey), 150)
		if _, err := tx.ExecContext(ctx, `
			insert into bet_items (
				ticket_id, period_id, option_type, option_key, option_label,
				odds_at_placement, stake, result, created_at, updated_at
			)
			values (
				$1, $2, $3, $4, $5,
				1::numeric(12,4), $6::numeric(20,8), 1, $7, $7
			)
		`, record.ID, period.ID, optionTypeValue, optionKey, optionLabel, stakeValueDB, now); err != nil {
			return BetTicketRecord{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (
			wallet_id, user_id, direction, amount, balance_before, balance_after,
			reference_type, reference_id, note, created_at
		)
		values ($1, $2, 2, $3::numeric(20,8), $4::numeric(20,8), $5::numeric(20,8),
		        $6, $7, $8, now())
	`, walletID, params.UserID, originalAmountDB, balanceBefore, balanceAfter, "App\\Models\\Bet\\BetTicket", record.ID, "Giảm số dư khả dụng khi đặt cược (khóa tiền)"); err != nil {
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
		select t.id, t.user_id, t.wallet_id, t.game_type, t.period_id, p.period_no, coalesce(p.period_index, 0),
		       t.request_id, t.connection_id, coalesce(t.total_stake, t.stake)::text,
		       coalesce(t.original_amount, t.total_stake, t.stake)::text,
		       coalesce(t.tax_amount, 0)::text,
		       coalesce(t.net_amount, t.total_stake, t.stake)::text,
		       coalesce(t.actual_payout, 0)::text,
		       (
		           case
		               when t.status = 2 then coalesce(t.actual_payout, 0) - coalesce(t.original_amount, t.total_stake, t.stake)
		               else 0
		           end
		       )::text,
		       t.status, t.items, t.created_at, t.settled_at, t.updated_at
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
			&record.PeriodIndex,
			&record.RequestID,
			&record.ConnectionID,
			&record.TotalStake,
			&record.OriginalAmount,
			&record.TaxAmount,
			&record.NetAmount,
			&record.ActualPayout,
			&record.ProfitLoss,
			&record.Status,
			&record.ItemsJSON,
			&record.CreatedAt,
			&record.SettledAt,
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
		select id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
		from game_periods
		where id = $1 and room_code = $2
		limit 1
		for update
	`, periodID, roomCode).Scan(
		&period.ID,
		&period.RoomCode,
		&period.GameType,
		&period.PeriodNo,
		&period.PeriodIndex,
		&period.OpenAt,
		&period.BetLockAt,
		&period.DrawAt,
		&period.Status,
		&period.ManualResultJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GamePeriodRecord{}, ErrPeriodNotFound
		}
		return GamePeriodRecord{}, err
	}

	return period, nil
}

func (r *GameRepository) lockCurrentOpenPeriodForBet(ctx context.Context, tx *sql.Tx, roomCode string, now time.Time) (GamePeriodRecord, error) {
	var period GamePeriodRecord
	err := tx.QueryRowContext(ctx, `
		select id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
		from game_periods
		where room_code = $1
		  and status = $2
		  and open_at <= $3
		  and $3 < bet_lock_at
		order by draw_at asc, open_at desc, id asc
		limit 1
		for update
	`, roomCode, periodStatusOpen, now).Scan(
		&period.ID,
		&period.RoomCode,
		&period.GameType,
		&period.PeriodNo,
		&period.PeriodIndex,
		&period.OpenAt,
		&period.BetLockAt,
		&period.DrawAt,
		&period.Status,
		&period.ManualResultJSON,
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
	var status int
	row := tx.QueryRowContext(ctx, `
		select id, balance::text, locked_balance::text, status
		from wallets
		where user_id = $1 and unit = $2
		limit 1
		for update
	`, userID, 1)

	if err := row.Scan(&walletID, &balanceBefore, &lockedBefore, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", "", ErrInsufficientBetBalance
		}
		return 0, "", "", err
	}

	if status != 1 {
		return 0, "", "", fmt.Errorf("ví đang bị khóa (status=%d)", status)
	}

	return walletID, balanceBefore, lockedBefore, nil
}

func (r *GameRepository) findTicketByRequestIDTx(ctx context.Context, tx *sql.Tx, requestID string) (BetTicketRecord, bool, error) {
	row := tx.QueryRowContext(ctx, `
		select t.id, t.user_id, t.wallet_id, t.game_type, t.period_id, p.period_no, coalesce(p.period_index, 0),
		       t.request_id, t.connection_id, coalesce(t.total_stake, t.stake)::text,
		       coalesce(t.original_amount, t.total_stake, t.stake)::text,
		       coalesce(t.tax_amount, 0)::text,
		       coalesce(t.net_amount, t.total_stake, t.stake)::text,
		       coalesce(t.actual_payout, 0)::text,
		       (
		           case
		               when t.status = 2 then coalesce(t.actual_payout, 0) - coalesce(t.original_amount, t.total_stake, t.stake)
		               else 0
		           end
		       )::text,
		       t.status, t.items, t.created_at, t.settled_at, t.updated_at
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
		&record.PeriodIndex,
		&record.RequestID,
		&record.ConnectionID,
		&record.TotalStake,
		&record.OriginalAmount,
		&record.TaxAmount,
		&record.NetAmount,
		&record.ActualPayout,
		&record.ProfitLoss,
		&record.Status,
		&record.ItemsJSON,
		&record.CreatedAt,
		&record.SettledAt,
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

func settlementOddsForItem(optionType, optionKey string) *big.Rat {
	key := strings.ToLower(strings.TrimSpace(optionKey))
	typ := strings.ToUpper(strings.TrimSpace(optionType))

	switch {
	case strings.HasPrefix(key, "number_"), strings.HasPrefix(key, "digit_"), strings.HasPrefix(key, "last_"):
		return big.NewRat(9, 1)
	case key == "violet":
		return big.NewRat(9, 2)
	case key == "green", key == "red", key == "big", key == "small", key == "odd", key == "even":
		return big.NewRat(2, 1)
	case strings.HasPrefix(key, "pair_"):
		return big.NewRat(1383, 100)
	case strings.HasPrefix(key, "sspair_"):
		return big.NewRat(1728, 25)
	case strings.HasPrefix(key, "triple_"):
		return big.NewRat(5184, 25)
	case key == "serial_any":
		return big.NewRat(216, 25)
	case strings.HasPrefix(key, "diff_"):
		return big.NewRat(864, 25)
	case strings.HasPrefix(key, "sum_"):
		if odds := sumOddsForKey(key); odds != nil {
			return odds
		}
	}

	switch typ {
	case "BIG_SMALL", "ODD_EVEN":
		return big.NewRat(2, 1)
	case "COLOR":
		if key == "violet" {
			return big.NewRat(9, 2)
		}
		return big.NewRat(2, 1)
	case "NUMBER":
		return big.NewRat(9, 1)
	case "SUM":
		if odds := sumOddsForKey(key); odds != nil {
			return odds
		}
		return big.NewRat(2, 1)
	case "COMBINATION":
		if strings.HasPrefix(key, "pair_") {
			return big.NewRat(1383, 100)
		}
		if strings.HasPrefix(key, "sspair_") {
			return big.NewRat(1728, 25)
		}
		if strings.HasPrefix(key, "triple_") {
			return big.NewRat(5184, 25)
		}
		if key == "serial_any" {
			return big.NewRat(216, 25)
		}
		if strings.HasPrefix(key, "diff_") {
			return big.NewRat(864, 25)
		}
	}

	return big.NewRat(2, 1)
}

func sumOddsForKey(key string) *big.Rat {
	switch key {
	case "sum_3", "sum_18":
		return big.NewRat(5184, 25)
	case "sum_4", "sum_17":
		return big.NewRat(1728, 25)
	case "sum_5", "sum_16":
		return big.NewRat(864, 25)
	case "sum_6", "sum_15", "sum_30":
		return big.NewRat(1037, 50)
	case "sum_7", "sum_14":
		return big.NewRat(1383, 100)
	case "sum_8", "sum_13":
		return big.NewRat(247, 25)
	case "sum_9", "sum_12":
		return big.NewRat(83, 10)
	case "sum_10", "sum_11":
		return big.NewRat(192, 25)
	default:
		return nil
	}
}

func estimateBetPotentialPayout(items []BetTicketItemRecord, taxAmount string) (string, error) {
	total := new(big.Rat)
	for _, item := range items {
		stake, err := parseNumeric(item.Stake)
		if err != nil {
			return "", err
		}
		odds := settlementOddsForItem(item.OptionType, item.OptionKey)
		if odds.Sign() <= 0 {
			continue
		}
		itemPayout := new(big.Rat).Mul(stake, odds)
		total.Add(total, itemPayout)
	}

	if strings.TrimSpace(taxAmount) != "" {
		taxRat, err := parseNumeric(taxAmount)
		if err != nil {
			return "", err
		}
		if taxRat.Sign() > 0 {
			total.Sub(total, taxRat)
		}
	}

	if total.Sign() < 0 {
		total.SetInt64(0)
	}

	return total.FloatString(8), nil
}

func buildTicketNo() string {
	now := clock.Now().Format("20060102150405")
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

func trimToVarchar(value string, maxChars int) string {
	trimmed := strings.TrimSpace(value)
	if maxChars <= 0 || trimmed == "" {
		return ""
	}
	runes := []rune(trimmed)
	if len(runes) <= maxChars {
		return trimmed
	}
	return string(runes[:maxChars])
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

func ParsePeriodID(value string) (int64, error) {
	parsed, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed <= 0 {
		return 0, ErrPeriodNotFound
	}
	return parsed, nil
}
