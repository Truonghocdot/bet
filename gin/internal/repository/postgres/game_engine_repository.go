package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
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
	log.Printf("[engine][period.ensure.start] room_code=%s game_type=%d duration_seconds=%d now_vn=%s", room.Code, room.GameType, room.DurationSeconds, nowVN.Format(time.RFC3339Nano))

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

	created := make([]GamePeriodRecord, 0, 1)

	var upcomingCount int
	err = tx.QueryRowContext(ctx, `
		select count(*)
		from game_periods
		where room_code = $1
		  and draw_at > $2
	`, room.Code, nowVN).Scan(&upcomingCount)
	if err != nil {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=count_upcoming err=%v", room.Code, err)
		return nil, err
	}
	log.Printf("[engine][period.ensure.state] room_code=%s upcoming_count=%d now_vn=%s", room.Code, upcomingCount, nowVN.Format(time.RFC3339Nano))
	if upcomingCount > 0 {
		if err := tx.Commit(); err != nil {
			log.Printf("[engine][period.ensure.error] room_code=%s stage=commit_skip err=%v", room.Code, err)
			return nil, err
		}
		log.Printf("[engine][period.ensure.skip] room_code=%s reason=upcoming_period_exists", room.Code)
		return nil, nil
	}

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

	nextOpenAt := nowVN
	if latestDrawAt.Valid {
		latestVN := latestDrawAt.Time.In(clock.Location())
		log.Printf("[engine][period.ensure.state] room_code=%s latest_draw_at_vn=%s action=create_next_from_now", room.Code, latestVN.Format(time.RFC3339Nano))
		if latestVN.After(nowVN) {
			nextOpenAt = latestVN
		}
	} else {
		log.Printf("[engine][period.ensure.state] room_code=%s latest_draw_at=nil action=bootstrap_initial_period", room.Code)
	}

	nextDrawAt := nextOpenAt.Add(duration)
	candidateNo := buildPeriodNo(room.Code, nextDrawAt)
	log.Printf(
		"[engine][period.ensure.compare] room_code=%s next_open_at=%s next_draw_at=%s period_no=%s",
		room.Code,
		nextOpenAt.Format(time.RFC3339Nano),
		nextDrawAt.Format(time.RFC3339Nano),
		candidateNo,
	)

	var existingID int64
	err = tx.QueryRowContext(ctx, `
		select id
		from game_periods
		where room_code = $1 and period_no = $2
		limit 1
	`, room.Code, candidateNo).Scan(&existingID)
	if err == nil {
		if err := tx.Commit(); err != nil {
			log.Printf("[engine][period.ensure.error] room_code=%s stage=commit_duplicate_skip err=%v", room.Code, err)
			return nil, err
		}
		log.Printf("[engine][period.ensure.skip] room_code=%s period_no=%s reason=duplicate_exists existing_id=%d", room.Code, candidateNo, existingID)
		return nil, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=check_duplicate period_no=%s err=%v", room.Code, candidateNo, err)
		return nil, err
	}

	status := periodStatusScheduled
	if nextOpenAt.Equal(nowVN) || nextOpenAt.Before(nowVN) {
		status = periodStatusOpen
	}
	record, inserted, err := r.insertPeriodTx(ctx, tx, room, nextOpenAt, nextDrawAt, status)
	if err != nil {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=insert_single open_at=%s draw_at=%s err=%v", room.Code, nextOpenAt.Format(time.RFC3339Nano), nextDrawAt.Format(time.RFC3339Nano), err)
		return nil, err
	}
	if inserted {
		created = append(created, record)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[engine][period.ensure.error] room_code=%s stage=commit err=%v", room.Code, err)
		return nil, err
	}

	log.Printf("[engine][period.ensure.done] room_code=%s created_count=%d", room.Code, len(created))
	return created, nil
}

func (r *GameRepository) MoveScheduledToOpen(ctx context.Context, now time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		update game_periods
		set status = $1,
		    updated_at = now()
		where status = $2 and open_at <= $3
	`, periodStatusOpen, periodStatusScheduled, now)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *GameRepository) MoveOpenToLocked(ctx context.Context, now time.Time) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		update game_periods
		set status = $1,
		    updated_at = now()
		where status = $2 and bet_lock_at <= $3
	`, periodStatusLocked, periodStatusOpen, now)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *GameRepository) ListLockedPeriodsForDraw(ctx context.Context, now time.Time, limit int) ([]GamePeriodRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, room_code, game_type, period_no, open_at, bet_lock_at, draw_at, status
		from game_periods
		where status = $1 and draw_at <= $2
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
			&item.OpenAt,
			&item.BetLockAt,
			&item.DrawAt,
			&item.Status,
		); err != nil {
			return nil, err
		}
		records = append(records, item)
	}

	return records, rows.Err()
}

