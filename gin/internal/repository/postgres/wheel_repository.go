package postgres

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gin/internal/domain/wheel"
)

var (
	ErrWheelInvitationNotFound = errors.New("wheel.invitation.not_found")
	ErrWheelInvitationInactive = errors.New("wheel.invitation.inactive")
	ErrWheelSessionNotStarted  = errors.New("wheel.session.not_started")
	ErrWheelSessionExpired     = errors.New("wheel.session.expired")
	ErrWheelSessionCompleted   = errors.New("wheel.session.completed")
	ErrWheelRoundOrder         = errors.New("wheel.round.invalid_order")
	ErrWheelRoundNotReady      = errors.New("wheel.round.not_ready")
	ErrWheelChatBanned         = errors.New("wheel.chat.banned")
)

const wheelTotalRounds = 3

type WheelRepository struct {
	db *sql.DB
}

type WheelInvitationAccess struct {
	ID              int64
	PublicID        string
	UserID          int64
	CampaignName    string
	Status          string
	ExpiresAt       *time.Time
	PopupSeenAt     *time.Time
	SessionID       *int64
	SessionPublicID *string
	SessionStatus   *string
}

type WheelMutationResult struct {
	SessionID int64
	State     wheel.State
	Round     wheel.Round
	OutboxIDs []int64
}

type WheelNotReadyError struct {
	AvailableAt time.Time
}

func (e *WheelNotReadyError) Error() string { return ErrWheelRoundNotReady.Error() }
func (e *WheelNotReadyError) Unwrap() error { return ErrWheelRoundNotReady }

func NewWheelRepository(db *sql.DB) *WheelRepository { return &WheelRepository{db: db} }

