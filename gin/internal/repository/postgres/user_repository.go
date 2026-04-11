package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gin/internal/domain/auth"
	"gin/internal/domain/user"
	"gin/internal/support/clock"
	"gin/internal/support/id"
	"gin/internal/support/message"
)

var (
	ErrEmailExists         = errors.New(message.EmailExists)
	ErrPhoneExists         = errors.New(message.PhoneExists)
	ErrAccountNotFound     = errors.New(message.AccountNotFound)
	ErrRefCodeNotFound     = errors.New(message.ReferralCodeNotFound)
	ErrUserDisabled        = errors.New(message.UserNotActive)
	ErrInvalidSelfReferral = errors.New(message.InvalidSelfReferral)
)

type UserRepository struct {
	db *sql.DB
}

type RegisterUserParams struct {
	Name         string
	Email        string
	Phone        *string
	PasswordHash string
	RefCode      string
	RegisterURL  string
}

type userRecord struct {
	auth.User
	PasswordHash string
	DeletedAt    *time.Time
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateRegisteredUser(ctx context.Context, params RegisterUserParams) (auth.UserProfile, error) {
	if exists, err := r.emailExists(ctx, params.Email); err != nil {
		return auth.UserProfile{}, err
	} else if exists {
		return auth.UserProfile{}, ErrEmailExists
	}

	if params.Phone != nil {
		if exists, err := r.phoneExists(ctx, *params.Phone); err != nil {
			return auth.UserProfile{}, err
		} else if exists {
			return auth.UserProfile{}, ErrPhoneExists
		}
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return auth.UserProfile{}, err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	createdUser, err := r.insertUser(ctx, tx, params)
	if err != nil {
		return auth.UserProfile{}, err
	}

	if err := r.insertDefaultWallets(ctx, tx, createdUser.ID); err != nil {
		return auth.UserProfile{}, err
	}

	affiliateProfile, err := r.insertAffiliateProfile(ctx, tx, createdUser.ID, params.RegisterURL)
	if err != nil {
		return auth.UserProfile{}, err
	}

	if strings.TrimSpace(params.RefCode) != "" {
		if err := r.insertAffiliateReferral(ctx, tx, strings.TrimSpace(params.RefCode), createdUser.ID); err != nil {
			return auth.UserProfile{}, err
		}
	}

	now := clock.Now()
	if err := r.updateLastLoginAt(ctx, tx, createdUser.ID, now); err != nil {
		return auth.UserProfile{}, err
	}
	createdUser.LastLoginAt = &now

	if err := tx.Commit(); err != nil {
		return auth.UserProfile{}, err
	}

	return auth.UserProfile{
		User:             createdUser,
		AffiliateProfile: affiliateProfile,
	}, nil
}

func (r *UserRepository) FindByAccount(ctx context.Context, account string) (auth.UserProfile, string, error) {
	record, err := r.findUserRecordByAccount(ctx, account)
	if err != nil {
		return auth.UserProfile{}, "", err
	}

	if record.DeletedAt != nil {
		return auth.UserProfile{}, "", ErrAccountNotFound
	}

	if record.Status != user.StatusActive {
		return auth.UserProfile{}, "", ErrUserDisabled
	}

	profile, err := r.findAffiliateProfileByUserID(ctx, record.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return auth.UserProfile{}, "", err
	}

	return auth.UserProfile{
		User:             record.User,
		AffiliateProfile: profile,
	}, record.PasswordHash, nil
}

func (r *UserRepository) MarkLoggedIn(ctx context.Context, userID int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `update users set last_login_at = $1, updated_at = $1 where id = $2`, at, userID)
	return err
}

func (r *UserRepository) FindProfileByUserID(ctx context.Context, userID int64) (auth.UserProfile, error) {
	record, err := r.findUserRecordByID(ctx, userID)
	if err != nil {
		return auth.UserProfile{}, err
	}

	if record.DeletedAt != nil {
		return auth.UserProfile{}, ErrAccountNotFound
	}

	if record.Status != user.StatusActive {
		return auth.UserProfile{}, ErrUserDisabled
	}

	profile, err := r.findAffiliateProfileByUserID(ctx, userID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return auth.UserProfile{}, err
	}

	return auth.UserProfile{
		User:             record.User,
		AffiliateProfile: profile,
	}, nil
}

func (r *UserRepository) findUserRecordByAccount(ctx context.Context, account string) (userRecord, error) {
	normalized := strings.TrimSpace(account)

	row := r.db.QueryRowContext(ctx, `
		select id, name, email, phone, password, role, status, email_verified_at, phone_verified_at, last_login_at, created_at, updated_at, deleted_at
		from users
		where lower(email) = $1 or phone = $2
		limit 1
	`, strings.ToLower(normalized), normalized)

	return scanUserRecord(row)
}

func (r *UserRepository) findUserRecordByID(ctx context.Context, userID int64) (userRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		select id, name, email, phone, password, role, status, email_verified_at, phone_verified_at, last_login_at, created_at, updated_at, deleted_at
		from users
		where id = $1
		limit 1
	`, userID)

	return scanUserRecord(row)
}

func (r *UserRepository) insertUser(ctx context.Context, tx *sql.Tx, params RegisterUserParams) (auth.User, error) {
	row := tx.QueryRowContext(ctx, `
		insert into users (name, email, phone, password, role, status, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, now(), now())
		returning id, name, email, phone, role, status, email_verified_at, phone_verified_at, last_login_at, created_at, updated_at
	`, params.Name, params.Email, params.Phone, params.PasswordHash, user.RoleClient, user.StatusActive)

	var result auth.User
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Email,
		&result.Phone,
		&result.Role,
		&result.Status,
		&result.EmailVerifiedAt,
		&result.PhoneVerifiedAt,
		&result.LastLoginAt,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return auth.User{}, err
	}

	return result, nil
}

func (r *UserRepository) insertDefaultWallets(ctx context.Context, tx *sql.Tx, userID int64) error {
	if _, err := tx.ExecContext(ctx, `
		insert into wallets (user_id, unit, balance, locked_balance, status, created_at, updated_at)
		values ($1, $2, 0, 0, $3, now(), now())
	`, userID, user.WalletUnitVND, user.WalletStatusActive); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, `
		insert into wallets (user_id, unit, balance, locked_balance, status, created_at, updated_at)
		values ($1, $2, 0, 0, $3, now(), now())
	`, userID, user.WalletUnitUSDT, user.WalletStatusActive)

	return err
}

func (r *UserRepository) insertAffiliateProfile(ctx context.Context, tx *sql.Tx, userID int64, registerURL string) (*auth.AffiliateProfile, error) {
	baseURL := strings.TrimSpace(registerURL)
	if baseURL == "" {
		baseURL = "http://localhost:3000/register"
	}

	for range 10 {
		refCode := strings.ToUpper(id.New()[:10])
		refLink := fmt.Sprintf("%s?ref_code=%s", strings.TrimRight(baseURL, "/"), refCode)

		row := tx.QueryRowContext(ctx, `
			insert into affiliate_profiles (user_id, ref_code, ref_link, status, created_at, updated_at)
			values ($1, $2, $3, $4, now(), now())
			on conflict (ref_code) do nothing
			returning id, ref_code, ref_link, status
		`, userID, refCode, refLink, user.AffiliateProfileStatusActive)

		var profile auth.AffiliateProfile
		err := row.Scan(&profile.ID, &profile.RefCode, &profile.RefLink, &profile.Status)
		if err == nil {
			return &profile, nil
		}

		if errors.Is(err, sql.ErrNoRows) {
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf(message.CannotGenerateAffiliate)
}

func (r *UserRepository) insertAffiliateReferral(ctx context.Context, tx *sql.Tx, refCode string, referredUserID int64) error {
	var affiliateProfileID int64
	var referrerUserID int64

	err := tx.QueryRowContext(ctx, `
		select id, user_id
		from affiliate_profiles
		where ref_code = $1
		limit 1
	`, refCode).Scan(&affiliateProfileID, &referrerUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrRefCodeNotFound
		}

		return err
	}

	if referrerUserID == referredUserID {
		return ErrInvalidSelfReferral
	}

	_, err = tx.ExecContext(ctx, `
		insert into affiliate_referrals (
			affiliate_profile_id,
			referrer_user_id,
			referred_user_id,
			status,
			created_at,
			updated_at
		)
		values ($1, $2, $3, $4, now(), now())
	`, affiliateProfileID, referrerUserID, referredUserID, user.AffiliateReferralStatusPending)

	return err
}

func (r *UserRepository) updateLastLoginAt(ctx context.Context, tx *sql.Tx, userID int64, at time.Time) error {
	_, err := tx.ExecContext(ctx, `update users set last_login_at = $1, updated_at = $1 where id = $2`, at, userID)
	return err
}

func (r *UserRepository) findAffiliateProfileByUserID(ctx context.Context, userID int64) (*auth.AffiliateProfile, error) {
	row := r.db.QueryRowContext(ctx, `
		select id, ref_code, ref_link, status
		from affiliate_profiles
		where user_id = $1
		limit 1
	`, userID)

	var profile auth.AffiliateProfile
	if err := row.Scan(&profile.ID, &profile.RefCode, &profile.RefLink, &profile.Status); err != nil {
		return nil, err
	}

	return &profile, nil
}

func (r *UserRepository) emailExists(ctx context.Context, email string) (bool, error) {
	return r.exists(ctx, `select exists(select 1 from users where lower(email) = $1)`, strings.ToLower(strings.TrimSpace(email)))
}

func (r *UserRepository) phoneExists(ctx context.Context, phone string) (bool, error) {
	return r.exists(ctx, `select exists(select 1 from users where phone = $1)`, strings.TrimSpace(phone))
}

func (r *UserRepository) exists(ctx context.Context, query string, arg string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, query, arg).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func scanUserRecord(row *sql.Row) (userRecord, error) {
	var result userRecord
	err := row.Scan(
		&result.ID,
		&result.Name,
		&result.Email,
		&result.Phone,
		&result.PasswordHash,
		&result.Role,
		&result.Status,
		&result.EmailVerifiedAt,
		&result.PhoneVerifiedAt,
		&result.LastLoginAt,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userRecord{}, ErrAccountNotFound
		}

		return userRecord{}, err
	}

	return result, nil
}
