package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gin/internal/domain/auth"
	"gin/internal/support/clock"
	"gin/internal/support/message"
	"gin/internal/support/phone"
)

var (
	ErrOTPNotFound       = errors.New(message.OTPInvalid)
	ErrOTPExpired        = errors.New(message.OTPExpired)
	ErrOTPLocked         = errors.New(message.OTPLocked)
	ErrResetTokenInvalid = errors.New(message.ResetTokenInvalid)
)

type OTPRequestRecord struct {
	ID           int64
	UserID       *int64
	Channel      int
	Purpose      int
	Target       string
	OTPHash      string
	OTPLast4     string
	RequestToken string
	AttemptCount int
	MaxAttempts  int
	ExpiresAt    time.Time
	VerifiedAt   *time.Time
	UsedAt       *time.Time
	LockedAt     *time.Time
	SentAt       *time.Time
	Status       int
}

func (r *UserRepository) FindUserByChannel(ctx context.Context, channel auth.OTPChannel, account string) (auth.UserProfile, error) {
	normalized := normalizeAccount(channel, account)
	if normalized == "" {
		return auth.UserProfile{}, ErrAccountNotFound
	}

	var query string
	switch channel {
	case auth.OTPChannelEmail:
		query = `
			select id, name, email, phone, password, role, status, email_verified_at, phone_verified_at, last_login_at, created_at, updated_at, deleted_at
			from users
			where lower(email) = $1
			limit 1
		`
	case auth.OTPChannelPhone:
		variants := phone.VNPhoneVariants(normalized)
		if len(variants) == 0 {
			return auth.UserProfile{}, ErrAccountNotFound
		}

		placeholders := make([]string, 0, len(variants))
		args := make([]any, 0, len(variants))
		for idx, variant := range variants {
			placeholders = append(placeholders, fmt.Sprintf("$%d", idx+1))
			args = append(args, variant)
		}

		query = `
			select id, name, email, phone, password, role, status, email_verified_at, phone_verified_at, last_login_at, created_at, updated_at, deleted_at
			from users
			where phone in (` + strings.Join(placeholders, ", ") + `)
			limit 1
		`

		record, err := scanUserRecord(r.db.QueryRowContext(ctx, query, args...))
		if err != nil {
			return auth.UserProfile{}, err
		}

		if record.DeletedAt != nil {
			return auth.UserProfile{}, ErrAccountNotFound
		}

		profile, err := r.findAffiliateProfileByUserID(ctx, record.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return auth.UserProfile{}, err
		}

		return auth.UserProfile{
			User:             record.User,
			AffiliateProfile: profile,
		}, nil
	default:
		return auth.UserProfile{}, ErrAccountNotFound
	}

	record, err := scanUserRecord(r.db.QueryRowContext(ctx, query, normalized))
	if err != nil {
		return auth.UserProfile{}, err
	}

	if record.DeletedAt != nil {
		return auth.UserProfile{}, ErrAccountNotFound
	}

	profile, err := r.findAffiliateProfileByUserID(ctx, record.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return auth.UserProfile{}, err
	}

	return auth.UserProfile{
		User:             record.User,
		AffiliateProfile: profile,
	}, nil
}

func (r *UserRepository) CreateForgotPasswordOTP(
	ctx context.Context,
	userID int64,
	channel int,
	target string,
	otpHash string,
	otpLast4 string,
	requestToken string,
	expiresAt time.Time,
	maxAttempts int,
) (OTPRequestRecord, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return OTPRequestRecord{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(ctx, `
		update auth_otp_requests
		set status = $1, updated_at = now()
		where user_id = $2 and channel = $3 and purpose = $4 and status in ($5, $6)
	`, auth.OTPStatusCancelled, userID, channel, auth.OTPPurposeResetPassword, auth.OTPStatusPending, auth.OTPStatusVerified); err != nil {
		return OTPRequestRecord{}, err
	}

	row := tx.QueryRowContext(ctx, `
		insert into auth_otp_requests (
			user_id, channel, purpose, target, otp_hash, otp_last4, request_token,
			attempt_count, max_attempts, expires_at, status, created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, 0, $8, $9, $10, now(), now())
		returning id, user_id, channel, purpose, target, otp_hash, otp_last4, request_token,
		          attempt_count, max_attempts, expires_at, verified_at, used_at, locked_at, sent_at, status
	`, userID, channel, auth.OTPPurposeResetPassword, target, otpHash, otpLast4, requestToken, maxAttempts, expiresAt, auth.OTPStatusPending)

	record, err := scanOTPRequest(row)
	if err != nil {
		return OTPRequestRecord{}, err
	}

	if err := tx.Commit(); err != nil {
		return OTPRequestRecord{}, err
	}

	return record, nil
}

func (r *UserRepository) MarkOTPSent(ctx context.Context, otpRequestID int64, sentAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		update auth_otp_requests
		set sent_at = $1, updated_at = $1
		where id = $2
	`, sentAt, otpRequestID)
	return err
}

func (r *UserRepository) FindLatestPendingOTP(ctx context.Context, channel int, target string) (OTPRequestRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		select id, user_id, channel, purpose, target, otp_hash, otp_last4, request_token,
		       attempt_count, max_attempts, expires_at, verified_at, used_at, locked_at, sent_at, status
		from auth_otp_requests
		where channel = $1 and target = $2 and purpose = $3 and status = $4
		order by id desc
		limit 1
	`, channel, target, auth.OTPPurposeResetPassword, auth.OTPStatusPending)

	record, err := scanOTPRequest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OTPRequestRecord{}, ErrOTPNotFound
		}

		return OTPRequestRecord{}, err
	}

	return record, nil
}