func (r *WheelRepository) ListInvitations(ctx context.Context, userID int64) ([]wheel.Invitation, error) {
	rows, err := r.db.QueryContext(ctx, `
		select wi.public_id::text, wc.name, wi.status, wi.expires_at, wi.popup_seen_at,
		       ws.public_id::text, ws.status
		from wheel_invitations wi
		join wheel_campaigns wc on wc.id = wi.campaign_id
		left join wheel_sessions ws on ws.invitation_id = wi.id
		where wi.user_id = $1
		  and wi.status <> 'draft'
		order by wi.activated_at desc nulls last, wi.id desc
		limit 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]wheel.Invitation, 0)
	for rows.Next() {
		var publicID, campaignName, status string
		var expiresAt, seenAt sql.NullTime
		var sessionID, sessionStatus sql.NullString
		if err := rows.Scan(&publicID, &campaignName, &status, &expiresAt, &seenAt, &sessionID, &sessionStatus); err != nil {
			return nil, err
		}
		item := wheel.Invitation{ID: publicID, CampaignName: campaignName, Status: status}
		if expiresAt.Valid {
			value := formatWheelTime(expiresAt.Time)
			item.ExpiresAt = &value
		}
		if seenAt.Valid {
			value := formatWheelTime(seenAt.Time)
			item.SeenAt = &value
		}
		if sessionID.Valid {
			value := sessionID.String
			item.SessionID = &value
		}
		if sessionStatus.Valid {
			value := sessionStatus.String
			item.SessionStatus = &value
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *WheelRepository) FindInvitationForUser(ctx context.Context, publicID string, userID int64) (WheelInvitationAccess, error) {
	return scanWheelInvitation(r.db.QueryRowContext(ctx, `
		select wi.id, wi.public_id::text, wi.user_id, wc.name, wi.status, wi.expires_at, wi.popup_seen_at,
		       ws.id, ws.public_id::text, ws.status
		from wheel_invitations wi
		join wheel_campaigns wc on wc.id = wi.campaign_id
		left join wheel_sessions ws on ws.invitation_id = wi.id
		where wi.public_id = $1::uuid and wi.user_id = $2`, publicID, userID))
}

func (r *WheelRepository) FindInvitationByID(ctx context.Context, invitationID, userID int64) (WheelInvitationAccess, error) {
	return scanWheelInvitation(r.db.QueryRowContext(ctx, `
		select wi.id, wi.public_id::text, wi.user_id, wc.name, wi.status, wi.expires_at, wi.popup_seen_at,
		       ws.id, ws.public_id::text, ws.status
		from wheel_invitations wi
		join wheel_campaigns wc on wc.id = wi.campaign_id
		left join wheel_sessions ws on ws.invitation_id = wi.id
		where wi.id = $1 and wi.user_id = $2`, invitationID, userID))
}

func scanWheelInvitation(row *sql.Row) (WheelInvitationAccess, error) {
	var record WheelInvitationAccess
	var expiresAt, seenAt sql.NullTime
	var sessionID sql.NullInt64
	var sessionPublicID, sessionStatus sql.NullString
	err := row.Scan(&record.ID, &record.PublicID, &record.UserID, &record.CampaignName, &record.Status, &expiresAt, &seenAt, &sessionID, &sessionPublicID, &sessionStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return WheelInvitationAccess{}, ErrWheelInvitationNotFound
	}
	if err != nil {
		return WheelInvitationAccess{}, err
	}
	if expiresAt.Valid {
		record.ExpiresAt = &expiresAt.Time
	}
	if seenAt.Valid {
		record.PopupSeenAt = &seenAt.Time
	}
	if sessionID.Valid {
		value := sessionID.Int64
		record.SessionID = &value
	}
	if sessionPublicID.Valid {
		value := sessionPublicID.String
		record.SessionPublicID = &value
	}
	if sessionStatus.Valid {
		value := sessionStatus.String
		record.SessionStatus = &value
	}
	return record, nil
}

func (r *WheelRepository) MarkInvitationSeen(ctx context.Context, publicID string, userID int64) error {
	result, err := r.db.ExecContext(ctx, `update wheel_invitations set popup_seen_at = coalesce(popup_seen_at, now()), updated_at = now() where public_id = $1::uuid and user_id = $2 and status <> 'draft'`, publicID, userID)
	if err != nil {
		return err
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return ErrWheelInvitationNotFound
	}
	return nil
}

func (r *WheelRepository) ActivateInvitationChat(ctx context.Context, invitationID, userID int64, durationSeconds int) error {
	if durationSeconds < 1 {
		durationSeconds = 300
	}
	_, err := r.db.ExecContext(ctx, `
		update chat_rooms as cr
		set enabled = true,
			next_bot_at = timezone('UTC', now()),
			bot_active_until = timezone('UTC', now()) + ($3 * interval '1 second'),
			updated_at = timezone('UTC', now())
		from wheel_invitations as wi
		where cr.wheel_invitation_id = wi.id
		  and wi.id = $1
		  and wi.user_id = $2
		  and wi.status in ('pending', 'started')
		  and wi.bot_chat_enabled = true`, invitationID, userID, durationSeconds)

	return err
}

func (r *WheelRepository) StartSession(ctx context.Context, invitationID, userID int64, durationSeconds int) (WheelMutationResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return WheelMutationResult{}, err
	}
	defer tx.Rollback()

	var invitationStatus, invitationPublicID, campaignName string
	var botChatEnabled bool
	var campaignID int64
	var expiresAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		select wi.status, wi.public_id::text, wi.campaign_id, wc.name, wi.expires_at, wi.bot_chat_enabled
		from wheel_invitations wi join wheel_campaigns wc on wc.id = wi.campaign_id
		where wi.id = $1 and wi.user_id = $2 for update`, invitationID, userID).Scan(&invitationStatus, &invitationPublicID, &campaignID, &campaignName, &expiresAt, &botChatEnabled)
	if errors.Is(err, sql.ErrNoRows) {
		return WheelMutationResult{}, ErrWheelInvitationNotFound
	}
	if err != nil {
		return WheelMutationResult{}, err
	}

	if invitationStatus == "started" || invitationStatus == "completed" || invitationStatus == "expired" {
		state, stateErr := r.loadStateTx(ctx, tx, invitationID, userID, durationSeconds)
		if stateErr != nil {
			return WheelMutationResult{}, stateErr
		}
		if err := tx.Commit(); err != nil {
			return WheelMutationResult{}, err
		}
		var internalID int64
		if state.SessionID != nil {
			_ = r.db.QueryRowContext(ctx, `select id from wheel_sessions where public_id = $1::uuid`, *state.SessionID).Scan(&internalID)
		}
		return WheelMutationResult{SessionID: internalID, State: state}, nil
	}
	if invitationStatus != "pending" || (expiresAt.Valid && !time.Now().Before(expiresAt.Time)) {
		return WheelMutationResult{}, ErrWheelInvitationInactive
	}

	var roundCount, requiredSecondRoundCount int
	if err := tx.QueryRowContext(ctx, `
		select count(*), count(*) filter (
			where round_no = 2 and segment_key = 'reward_39m' and prize_amount = 39000000
		)
		from wheel_invitation_rounds where invitation_id = $1`, invitationID).Scan(&roundCount, &requiredSecondRoundCount); err != nil {
		return WheelMutationResult{}, err
	}
	if roundCount != wheelTotalRounds || requiredSecondRoundCount != 1 {
		return WheelMutationResult{}, fmt.Errorf("wheel invitation requires three rounds with a 39m second round")
	}

	now := time.Now().UTC()
	endsAt := now.Add(time.Duration(durationSeconds) * time.Second)
	publicID, err := randomUUID()
	if err != nil {
		return WheelMutationResult{}, err
	}
	var sessionID int64
	err = tx.QueryRowContext(ctx, `
		insert into wheel_sessions (public_id, invitation_id, user_id, status, current_round, started_at, ends_at, version, created_at, updated_at)
		values ($1::uuid, $2, $3, 'active', 1, $4, $5, 1, now(), now()) returning id`, publicID, invitationID, userID, now, endsAt).Scan(&sessionID)
	if err != nil {
		return WheelMutationResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `update wheel_invitations set status = 'started', updated_at = now() where id = $1`, invitationID); err != nil {
		return WheelMutationResult{}, err
	}
	var roomID int64
	err = tx.QueryRowContext(ctx, `select id from chat_rooms where wheel_invitation_id = $1 for update`, invitationID).Scan(&roomID)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `insert into chat_rooms (wheel_session_id, wheel_invitation_id, code, name, enabled, next_bot_at, bot_active_until, bot_message_count, created_at, updated_at) values ($1, $2, $3, $4, true, case when $5 then timezone('UTC', now()) else null end, case when $5 then $6 else null end, 0, timezone('UTC', now()), timezone('UTC', now())) returning id`, sessionID, invitationID, fmt.Sprintf("wheel-session-%d", sessionID), "Phòng sự kiện "+campaignName, botChatEnabled, endsAt).Scan(&roomID)
	} else if err == nil {
		_, err = tx.ExecContext(ctx, `update chat_rooms set wheel_session_id = $1, enabled = true, next_bot_at = case when $2 then timezone('UTC', now()) else null end, bot_active_until = case when $2 then $3 else null end, updated_at = timezone('UTC', now()) where id = $4`, sessionID, botChatEnabled, endsAt, roomID)
	}
	if err != nil {
		return WheelMutationResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `insert into wheel_audit_logs (campaign_id,invitation_id,session_id,actor_user_id,action,new_values,created_at) values ($1,$2,$3,$4,'session.started',$5::jsonb,now())`, campaignID, invitationID, sessionID, userID, fmt.Sprintf(`{"ends_at":%q}`, formatWheelTime(endsAt))); err != nil {
		return WheelMutationResult{}, err
	}
	payload := map[string]any{"session_id": publicID, "invitation_id": invitationPublicID, "started_at": formatWheelTime(now), "ends_at": formatWheelTime(endsAt)}
	outboxID, err := insertWheelOutboxTx(ctx, tx, fmt.Sprintf("stream:wheel:session:%d", sessionID), "wheel.session.started", payload)
	if err != nil {
		return WheelMutationResult{}, err
	}
	state, err := r.loadStateTx(ctx, tx, invitationID, userID, durationSeconds)
	if err != nil {
		return WheelMutationResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return WheelMutationResult{}, err
	}
	return WheelMutationResult{SessionID: sessionID, State: state, OutboxIDs: []int64{outboxID}}, nil
}

