package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"log"
	"math/big"
	"strings"
	"time"

	"gin/internal/support/clock"
)

const periodGenerationBufferCount = 3

type DrawResult struct {
	Result      string
	BigSmall    string
	Color       string
	PayloadJSON []byte
}

type GamePeriodSettlementRecord struct {
	ID            int64
	RoomCode      string
	PeriodNo      string
	GameType      int
	ResultPayload []byte
	DrawAt        time.Time
}

type ticketSettleItemOutcome struct {
	OptionType string
	OptionKey  string
	Stake      string
	IsWin      bool
	Payout     string
}

func (r *GameRepository) EnsureRoomPeriods(ctx context.Context, room GameRoomRecord, now time.Time) ([]GamePeriodRecord, error) {
	nowVN := now.In(clock.Location())

	duration := time.Duration(room.DurationSeconds) * time.Second
	if duration <= 0 {
		err := fmt.Errorf("room duration không hợp lệ: room_code=%s duration_seconds=%d", room.Code, room.DurationSeconds)
		log.Printf("[engine][period.ensure.error] room_code=%s stage=invalid_duration err=%v", room.Code, err)
		return nil, err
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=begin_tx err=%v", room.Code, err)
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	created := make([]GamePeriodRecord, 0, periodGenerationBufferCount)

	var latestDrawAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		select draw_at
		from game_periods
		where room_code = $1
		order by draw_at desc
		limit 1
		for update
	`, room.Code).Scan(&latestDrawAt)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=latest_draw_at err=%v", room.Code, err)
		return nil, err
	}
	firstDrawAt := alignNextDrawAt(nowVN, duration)

	if !latestDrawAt.Valid {
		// Log bootstrap only
		log.Printf("[engine][period.ensure.state] room_code=%s action=bootstrap_initial_period", room.Code)
	}

	for offset := 0; offset < periodGenerationBufferCount; offset += 1 {
		drawAt := firstDrawAt.Add(time.Duration(offset) * duration)
		openAt := drawAt.Add(-duration)
		status := periodStatusScheduled
		if openAt.Equal(nowVN) || openAt.Before(nowVN) {
			status = periodStatusOpen
		}

		record, inserted, insertErr := r.insertPeriodTx(ctx, tx, room, openAt, drawAt, status)
		if insertErr != nil {
			log.Printf(
				"[engine][period.ensure.error] room_code=%s stage=insert_buffer offset=%d open_at=%s draw_at=%s err=%v",
				room.Code,
				offset,
				openAt.Format(time.RFC3339Nano),
				drawAt.Format(time.RFC3339Nano),
				insertErr,
			)
			return nil, insertErr
		}
		if inserted {
			created = append(created, record)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=commit err=%v", room.Code, err)
		return nil, err
	}
	return created, nil
}

func alignNextDrawAt(now time.Time, duration time.Duration) time.Time {
	if duration <= 0 {
		return now
	}

	base := now.Truncate(duration)
	next := base.Add(duration)
	if !next.After(now) {
		next = next.Add(duration)
	}

	return next
}

func (r *GameRepository) MoveScheduledToOpen(ctx context.Context, now time.Time) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		with moved as (
			update game_periods
			set status = $1,
			    updated_at = now()
			where id in (
				select id
				from (
					select distinct on (room_code) id
					from game_periods
					where status = $2
					  and open_at <= $3
					order by room_code asc, open_at asc, id asc
				) due_periods
			)
			returning room_code
		)
		select distinct room_code
		from moved
		order by room_code asc
	`, periodStatusOpen, periodStatusScheduled, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roomCodes := make([]string, 0)
	for rows.Next() {
		var roomCode string
		if err := rows.Scan(&roomCode); err != nil {
			return nil, err
		}
		roomCodes = append(roomCodes, roomCode)
	}
	return roomCodes, rows.Err()
}

func (r *GameRepository) MoveOpenToLocked(ctx context.Context, now time.Time) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		with moved as (
			update game_periods
			set status = $1,
			    updated_at = now()
			where id in (
				select id
				from (
					select distinct on (room_code) id
					from game_periods
					where status = $2
					  and bet_lock_at <= $3
					order by room_code asc, bet_lock_at asc, id asc
				) due_periods
			)
			returning room_code
		)
		select distinct room_code
		from moved
		order by room_code asc
	`, periodStatusLocked, periodStatusOpen, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roomCodes := make([]string, 0)
	for rows.Next() {
		var roomCode string
		if err := rows.Scan(&roomCode); err != nil {
			return nil, err
		}
		roomCodes = append(roomCodes, roomCode)
	}
	return roomCodes, rows.Err()
}

func (r *GameRepository) ListLockedPeriodsForDraw(ctx context.Context, now time.Time, limit int) ([]GamePeriodRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
		from (
			select distinct on (room_code) id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
			from game_periods
			where status = $1
			  and draw_at <= $2
			order by room_code asc, draw_at asc, id asc
		) due_periods
		order by draw_at asc, id asc
		limit $3
	`, periodStatusLocked, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]GamePeriodRecord, 0)
	for rows.Next() {
		var item GamePeriodRecord
		if err := rows.Scan(
			&item.ID,
			&item.RoomCode,
			&item.GameType,
			&item.PeriodNo,
			&item.PeriodIndex,
			&item.OpenAt,
			&item.BetLockAt,
			&item.DrawAt,
			&item.Status,
			&item.ManualResultJSON,
		); err != nil {
			return nil, err
		}
		records = append(records, item)
	}

	return records, rows.Err()
}