func (r *UserRepository) MarkOTPAttempt(ctx context.Context, record OTPRequestRecord, locked bool, lockedAt *time.Time) error {
	nextStatus := auth.OTPStatusPending
	if locked {
		nextStatus = auth.OTPStatusLocked
	}

	_, err := r.db.ExecContext(ctx, `
		update auth_otp_requests
		set attempt_count = $1, status = $2, locked_at = $3, updated_at = now()
		where id = $4
	`, record.AttemptCount, nextStatus, lockedAt, record.ID)

	return err
}

func (r *UserRepository) MarkOTPVerified(ctx context.Context, requestID int64, verifiedAt time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		update auth_otp_requests
		set status = $1, verified_at = $2, updated_at = $2
		where id = $3
	`, auth.OTPStatusVerified, verifiedAt, requestID)
	return err
}

func (r *UserRepository) ResetPasswordWithVerifiedOTP(ctx context.Context, requestToken string, passwordHash string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	row := tx.QueryRowContext(ctx, `
		select id, user_id, channel, purpose, target, otp_hash, otp_last4, request_token,
		       attempt_count, max_attempts, expires_at, verified_at, used_at, locked_at, sent_at, status
		from auth_otp_requests
		where request_token = $1 and purpose = $2 and status = $3
		limit 1
	`, requestToken, auth.OTPPurposeResetPassword, auth.OTPStatusVerified)

	record, err := scanOTPRequest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrResetTokenInvalid
		}
		return err
	}

	if clock.Now().After(record.ExpiresAt) {
		return ErrOTPExpired
	}

	if record.UserID == nil || *record.UserID == 0 {
		return ErrResetTokenInvalid
	}

	now := clock.Now()
	if _, err := tx.ExecContext(ctx, `
		update users set password = $1, updated_at = $2 where id = $3
	`, passwordHash, now, *record.UserID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		update auth_otp_requests
		set status = $1, used_at = $2, updated_at = $2
		where id = $3
	`, auth.OTPStatusUsed, now, record.ID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		update auth_otp_requests
		set status = $1, updated_at = $2
		where user_id = $3 and purpose = $4 and status in ($5, $6) and id <> $7
	`, auth.OTPStatusCancelled, now, *record.UserID, auth.OTPPurposeResetPassword, auth.OTPStatusPending, auth.OTPStatusVerified, record.ID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func scanOTPRequest(row *sql.Row) (OTPRequestRecord, error) {
	var record OTPRequestRecord
	err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.Channel,
		&record.Purpose,
		&record.Target,
		&record.OTPHash,
		&record.OTPLast4,
		&record.RequestToken,
		&record.AttemptCount,
		&record.MaxAttempts,
		&record.ExpiresAt,
		&record.VerifiedAt,
		&record.UsedAt,
		&record.LockedAt,
		&record.SentAt,
		&record.Status,
	)
	return record, err
}

func normalizeAccount(channel auth.OTPChannel, account string) string {
	trimmed := strings.TrimSpace(account)
	switch channel {
	case auth.OTPChannelEmail:
		return strings.ToLower(trimmed)
	case auth.OTPChannelPhone:
		return phone.NormalizeVNPhone(trimmed)
	default:
		return ""
	}
}