func (r *WheelRepository) State(ctx context.Context, invitationID, userID int64, spinDuration int) (wheel.State, error) {
	return r.loadStateQuery(ctx, r.db, invitationID, userID, spinDuration)
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func (r *WheelRepository) loadStateTx(ctx context.Context, tx *sql.Tx, invitationID, userID int64, spinDuration int) (wheel.State, error) {
	return r.loadStateQuery(ctx, tx, invitationID, userID, spinDuration)
}

func (r *WheelRepository) loadStateQuery(ctx context.Context, q queryer, invitationID, userID int64, spinDuration int) (wheel.State, error) {
	var state wheel.State
	var sessionPublicID, sessionStatus sql.NullString
	var startedAt, endsAt sql.NullTime
	var currentRound sql.NullInt64
	err := q.QueryRowContext(ctx, `
		select wi.public_id::text, wc.name, wi.status, ws.public_id::text, ws.status, ws.started_at, ws.ends_at, ws.current_round
		from wheel_invitations wi join wheel_campaigns wc on wc.id = wi.campaign_id
		left join wheel_sessions ws on ws.invitation_id = wi.id
		where wi.id = $1 and wi.user_id = $2`, invitationID, userID).Scan(&state.InvitationID, &state.CampaignName, &state.SessionStatus, &sessionPublicID, &sessionStatus, &startedAt, &endsAt, &currentRound)
	if errors.Is(err, sql.ErrNoRows) {
		return wheel.State{}, ErrWheelInvitationNotFound
	}
	if err != nil {
		return wheel.State{}, err
	}

	now := time.Now().UTC()
	state.ServerNow = formatWheelTime(now)
	state.SpinDurationSeconds = spinDuration
	state.TotalReward = "0"
	state.Rounds = make([]wheel.Round, 0, wheelTotalRounds)
	state.PaidRewards = make([]wheel.Reward, 0)
	if sessionPublicID.Valid {
		value := sessionPublicID.String
		state.SessionID = &value
	}
	if sessionStatus.Valid {
		state.SessionStatus = sessionStatus.String
	}
	if startedAt.Valid {
		value := formatWheelTime(startedAt.Time)
		state.StartedAt = &value
	}
	if endsAt.Valid {
		value := formatWheelTime(endsAt.Time)
		state.EndsAt = &value
	}
	if currentRound.Valid {
		state.CurrentRound = int(currentRound.Int64)
	} else {
		state.CurrentRound = 1
	}
	if state.SessionStatus == "active" && endsAt.Valid && !now.Before(endsAt.Time) {
		state.SessionStatus = "expired"
	}

	rows, err := q.QueryContext(ctx, `select round_no, status, segment_key, result_label, prize_amount::text, spun_at from wheel_invitation_rounds where invitation_id = $1 order by round_no`, invitationID)
	if err != nil {
		return wheel.State{}, err
	}
	defer rows.Close()
	var previousSpunAt *time.Time
	for rows.Next() {
		var roundNo int
		var status, segmentKey, resultLabel, prize string
		var spunAt sql.NullTime
		if err := rows.Scan(&roundNo, &status, &segmentKey, &resultLabel, &prize, &spunAt); err != nil {
			return wheel.State{}, err
		}
		publicStatus := status
		if status == "pending" && state.SessionStatus == "active" && roundNo == state.CurrentRound {
			available := roundNo == 1 || (previousSpunAt != nil && !now.Before(previousSpunAt.Add(time.Duration(spinDuration)*time.Second)))
			if available {
				publicStatus = "ready"
			}
		}
		item := wheel.Round{RoundNo: roundNo, Status: publicStatus}
		if status == "spun" {
			item.SegmentKey, item.ResultLabel, item.PrizeAmount = &segmentKey, &resultLabel, &prize
			if spunAt.Valid {
				value := formatWheelTime(spunAt.Time)
				item.SpunAt = &value
				t := spunAt.Time
				previousSpunAt = &t
			}
		}
		state.Rounds = append(state.Rounds, item)
		if status == "spun" && roundNo == state.CurrentRound-1 && state.SessionStatus == "active" && spunAt.Valid {
			value := formatWheelTime(spunAt.Time.Add(time.Duration(spinDuration) * time.Second))
			state.NextRoundAvailableAt = &value
		}
	}
	if err := rows.Err(); err != nil {
		return wheel.State{}, err
	}

	rewardRows, err := q.QueryContext(ctx, `select round_no, amount::text, status, paid_at from wheel_rewards where session_id = (select id from wheel_sessions where invitation_id = $1) order by round_no`, invitationID)
	if err != nil {
		return wheel.State{}, err
	}
	defer rewardRows.Close()
	for rewardRows.Next() {
		var item wheel.Reward
		var paidAt sql.NullTime
		if err := rewardRows.Scan(&item.RoundNo, &item.Amount, &item.Status, &paidAt); err != nil {
			return wheel.State{}, err
		}
		if paidAt.Valid {
			value := formatWheelTime(paidAt.Time)
			item.PaidAt = &value
		}
		state.PaidRewards = append(state.PaidRewards, item)
	}
	if err := rewardRows.Err(); err != nil {
		return wheel.State{}, err
	}
	if err := q.QueryRowContext(ctx, `select coalesce(sum(amount) filter (where status = 'paid'), 0)::text from wheel_rewards where session_id = (select id from wheel_sessions where invitation_id = $1)`, invitationID).Scan(&state.TotalReward); err != nil {
		return wheel.State{}, err
	}
	return state, nil
}

func (r *WheelRepository) Spin(ctx context.Context, invitationID, userID int64, roundNo, spinDuration int, siteCode string) (WheelMutationResult, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return WheelMutationResult{}, err
	}
	defer tx.Rollback()

	var sessionID int64
	var sessionStatus string
	var currentRound int
	var endsAt time.Time
	err = tx.QueryRowContext(ctx, `select id, status, current_round, ends_at from wheel_sessions where invitation_id = $1 and user_id = $2 for update`, invitationID, userID).Scan(&sessionID, &sessionStatus, &currentRound, &endsAt)
	if errors.Is(err, sql.ErrNoRows) {
		return WheelMutationResult{}, ErrWheelSessionNotStarted
	}
	if err != nil {
		return WheelMutationResult{}, err
	}
	if sessionStatus != "active" && sessionStatus != "completed" {
		return WheelMutationResult{}, ErrWheelSessionExpired
	}
	if sessionStatus == "active" && !time.Now().Before(endsAt) {
		return WheelMutationResult{}, ErrWheelSessionExpired
	}

	var invitationRoundID int64
	var roundStatus, segmentKey, resultLabel, prizeAmount string
	var existingSpunAt sql.NullTime
	err = tx.QueryRowContext(ctx, `select id, status, segment_key, result_label, prize_amount::text, spun_at from wheel_invitation_rounds where invitation_id = $1 and round_no = $2 for update`, invitationID, roundNo).Scan(&invitationRoundID, &roundStatus, &segmentKey, &resultLabel, &prizeAmount, &existingSpunAt)
	if errors.Is(err, sql.ErrNoRows) {
		return WheelMutationResult{}, ErrWheelRoundOrder
	}
	if err != nil {
		return WheelMutationResult{}, err
	}
	if roundStatus == "spun" {
		state, err := r.loadStateTx(ctx, tx, invitationID, userID, spinDuration)
		if err != nil {
			return WheelMutationResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return WheelMutationResult{}, err
		}
		return WheelMutationResult{SessionID: sessionID, State: state, Round: findWheelRound(state.Rounds, roundNo)}, nil
	}
	if sessionStatus == "completed" {
		return WheelMutationResult{}, ErrWheelSessionCompleted
	}
	if roundNo != currentRound || roundNo < 1 || roundNo > wheelTotalRounds {
		return WheelMutationResult{}, ErrWheelRoundOrder
	}

	now := time.Now().UTC()
	if roundNo > 1 {
		var previousSpunAt time.Time
		if err := tx.QueryRowContext(ctx, `select spun_at from wheel_invitation_rounds where invitation_id = $1 and round_no = $2 and status = 'spun'`, invitationID, roundNo-1).Scan(&previousSpunAt); err != nil {
			return WheelMutationResult{}, ErrWheelRoundOrder
		}
		availableAt := previousSpunAt.Add(time.Duration(spinDuration) * time.Second)
		if now.Before(availableAt) {
			return WheelMutationResult{}, &WheelNotReadyError{AvailableAt: availableAt}
		}
	}

	if _, err := tx.ExecContext(ctx, `update wheel_invitation_rounds set status = 'spun', spun_at = $2, updated_at = now() where id = $1`, invitationRoundID, now); err != nil {
		return WheelMutationResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `insert into wheel_audit_logs (invitation_id,session_id,actor_user_id,action,new_values,created_at) values ($1,$2,$3,'round.spun',$4::jsonb,now())`, invitationID, sessionID, userID, fmt.Sprintf(`{"round_no":%d,"result_label":%q,"prize_amount":%q}`, roundNo, resultLabel, prizeAmount)); err != nil {
		return WheelMutationResult{}, err
	}
	outboxIDs := make([]int64, 0, 3)
	roundPayload := map[string]any{"round_no": roundNo, "segment_key": segmentKey, "result_label": resultLabel, "prize_amount": prizeAmount, "spun_at": formatWheelTime(now)}
	if id, err := insertWheelOutboxTx(ctx, tx, fmt.Sprintf("stream:wheel:session:%d", sessionID), "wheel.round.revealed", roundPayload); err != nil {
		return WheelMutationResult{}, err
	} else {
		outboxIDs = append(outboxIDs, id)
	}

	if positiveNumeric(prizeAmount) {
		var walletID int64
		var balanceBefore string
		err := tx.QueryRowContext(ctx, `select id, balance::text from wallets where user_id = $1 and unit = 1 and status = 1 for update`, userID).Scan(&walletID, &balanceBefore)
		if errors.Is(err, sql.ErrNoRows) {
			if _, err := tx.ExecContext(ctx, `insert into wallets (user_id, unit, balance, locked_balance, status, created_at, updated_at) values ($1, 1, 0, 0, 1, now(), now()) on conflict (user_id, unit) do nothing`, userID); err != nil {
				return WheelMutationResult{}, err
			}
			err = tx.QueryRowContext(ctx, `select id, balance::text from wallets where user_id = $1 and unit = 1 and status = 1 for update`, userID).Scan(&walletID, &balanceBefore)
		}
		if err != nil {
			return WheelMutationResult{}, err
		}
		idempotencyKey := fmt.Sprintf("wheel:%s:%d:%d", siteCode, sessionID, roundNo)
		var rewardID int64
		if err := tx.QueryRowContext(ctx, `insert into wheel_rewards (session_id, invitation_round_id, user_id, round_no, unit, amount, status, idempotency_key, attempts, created_at, updated_at) values ($1,$2,$3,$4,1,$5::numeric,'pending',$6,1,now(),now()) returning id`, sessionID, invitationRoundID, userID, roundNo, prizeAmount, idempotencyKey).Scan(&rewardID); err != nil {
			return WheelMutationResult{}, err
		}
		var balanceAfter string
		if err := tx.QueryRowContext(ctx, `update wallets set balance = balance + $1::numeric, updated_at = now() where id = $2 returning balance::text`, prizeAmount, walletID).Scan(&balanceAfter); err != nil {
			return WheelMutationResult{}, err
		}
		var ledgerID int64
		if err := tx.QueryRowContext(ctx, `insert into wallet_ledger_entries (wallet_id,user_id,direction,amount,balance_before,balance_after,reference_type,reference_id,note,created_at) values ($1,$2,1,$3::numeric,$4::numeric,$5::numeric,'wheel_reward',$6,$7,now()) returning id`, walletID, userID, prizeAmount, balanceBefore, balanceAfter, rewardID, fmt.Sprintf("Thưởng vòng quay sự kiện - lượt %d", roundNo)).Scan(&ledgerID); err != nil {
			return WheelMutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `update wheel_rewards set status = 'paid', wallet_ledger_entry_id = $2, paid_at = now(), updated_at = now() where id = $1`, rewardID, ledgerID); err != nil {
			return WheelMutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `insert into wheel_audit_logs (invitation_id,session_id,actor_user_id,action,new_values,created_at) values ($1,$2,$3,'reward.paid',$4::jsonb,now())`, invitationID, sessionID, userID, fmt.Sprintf(`{"reward_id":%d,"round_no":%d,"amount":%q,"ledger_id":%d}`, rewardID, roundNo, prizeAmount, ledgerID)); err != nil {
			return WheelMutationResult{}, err
		}
		rewardPayload := map[string]any{"round_no": roundNo, "amount": prizeAmount, "status": "paid", "balance": balanceAfter}
		if id, err := insertWheelOutboxTx(ctx, tx, fmt.Sprintf("stream:wheel:session:%d", sessionID), "wheel.reward.paid", rewardPayload); err != nil {
			return WheelMutationResult{}, err
		} else {
			outboxIDs = append(outboxIDs, id)
		}
	}

	if roundNo == wheelTotalRounds {
		if _, err := tx.ExecContext(ctx, `update wheel_sessions set status = 'completed', current_round = $2, completed_at = now(), version = version + 1, updated_at = now() where id = $1`, sessionID, wheelTotalRounds); err != nil {
			return WheelMutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `update wheel_invitations set status = 'completed', updated_at = now() where id = $1`, invitationID); err != nil {
			return WheelMutationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `update chat_rooms set enabled = false, updated_at = now() where wheel_session_id = $1`, sessionID); err != nil {
			return WheelMutationResult{}, err
		}
		if id, err := insertWheelOutboxTx(ctx, tx, fmt.Sprintf("stream:wheel:session:%d", sessionID), "wheel.session.completed", map[string]any{"completed_at": formatWheelTime(now)}); err != nil {
			return WheelMutationResult{}, err
		} else {
			outboxIDs = append(outboxIDs, id)
		}
	} else {
		if _, err := tx.ExecContext(ctx, `update wheel_sessions set current_round = $2, version = version + 1, updated_at = now() where id = $1`, sessionID, roundNo+1); err != nil {
			return WheelMutationResult{}, err
		}
	}
	state, err := r.loadStateTx(ctx, tx, invitationID, userID, spinDuration)
	if err != nil {
		return WheelMutationResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return WheelMutationResult{}, err
	}
	return WheelMutationResult{SessionID: sessionID, State: state, Round: findWheelRound(state.Rounds, roundNo), OutboxIDs: outboxIDs}, nil
}