func (r *GameRepository) MarkPeriodDrawn(ctx context.Context, period GamePeriodRecord, draw DrawResult) error {

	hashBytes := sha256.Sum256(draw.PayloadJSON)
	hashValue := hex.EncodeToString(hashBytes[:])
	now := clock.Now()

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("[engine][period.draw.error] period_id=%d stage=begin_tx err=%v", period.ID, err)
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	result, err := tx.ExecContext(ctx, `
		update game_periods
		set status = $1,
		    draw_source = 1,
		    result_payload = $2::jsonb,
		    result_hash = $3,
		    updated_at = $4
		where id = $5 and status = $6
	`, periodStatusDrawn, draw.PayloadJSON, hashValue, now, period.ID, periodStatusLocked)
	if err != nil {
		log.Printf("[engine][period.draw.error] period_id=%d stage=update err=%v", period.ID, err)
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[engine][period.draw.error] period_id=%d stage=rows_affected err=%v", period.ID, err)
		return err
	}
	if affected == 0 {
		log.Printf("[engine][period.draw.skip] period_id=%d reason=not_locked_or_already_drawn", period.ID)
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		insert into game_round_histories (
			game_type, room_code, period_no, result, big_small, color, draw_at, status, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, 'DRAWN', $8, $8)
	`, toGameTypeSlug(period.GameType), period.RoomCode, period.PeriodNo, draw.Result, draw.BigSmall, draw.Color, period.DrawAt, now); err != nil {
		log.Printf("[engine][period.draw.error] period_id=%d stage=insert_round_history err=%v", period.ID, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[engine][period.draw.error] period_id=%d stage=commit err=%v", period.ID, err)
		return err
	}

	return nil
}

func (r *GameRepository) ListDrawnPeriodsForSettlement(ctx context.Context, limit int) ([]GamePeriodSettlementRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, room_code, period_no, game_type, result_payload, draw_at
		from (
			select distinct on (room_code) id, room_code, period_no, game_type, result_payload, draw_at
			from game_periods
			where status = $1
			  and result_payload is not null
			order by room_code asc, draw_at asc, id asc
		) due_periods
		order by draw_at asc, id asc
		limit $2
	`, periodStatusDrawn, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	periods := make([]GamePeriodSettlementRecord, 0)
	for rows.Next() {
		var item GamePeriodSettlementRecord
		if err := rows.Scan(&item.ID, &item.RoomCode, &item.PeriodNo, &item.GameType, &item.ResultPayload, &item.DrawAt); err != nil {
			return nil, err
		}
		periods = append(periods, item)
	}

	return periods, rows.Err()
}

func (r *GameRepository) SettlePeriod(ctx context.Context, period GamePeriodSettlementRecord) ([]int64, error) {

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=begin_tx err=%v", period.ID, err)
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var currentStatus int
	if err := tx.QueryRowContext(ctx, `
		select status
		from game_periods
		where id = $1
		for update
	`, period.ID).Scan(&currentStatus); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=lock_status err=%v", period.ID, err)
		return nil, err
	}
	if currentStatus != periodStatusDrawn {
		log.Printf("[engine][period.settle.skip] period_id=%d current_status=%d", period.ID, currentStatus)
		return nil, nil
	}

	tags, err := decodeResultTags(period.ResultPayload)
	if err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=decode_tags err=%v", period.ID, err)
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		select id, user_id, wallet_id,
		       coalesce(original_amount, total_stake, stake)::text,
		       coalesce(tax_amount, 0)::text,
		       coalesce(net_amount, total_stake, stake)::text,
		       coalesce(total_stake, stake)::text,
		       items
		from bet_tickets
		where period_id = $1 and status = $2
		order by id asc
		for update
	`, period.ID, betStatusPending)
	if err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=load_tickets err=%v", period.ID, err)
		return nil, err
	}
	defer rows.Close()

	tickets := make([]SettleBetTicketRecord, 0)
	for rows.Next() {
		var ticket SettleBetTicketRecord
		if err := rows.Scan(&ticket.ID, &ticket.UserID, &ticket.WalletID, &ticket.OriginalAmount, &ticket.TaxAmount, &ticket.NetAmount, &ticket.TotalStake, &ticket.ItemsJSON); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d stage=scan_ticket err=%v", period.ID, err)
			return nil, err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=rows_err err=%v", period.ID, err)
		return nil, err
	}

	now := clock.Now()
	for _, ticket := range tickets {
		originalAmount := ticket.OriginalAmount
		if strings.TrimSpace(originalAmount) == "" {
			originalAmount = ticket.TotalStake
		}
		taxAmount := strings.TrimSpace(ticket.TaxAmount)
		if taxAmount == "" {
			computedTax, computedNet, calcErr := calculateBetTaxAndNet(originalAmount)
			if calcErr == nil {
				taxAmount = computedTax
				if strings.TrimSpace(ticket.NetAmount) == "" {
					ticket.NetAmount = computedNet
				}
			} else {
				taxAmount = "0"
			}
		}
		netAmount := ticket.NetAmount
		if strings.TrimSpace(netAmount) == "" {
			if strings.TrimSpace(taxAmount) != "" && compareNumeric(taxAmount, "0") > 0 {
				afterTax, subErr := subtractNumeric(originalAmount, taxAmount)
				if subErr == nil && compareNumeric(afterTax, "0") > 0 {
					netAmount = afterTax
				}
			}
		}
		if strings.TrimSpace(netAmount) == "" {
			netAmount = originalAmount
		}
		outcomes, payoutTotal, err := settleTicketItems(ticket.ItemsJSON, tags, originalAmount, taxAmount)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=settle_ticket_items err=%v", period.ID, ticket.ID, err)
			return nil, err
		}

		statusAfter := betStatusLost
		if compareNumeric(payoutTotal, "0") > 0 {
			statusAfter = betStatusWon
		}

		var balanceBefore, lockedBefore string
		if err := tx.QueryRowContext(ctx, `
			select balance::text, locked_balance::text
			from wallets
			where id = $1
			for update
		`, ticket.WalletID).Scan(&balanceBefore, &lockedBefore); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=load_wallet err=%v", period.ID, ticket.ID, err)
			return nil, err
		}

		lockedAfter, err := subtractNumeric(lockedBefore, originalAmount)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=subtract_locked err=%v", period.ID, ticket.ID, err)
			return nil, err
		}
		if compareNumeric(lockedAfter, "0") < 0 {
			lockedAfter = "0"
		}

		balanceAfter, err := addNumeric(balanceBefore, payoutTotal)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=add_balance err=%v", period.ID, ticket.ID, err)
			return nil, err
		}

		if _, err := tx.ExecContext(ctx, `
			update wallets
			set balance = balance + $1::numeric(20,8),
			    locked_balance = locked_balance - $2::numeric(20,8),
			    updated_at = $3
			where id = $4
		`, payoutTotal, originalAmount, now, ticket.WalletID); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=update_wallet err=%v", period.ID, ticket.ID, err)
			return nil, err
		}

		for _, outcome := range outcomes {
			resultStatus := 3
			if outcome.IsWin {
				resultStatus = 2
			}
			if _, err := tx.ExecContext(ctx, `
				update bet_items
				set result = $1,
				    payout_amount = $2::numeric(20,8),
				    settled_at = $3,
				    updated_at = $3
				where ticket_id = $4
				  and option_type = $5
				  and option_key = $6
				  and result = 1
			`, resultStatus, outcome.Payout, now, ticket.ID, mapOptionType(outcome.OptionType), strings.TrimSpace(outcome.OptionKey)); err != nil {
				log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=update_bet_item err=%v", period.ID, ticket.ID, err)
				return nil, err
			}
		}

		if _, err := tx.ExecContext(ctx, `
			update bet_tickets
			set status = $1,
			    actual_payout = $2::numeric(20,8),
			    settled_at = $3,
			    updated_at = $3
			where id = $4
		`, statusAfter, payoutTotal, now, ticket.ID); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=update_ticket err=%v", period.ID, ticket.ID, err)
			return nil, err
		}

		profitLoss, err := subtractNumeric(payoutTotal, originalAmount)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=profit_loss err=%v", period.ID, ticket.ID, err)
			return nil, err
		}
		if _, err := tx.ExecContext(ctx, `
			insert into bet_settlements (
				ticket_id, period_id, settlement_type, before_status, after_status,
				payout_amount, profit_loss, note, created_at
			)
			values ($1, $2, 1, $3, $4, $5::numeric(20,8), $6::numeric(20,8), $7, $8)
		`, ticket.ID, period.ID, betStatusPending, statusAfter, payoutTotal, profitLoss, "Engine settlement tự động", now); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=insert_settlement err=%v", period.ID, ticket.ID, err)
			return nil, err
		}

		if compareNumeric(payoutTotal, "0") > 0 {
			if _, err := tx.ExecContext(ctx, `
				insert into wallet_ledger_entries (
					wallet_id, user_id, direction, amount, balance_before, balance_after,
					reference_type, reference_id, note, created_at
				)
				values ($1, $2, 1, $3::numeric(20,8), $4::numeric(20,8), $5::numeric(20,8),
				        $6, $7, $8, $9)
			`, ticket.WalletID, ticket.UserID, payoutTotal, balanceBefore, balanceAfter, "App\\Models\\Bet\\BetTicket", ticket.ID, "Cộng tiền thắng cược", now); err != nil {
				log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=insert_ledger err=%v", period.ID, ticket.ID, err)
				return nil, err
			}
		} else {
			// Thêm lịch sử ghi nhận khi thua để minh bạch số dư
			if _, err := tx.ExecContext(ctx, `
				insert into wallet_ledger_entries (
					wallet_id, user_id, direction, amount, balance_before, balance_after,
					reference_type, reference_id, note, created_at
				)
				values ($1, $2, 3, $3::numeric(20,8), $4::numeric(20,8), $5::numeric(20,8),
				        $6, $7, $8, $9)
			`, ticket.WalletID, ticket.UserID, "0", balanceBefore, balanceAfter, "App\\Models\\Bet\\BetTicket", ticket.ID, "Giải phóng tiền cược (Thua)", now); err != nil {
				log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=insert_ledger_lost err=%v", period.ID, ticket.ID, err)
				return nil, err
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		update game_periods
		set status = $1,
		    settled_at = $2,
		    updated_at = $2
		where id = $3 and status = $4
	`, periodStatusSettled, now, period.ID, periodStatusDrawn); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=update_period err=%v", period.ID, err)
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		update game_round_histories
		set status = 'SETTLED',
		    updated_at = $1
		where room_code = $2
		  and period_no = $3
	`, now, period.RoomCode, period.PeriodNo); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=update_round_history err=%v", period.ID, err)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=commit err=%v", period.ID, err)
		return nil, err
	}

	userIDs := make([]int64, 0, len(tickets))
	seenUserIDs := make(map[int64]struct{}, len(tickets))
	for _, ticket := range tickets {
		if _, exists := seenUserIDs[ticket.UserID]; exists {
			continue
		}
		seenUserIDs[ticket.UserID] = struct{}{}
		userIDs = append(userIDs, ticket.UserID)
	}
	return userIDs, nil
}

func (r *GameRepository) insertPeriodTx(ctx context.Context, tx *sql.Tx, room GameRoomRecord, openAt, drawAt time.Time, status int) (GamePeriodRecord, bool, error) {
	lockAt := drawAt.Add(-time.Duration(room.BetCutoffSeconds) * time.Second)
	periodNo := buildPeriodNo(room.Code, drawAt)
	periodIndex, err := r.buildNextPeriodIndexTx(ctx, tx, room.Code, drawAt)
	if err != nil {
		log.Printf("[engine][period.insert.error] room_code=%s period_no=%s stage=build_period_index err=%v", room.Code, periodNo, err)
		return GamePeriodRecord{}, false, err
	}
	var record GamePeriodRecord
	err = tx.QueryRowContext(ctx, `
		insert into game_periods (
			game_type, period_no, period_index, room_code, open_at, close_at, bet_lock_at, draw_at, status, created_at, updated_at
		)
		select
			$1, $2, $3, $4, $5, $6, $6, $7, $8, now(), now()
		on conflict (room_code, period_no) do nothing
		returning id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result
	`, room.GameType, periodNo, periodIndex, room.Code, openAt, lockAt, drawAt, status).Scan(
		&record.ID,
		&record.RoomCode,
		&record.GameType,
		&record.PeriodNo,
		&record.PeriodIndex,
		&record.OpenAt,
		&record.BetLockAt,
		&record.DrawAt,
		&record.Status,
		&record.ManualResultJSON,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return GamePeriodRecord{}, false, nil
		}
		log.Printf("[engine][period.insert.error] room_code=%s period_no=%s err=%v", room.Code, periodNo, err)
		return GamePeriodRecord{}, false, err
	}

	return record, true, nil
}

func (r *GameRepository) buildNextPeriodIndexTx(ctx context.Context, tx *sql.Tx, roomCode string, drawAt time.Time) (int64, error) {
	const periodIndexYearSpan int64 = 100000000
	const periodIndexSeedOffset int64 = 10000000
	const periodIndexSeedRange int64 = 50000000

	yearPrefix := int64(drawAt.In(clock.Location()).Year())
	baseIndex := yearPrefix * periodIndexYearSpan
	maxIndex := baseIndex + (periodIndexYearSpan - 1)
	startIndex := baseIndex + periodIndexSeedOffset + (int64(crc32.ChecksumIEEE([]byte(strings.ToLower(strings.TrimSpace(roomCode))))) % periodIndexSeedRange)

	var latest sql.NullInt64
	if err := tx.QueryRowContext(ctx, `
		select max(period_index)
		from game_periods
		where room_code = $1
		  and period_index between $2 and $3
	`, roomCode, baseIndex, maxIndex).Scan(&latest); err != nil {
		return 0, err
	}

	if !latest.Valid || latest.Int64 < startIndex {
		return startIndex, nil
	}

	nextIndex := latest.Int64 + 1
	if nextIndex > maxIndex {
		return maxIndex, nil
	}

	return nextIndex, nil
}

func settleTicketItems(rawItems []byte, tags map[string]struct{}, originalAmount string, taxAmount string) ([]ticketSettleItemOutcome, string, error) {
	var items []BetTicketItemRecord
	if len(rawItems) > 0 {
		if err := json.Unmarshal(rawItems, &items); err != nil {
			return nil, "", err
		}
	}

	outcomes := make([]ticketSettleItemOutcome, 0, len(items))
	payoutTotal := "0"
	originalRat, err := parseNumeric(originalAmount)
	if err != nil {
		return nil, "", err
	}
	if originalRat.Sign() <= 0 {
		return nil, "", fmt.Errorf("invalid bet amount for settlement")
	}

	taxRat := new(big.Rat)
	if strings.TrimSpace(taxAmount) != "" {
		parsedTax, taxErr := parseNumeric(taxAmount)
		if taxErr != nil {
			return nil, "", taxErr
		}
		if parsedTax.Sign() > 0 {
			taxRat = parsedTax
		}
	}

	grossPayoutTotal := new(big.Rat)
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item.OptionKey))
		_, isWin := tags[normalized]

		payout := "0"
		if isWin {
			itemStake, err := parseNumeric(item.Stake)
			if err != nil {
				return nil, "", err
			}
			odds := settlementOddsForItem(item.OptionType, item.OptionKey)
			if odds.Sign() <= 0 {
				odds = big.NewRat(2, 1)
			}
			grossPayout := new(big.Rat).Mul(itemStake, odds)
			payout = grossPayout.FloatString(8)
			grossPayoutTotal.Add(grossPayoutTotal, grossPayout)
		}

		outcomes = append(outcomes, ticketSettleItemOutcome{
			OptionType: item.OptionType,
			OptionKey:  item.OptionKey,
			Stake:      item.Stake,
			IsWin:      isWin,
			Payout:     payout,
		})
	}

	if grossPayoutTotal.Sign() > 0 && taxRat.Sign() > 0 {
		grossPayoutTotal.Sub(grossPayoutTotal, taxRat)
	}
	if grossPayoutTotal.Sign() < 0 {
		grossPayoutTotal.SetInt64(0)
	}
	payoutTotal = grossPayoutTotal.FloatString(8)

	return outcomes, payoutTotal, nil
}

func decodeResultTags(payload []byte) (map[string]struct{}, error) {
	var decoded struct {
		Tags []string `json:"tags"`
	}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		return nil, err
	}

	tags := make(map[string]struct{}, len(decoded.Tags))
	for _, item := range decoded.Tags {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		tags[normalized] = struct{}{}
	}

	return tags, nil
}

func buildPeriodNo(roomCode string, drawAt time.Time) string {
	base := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(roomCode), "-", "_"))
	return fmt.Sprintf("%s_%d", base, drawAt.Unix())
}

func toGameTypeSlug(gameType int) string {
	switch gameType {
	case 1:
		return "wingo"
	case 2:
		return "k3"
	case 3:
		return "lottery"
	default:
		return "unknown"
	}
}