func (r *GameRepository) MarkPeriodDrawn(ctx context.Context, period GamePeriodRecord, draw DrawResult) error {
	log.Printf("[engine][period.draw.start] period_id=%d room_code=%s period_no=%s draw_at=%s result=%s big_small=%s color=%s", period.ID, period.RoomCode, period.PeriodNo, period.DrawAt.Format(time.RFC3339Nano), draw.Result, draw.BigSmall, draw.Color)

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

	log.Printf("[engine][period.draw.done] period_id=%d room_code=%s period_no=%s", period.ID, period.RoomCode, period.PeriodNo)
	return nil
}

func (r *GameRepository) ListDrawnPeriodsForSettlement(ctx context.Context, limit int) ([]GamePeriodSettlementRecord, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(ctx, `
		select id, room_code, period_no, game_type, result_payload
		from game_periods
		where status = $1 and result_payload is not null
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
		if err := rows.Scan(&item.ID, &item.RoomCode, &item.PeriodNo, &item.GameType, &item.ResultPayload); err != nil {
			return nil, err
		}
		periods = append(periods, item)
	}

	return periods, rows.Err()
}

func (r *GameRepository) SettlePeriod(ctx context.Context, period GamePeriodSettlementRecord) error {
	log.Printf("[engine][period.settle.start] period_id=%d room_code=%s period_no=%s", period.ID, period.RoomCode, period.PeriodNo)

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=begin_tx err=%v", period.ID, err)
		return err
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
		return err
	}
	if currentStatus != periodStatusDrawn {
		log.Printf("[engine][period.settle.skip] period_id=%d current_status=%d", period.ID, currentStatus)
		return nil
	}

	tags, err := decodeResultTags(period.ResultPayload)
	if err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=decode_tags err=%v", period.ID, err)
		return err
	}

	rows, err := tx.QueryContext(ctx, `
		select id, user_id, wallet_id, coalesce(total_stake, stake)::text, items
		from bet_tickets
		where period_id = $1 and status = $2
		order by id asc
		for update
	`, period.ID, betStatusPending)
	if err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=load_tickets err=%v", period.ID, err)
		return err
	}
	defer rows.Close()

	tickets := make([]SettleBetTicketRecord, 0)
	for rows.Next() {
		var ticket SettleBetTicketRecord
		if err := rows.Scan(&ticket.ID, &ticket.UserID, &ticket.WalletID, &ticket.TotalStake, &ticket.ItemsJSON); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d stage=scan_ticket err=%v", period.ID, err)
			return err
		}
		tickets = append(tickets, ticket)
	}
	if err := rows.Err(); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=rows_err err=%v", period.ID, err)
		return err
	}
	log.Printf("[engine][period.settle.info] period_id=%d tickets=%d", period.ID, len(tickets))

	now := clock.Now()
	for _, ticket := range tickets {
		outcomes, payoutTotal, err := settleTicketItems(ticket.ItemsJSON, tags)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=settle_ticket_items err=%v", period.ID, ticket.ID, err)
			return err
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
			return err
		}

		lockedAfter, err := subtractNumeric(lockedBefore, ticket.TotalStake)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=subtract_locked err=%v", period.ID, ticket.ID, err)
			return err
		}
		if compareNumeric(lockedAfter, "0") < 0 {
			lockedAfter = "0"
		}

		balanceAfter, err := addNumeric(balanceBefore, payoutTotal)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=add_balance err=%v", period.ID, ticket.ID, err)
			return err
		}
		balanceAfter, err = subtractNumeric(balanceAfter, ticket.TotalStake)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=subtract_stake err=%v", period.ID, ticket.ID, err)
			return err
		}

		if _, err := tx.ExecContext(ctx, `
			update wallets
			set balance = $1::numeric(20,8),
			    locked_balance = $2::numeric(20,8),
			    updated_at = $3
			where id = $4
		`, balanceAfter, lockedAfter, now, ticket.WalletID); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=update_wallet err=%v", period.ID, ticket.ID, err)
			return err
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
				return err
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
			return err
		}

		profitLoss, err := subtractNumeric(ticket.TotalStake, payoutTotal)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=profit_loss err=%v", period.ID, ticket.ID, err)
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			insert into bet_settlements (
				ticket_id, period_id, settlement_type, before_status, after_status,
				payout_amount, profit_loss, note, created_at
			)
			values ($1, $2, 1, $3, $4, $5::numeric(20,8), $6::numeric(20,8), $7, $8)
		`, ticket.ID, period.ID, betStatusPending, statusAfter, payoutTotal, profitLoss, "Engine settlement tự động", now); err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=insert_settlement err=%v", period.ID, ticket.ID, err)
			return err
		}

		netDelta, err := subtractNumeric(payoutTotal, ticket.TotalStake)
		if err != nil {
			log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=net_delta err=%v", period.ID, ticket.ID, err)
			return err
		}
		if compareNumeric(netDelta, "0") != 0 {
			direction := 1
			ledgerAmount := netDelta
			note := "Cộng tiền thắng cược"
			if compareNumeric(netDelta, "0") < 0 {
				direction = 2
				ledgerAmount, err = subtractNumeric("0", netDelta)
				if err != nil {
					log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=ledger_amount err=%v", period.ID, ticket.ID, err)
					return err
				}
				note = "Trừ tiền thua cược"
			}
			if _, err := tx.ExecContext(ctx, `
				insert into wallet_ledger_entries (
					wallet_id, user_id, direction, amount, balance_before, balance_after,
					reference_type, reference_id, note, created_at
				)
				values ($1, $2, $3, $4::numeric(20,8), $5::numeric(20,8), $6::numeric(20,8),
				        $7, $8, $9, $10)
			`, ticket.WalletID, ticket.UserID, direction, ledgerAmount, balanceBefore, balanceAfter, "App\\Models\\Bet\\BetTicket", ticket.ID, note, now); err != nil {
				log.Printf("[engine][period.settle.error] period_id=%d ticket_id=%d stage=insert_ledger err=%v", period.ID, ticket.ID, err)
				return err
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
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		update game_round_histories
		set status = 'SETTLED',
		    updated_at = $1
		where room_code = $2
		  and period_no = $3
	`, now, period.RoomCode, period.PeriodNo); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=update_round_history err=%v", period.ID, err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[engine][period.settle.error] period_id=%d stage=commit err=%v", period.ID, err)
		return err
	}

	log.Printf("[engine][period.settle.done] period_id=%d room_code=%s period_no=%s tickets=%d", period.ID, period.RoomCode, period.PeriodNo, len(tickets))
	return nil
}

func (r *GameRepository) insertPeriodTx(ctx context.Context, tx *sql.Tx, room GameRoomRecord, openAt, drawAt time.Time, status int) (GamePeriodRecord, bool, error) {
	lockAt := drawAt.Add(-time.Duration(room.BetCutoffSeconds) * time.Second)
	periodNo := buildPeriodNo(room.Code, drawAt)
	var record GamePeriodRecord
	err := tx.QueryRowContext(ctx, `
		insert into game_periods (
			game_type, period_no, room_code, open_at, close_at, bet_lock_at, draw_at, status, created_at, updated_at
		)
		values (
			$1, $2, $3, $4, $5, $5, $6, $7, now(), now()
		)
		on conflict (room_code, period_no) do nothing
		returning id, room_code, game_type, period_no, open_at, bet_lock_at, draw_at, status
	`, room.GameType, periodNo, room.Code, openAt, lockAt, drawAt, status).Scan(
		&record.ID,
		&record.RoomCode,
		&record.GameType,
		&record.PeriodNo,
		&record.OpenAt,
		&record.BetLockAt,
		&record.DrawAt,
		&record.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[engine][period.insert.skip] room_code=%s period_no=%s reason=conflict_or_no_rows", room.Code, periodNo)
			return GamePeriodRecord{}, false, nil
		}
		log.Printf("[engine][period.insert.error] room_code=%s period_no=%s err=%v", room.Code, periodNo, err)
		return GamePeriodRecord{}, false, err
	}

	log.Printf("[engine][period.insert.done] room_code=%s period_id=%d period_no=%s open_at=%s draw_at=%s status=%d", record.RoomCode, record.ID, record.PeriodNo, record.OpenAt.Format(time.RFC3339Nano), record.DrawAt.Format(time.RFC3339Nano), record.Status)
	return record, true, nil
}

func settleTicketItems(rawItems []byte, tags map[string]struct{}) ([]ticketSettleItemOutcome, string, error) {
	var items []BetTicketItemRecord
	if len(rawItems) > 0 {
		if err := json.Unmarshal(rawItems, &items); err != nil {
			return nil, "", err
		}
	}

	outcomes := make([]ticketSettleItemOutcome, 0, len(items))
	payoutTotal := "0"
	for _, item := range items {
		normalized := strings.ToLower(strings.TrimSpace(item.OptionKey))
		_, isWin := tags[normalized]

		payout := "0"
		if isWin {
			// Phase 1 payout rule: cửa thắng trả x2 stake.
			twoTimes, err := addNumeric(item.Stake, item.Stake)
			if err != nil {
				return nil, "", err
			}
			payout = twoTimes
		}

		newTotal, err := addNumeric(payoutTotal, payout)
		if err != nil {
			return nil, "", err
		}
		payoutTotal = newTotal
		outcomes = append(outcomes, ticketSettleItemOutcome{
			OptionType: item.OptionType,
			OptionKey:  item.OptionKey,
			Stake:      item.Stake,
			IsWin:      isWin,
			Payout:     payout,
		})
	}

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