func (r *WheelRepository) MarkOutboxPublished(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	args := make([]any, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}
	_, err := r.db.ExecContext(ctx, `update wheel_outbox_events set published_at = now(), attempts = attempts + 1, last_error = null, updated_at = now() where id in (`+strings.Join(placeholders, ",")+`)`, args...)
	return err
}

func (r *WheelRepository) ListChat(ctx context.Context, invitationID, userID, before, limit int64) ([]wheel.ChatMessage, *int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		select cm.id, cm.display_name, cm.body, cm.actor_type, cm.created_at
		from chat_messages cm join chat_rooms cr on cr.id = cm.room_id
		left join wheel_sessions ws on ws.id = cr.wheel_session_id
		join wheel_invitations wi on wi.id = coalesce(cr.wheel_invitation_id, ws.invitation_id)
		where wi.id = $1 and wi.user_id = $2 and cm.status = 1 and cm.created_at >= timezone('UTC', now()) - interval '6 hours'
		  and ($3::bigint = 0 or cm.id < $3)
		order by cm.id desc limit $4`, invitationID, userID, before, limit+1)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	items := make([]wheel.ChatMessage, 0, limit)
	for rows.Next() {
		var item wheel.ChatMessage
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.DisplayName, &item.Body, &item.ActorType, &createdAt); err != nil {
			return nil, nil, err
		}
		item.CreatedAt = formatWheelTime(createdAt)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	var next *int64
	if int64(len(items)) > limit {
		cursor := items[limit-1].ID
		next = &cursor
		items = items[:limit]
	}
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return items, next, nil
}

func (r *WheelRepository) CreateChat(ctx context.Context, invitationID, userID int64, body string) (wheel.ChatMessage, int64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	defer tx.Rollback()
	var roomID int64
	var invitationStatus string
	var sessionID sql.NullInt64
	var sessionStatus sql.NullString
	var endsAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		select cr.id, wi.status, ws.id, ws.status, ws.ends_at
		from wheel_invitations wi
		join chat_rooms cr on cr.wheel_invitation_id = wi.id and cr.enabled = true
		left join wheel_sessions ws on ws.invitation_id = wi.id
		where wi.id = $1 and wi.user_id = $2
		for update of wi, cr`, invitationID, userID).Scan(&roomID, &invitationStatus, &sessionID, &sessionStatus, &endsAt)
	if errors.Is(err, sql.ErrNoRows) {
		return wheel.ChatMessage{}, 0, ErrWheelSessionNotStarted
	}
	if err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	pending := invitationStatus == "pending" && !sessionID.Valid
	active := invitationStatus == "started" && sessionID.Valid && sessionStatus.String == "active" && endsAt.Valid && time.Now().Before(endsAt.Time)
	if !pending && !active {
		return wheel.ChatMessage{}, 0, ErrWheelSessionExpired
	}
	var banned bool
	if err := tx.QueryRowContext(ctx, `select exists(select 1 from chat_bans where user_id = $1 and (room_id is null or room_id = $2) and revoked_at is null and (expires_at is null or expires_at > now()))`, userID, roomID).Scan(&banned); err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	if banned {
		return wheel.ChatMessage{}, 0, ErrWheelChatBanned
	}
	var lastBody sql.NullString
	_ = tx.QueryRowContext(ctx, `select body from chat_messages where room_id = $1 and user_id = $2 order by id desc limit 1`, roomID, userID).Scan(&lastBody)
	if lastBody.Valid && strings.EqualFold(strings.TrimSpace(lastBody.String), strings.TrimSpace(body)) {
		return wheel.ChatMessage{}, 0, errors.New("wheel.chat.duplicate")
	}
	displayName := fmt.Sprintf("ID game #%d", userID)
	if _, err := tx.ExecContext(ctx, `insert into chat_user_profiles (user_id, display_name, created_at, updated_at) values ($1,$2,now(),now()) on conflict (user_id) do update set display_name = excluded.display_name, updated_at = excluded.updated_at`, userID, displayName); err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	var item wheel.ChatMessage
	var createdAt time.Time
	createdAtUTC := time.Now().UTC()
	if err := tx.QueryRowContext(ctx, `insert into chat_messages (room_id,actor_type,user_id,display_name,body,status,created_at,updated_at) values ($1,'user',$2,$3,$4,1,$5,$5) returning id,created_at`, roomID, userID, displayName, body, createdAtUTC).Scan(&item.ID, &createdAt); err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	item.DisplayName, item.Body, item.ActorType, item.CreatedAt = displayName, body, "user", formatWheelTime(createdAt)
	payload, _ := json.Marshal(item)
	topic := fmt.Sprintf("stream:wheel:invitation:%d", invitationID)
	if sessionID.Valid {
		topic = fmt.Sprintf("stream:wheel:session:%d", sessionID.Int64)
	}
	if _, err := insertWheelOutboxTx(ctx, tx, topic, "chat.message.created", json.RawMessage(payload)); err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	if err := tx.Commit(); err != nil {
		return wheel.ChatMessage{}, 0, err
	}
	return item, sessionID.Int64, nil
}

func insertWheelOutboxTx(ctx context.Context, tx *sql.Tx, topic, event string, payload any) (int64, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRowContext(ctx, `insert into wheel_outbox_events (topic,event,payload,attempts,available_at,created_at,updated_at) values ($1,$2,$3::jsonb,0,now(),now(),now()) returning id`, topic, event, string(raw)).Scan(&id)
	return id, err
}

func findWheelRound(rounds []wheel.Round, roundNo int) wheel.Round {
	for _, item := range rounds {
		if item.RoundNo == roundNo {
			return item
		}
	}
	return wheel.Round{RoundNo: roundNo}
}
func formatWheelTime(value time.Time) string { return value.UTC().Format(time.RFC3339Nano) }
func positiveNumeric(value string) bool {
	rat := new(big.Rat)
	if _, ok := rat.SetString(value); !ok {
		return false
	}
	return rat.Sign() > 0
}

func randomUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], bytes[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], bytes[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], bytes[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], bytes[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], bytes[10:16])
	return string(buf), nil
}
