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

	"gin/internal/support/clock"
	"gin/internal/support/message"
)

var (
	ErrDepositNotFound         = errors.New(message.DepositIntentNotFound)
	ErrDepositAlreadyDone      = errors.New(message.DepositAlreadyCompleted)
	ErrDepositProviderInvalid  = errors.New(message.DepositProviderInvalid)
	ErrDepositReceivingAccount = errors.New(message.DepositReceivingAccountMissing)
	ErrDepositWalletNotFound   = errors.New(message.DepositWalletMissing)
	ErrDepositAmountInvalid    = errors.New(message.DepositAmountInvalid)
)

type DepositRepository struct {
	db *sql.DB
}

type ReceivingAccountRecord struct {
	ID            int64
	Type          int
	Unit          int
	ProviderCode  *string
	AccountName   *string
	AccountNumber *string
	Status        int
	IsDefault     bool
	SortOrder     int
}

type VietQrBankRecord struct {
	ProviderCode string
	ShortName    string
	Name         string
	Bin          string
	Logo         *string
	AccountCount int
	IsDefault    bool
}

type DepositTransactionRecord struct {
	ID                 int64
	UserID             int64
	WalletID           int64
	ClientRef          string
	Unit               int
	Type               int
	Amount             string
	NetAmount          string
	Status             int
	Provider           string
	ProviderTxnID      *string
	ReceivingAccountID *int64
	Meta               []byte
	ReasonFailed       *string
	ApprovedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ReceivingAccount   *ReceivingAccountRecord
}

type DepositApplyResult struct {
	Transaction   DepositTransactionRecord
	Applied       bool
	AlreadyDone   bool
	WalletBalance string
}

func NewDepositRepository(db *sql.DB) *DepositRepository {
	return &DepositRepository{db: db}
}

func (r *DepositRepository) ListActiveReceivingAccounts(ctx context.Context) ([]ReceivingAccountRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		select id, type, unit, provider_code, account_name, account_number,
		       status, is_default, sort_order
		from payment_receiving_accounts
		where deleted_at is null and status = $1
		order by is_default desc, unit asc, type asc, sort_order asc, id asc
	`, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []ReceivingAccountRecord
	for rows.Next() {
		var record ReceivingAccountRecord
		if err := rows.Scan(
			&record.ID,
			&record.Type,
			&record.Unit,
			&record.ProviderCode,
			&record.AccountName,
			&record.AccountNumber,
			&record.Status,
			&record.IsDefault,
			&record.SortOrder,
		); err != nil {
			return nil, err
		}

		accounts = append(accounts, record)
	}

	return accounts, rows.Err()
}

func (r *DepositRepository) ListVietQrBanks(ctx context.Context) ([]VietQrBankRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		select
			p.provider_code,
			coalesce(b.short_name, p.provider_code) as short_name,
			coalesce(b.name, p.provider_code) as name,
			coalesce(b.bin, '') as bin,
			b.logo,
			count(*) as account_count,
			max(case when p.is_default then 1 else 0 end) as is_default
		from payment_receiving_accounts p
		left join vietqr_banks b on b.code = p.provider_code
		where p.deleted_at is null
		  and p.status = $1
		  and p.type = $2
		  and p.unit = $3
		group by p.provider_code, b.short_name, b.name, b.bin, b.logo
		order by is_default desc, account_count desc, short_name asc, p.provider_code asc
	`, 1, 1, 1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banks []VietQrBankRecord
	for rows.Next() {
		var record VietQrBankRecord
		var logo sql.NullString
		var isDefault int

		if err := rows.Scan(
			&record.ProviderCode,
			&record.ShortName,
			&record.Name,
			&record.Bin,
			&logo,
			&record.AccountCount,
			&isDefault,
		); err != nil {
			return nil, err
		}

		if logo.Valid {
			record.Logo = &logo.String
		}
		record.IsDefault = isDefault > 0
		banks = append(banks, record)
	}

	return banks, rows.Err()
}

func (r *DepositRepository) FindWalletByUserAndUnit(ctx context.Context, userID int64, unit int) (walletID int64, balance string, err error) {
	row := r.db.QueryRowContext(ctx, `
		select id, balance::text
		from wallets
		where user_id = $1 and unit = $2 and status = $3
		limit 1
	`, userID, unit, 1)

	if err := row.Scan(&walletID, &balance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrDepositWalletNotFound
		}

		return 0, "", err
	}

	return walletID, balance, nil
}

func (r *DepositRepository) CreateDepositIntent(ctx context.Context, params CreateDepositIntentParams) (DepositTransactionRecord, error) {
	var record DepositTransactionRecord
	metaJSON, err := marshalJSON(params.Meta)
	if err != nil {
		return DepositTransactionRecord{}, err
	}

	row := r.db.QueryRowContext(ctx, `
		insert into transactions (
			user_id, wallet_id, client_ref, unit, type, amount, fee, net_amount,
			status, provider, provider_txn_id, receiving_account_id, meta,
			created_at, updated_at
		)
		values ($1, $2, $3, $4, $5, $6::numeric(20,8), 0, $6::numeric(20,8), $7, $8, $9, $10, $11, now(), now())
		on conflict (client_ref) do update
		set updated_at = excluded.updated_at
		returning id, user_id, wallet_id, client_ref, unit, type, amount::text, net_amount::text,
		          status, provider, provider_txn_id, receiving_account_id, meta, reason_failed,
		          approved_by, approved_at, created_at, updated_at
	`, params.UserID, params.WalletID, params.ClientRef, params.Unit, params.Type, params.Amount, params.Status, params.Provider, params.ProviderTxnID, params.ReceivingAccountID, metaJSON)

	if err := scanDepositTransaction(row, &record); err != nil {
		return DepositTransactionRecord{}, err
	}

	return record, nil
}

func (r *DepositRepository) FindDepositIntentByClientRef(ctx context.Context, clientRef string) (DepositTransactionRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		select t.id, t.user_id, t.wallet_id, t.client_ref, t.unit, t.type, t.amount::text, t.net_amount::text,
		       t.status, t.provider, t.provider_txn_id, t.receiving_account_id, t.meta, t.reason_failed,
		       t.approved_by, t.approved_at, t.created_at, t.updated_at,
		       p.id, p.type, p.unit, p.provider_code, p.account_name, p.account_number,
		       p.status, p.is_default, p.sort_order
		from transactions t
		left join payment_receiving_accounts p on p.id = t.receiving_account_id
		where t.client_ref = $1 and t.deleted_at is null
		limit 1
	`, clientRef)

	var record DepositTransactionRecord
	if err := scanDepositTransactionWithAccount(row, &record); err != nil {
		return DepositTransactionRecord{}, err
	}

	return record, nil
}

func (r *DepositRepository) FindDepositIntentByProviderTxnID(ctx context.Context, provider, providerTxnID string) (DepositTransactionRecord, error) {
	row := r.db.QueryRowContext(ctx, `
		select t.id, t.user_id, t.wallet_id, t.client_ref, t.unit, t.type, t.amount::text, t.net_amount::text,
		       t.status, t.provider, t.provider_txn_id, t.receiving_account_id, t.meta, t.reason_failed,
		       t.approved_by, t.approved_at, t.created_at, t.updated_at,
		       p.id, p.type, p.unit, p.provider_code, p.account_name, p.account_number,
		       p.status, p.is_default, p.sort_order
		from transactions t
		left join payment_receiving_accounts p on p.id = t.receiving_account_id
		where t.provider = $1 and t.provider_txn_id = $2 and t.deleted_at is null
		limit 1
	`, provider, providerTxnID)

	var record DepositTransactionRecord
	if err := scanDepositTransactionWithAccount(row, &record); err != nil {
		return DepositTransactionRecord{}, err
	}

	return record, nil
}

var ErrDepositCancelForbidden = errors.New("giao dịch không thể hủy ở trạng thái hiện tại")

// CancelDeposit hủy một lệnh nạp tiền đang chờ xử lý (PENDING/CONFIRMED) do người dùng yêu cầu.
func (r *DepositRepository) CancelDeposit(ctx context.Context, userID, txnID int64) (DepositTransactionRecord, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return DepositTransactionRecord{}, err
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowContext(ctx, `
		select id, user_id, wallet_id, client_ref, unit, type, amount::text, net_amount::text,
		       status, provider, provider_txn_id, receiving_account_id, meta, reason_failed,
		       approved_by, approved_at, created_at, updated_at
		from transactions
		where id = $1 and user_id = $2 and type = 1 and deleted_at is null
		limit 1
		for update
	`, txnID, userID)

	var record DepositTransactionRecord
	if err := scanDepositTransaction(row, &record); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return DepositTransactionRecord{}, ErrDepositNotFound
		}
		return DepositTransactionRecord{}, err
	}

	// Chỉ hủy khi đang PENDING (1) hoặc CONFIRMED (2)
	if record.Status != 1 && record.Status != 2 {
		return DepositTransactionRecord{}, ErrDepositCancelForbidden
	}

	if _, err := tx.ExecContext(ctx, `
		update transactions set status = 5, reason_failed = 'Người dùng tự hủy', updated_at = now()
		where id = $1
	`, record.ID); err != nil {
		return DepositTransactionRecord{}, err
	}

	if err := tx.Commit(); err != nil {
		return DepositTransactionRecord{}, err
	}

	record.Status = 5
	return record, nil
}

func (r *DepositRepository) ApplyDeposit(ctx context.Context, params ApplyDepositParams) (DepositApplyResult, error) {
	if strings.TrimSpace(params.ClientRef) == "" && strings.TrimSpace(params.ProviderTxnID) == "" {
		return DepositApplyResult{}, ErrDepositNotFound
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return DepositApplyResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	record, err := r.findDepositIntentForUpdate(ctx, tx, params)
	if err != nil {
		return DepositApplyResult{}, err
	}

	if record.Status == 3 {
		result := DepositApplyResult{Transaction: record, Applied: false, AlreadyDone: true}
		if err := tx.Commit(); err != nil {
			return DepositApplyResult{}, err
		}
		return result, nil
	}

	if record.Status == 4 {
		result := DepositApplyResult{Transaction: record, Applied: false, AlreadyDone: false}
		if err := tx.Commit(); err != nil {
			return DepositApplyResult{}, err
		}
		return result, nil
	}

	success, failed := classifyProviderStatus(params.ProviderStatus)
	now := params.PaidAt
	if now.IsZero() {
		now = clock.Now()
	}

	metaJSON, err := marshalJSON(params.Raw)
	if err != nil {
		return DepositApplyResult{}, err
	}

	if params.ProviderTxnID != "" {
		if record.ProviderTxnID != nil && *record.ProviderTxnID != "" && *record.ProviderTxnID != params.ProviderTxnID {
			return DepositApplyResult{}, fmt.Errorf(message.DepositProviderTxnIDMismatch)
		}

		if _, err := tx.ExecContext(ctx, `
			update transactions
			set provider_txn_id = coalesce(nullif($1, ''), provider_txn_id),
			    meta = coalesce($2::json, meta),
			    updated_at = now()
			where id = $3
		`, params.ProviderTxnID, metaJSON, record.ID); err != nil {
			return DepositApplyResult{}, err
		}

		record.ProviderTxnID = stringPtr(params.ProviderTxnID, record.ProviderTxnID)
	}

	if failed {
		reason := params.ProviderStatus
		if reason == "" {
			reason = "failed"
		}

		if _, err := tx.ExecContext(ctx, `
			update transactions
			set status = $1, reason_failed = $2, meta = coalesce($3::json, meta), updated_at = now()
			where id = $4
		`, 4, reason, metaJSON, record.ID); err != nil {
			return DepositApplyResult{}, err
		}

		if err := tx.Commit(); err != nil {
			return DepositApplyResult{}, err
		}

		record.Status = 4
		record.ReasonFailed = &reason
		record.Meta = metaJSON
		return DepositApplyResult{Transaction: record, Applied: false, AlreadyDone: false}, nil
	}

	if !success {
		if _, err := tx.ExecContext(ctx, `
			update transactions
			set meta = coalesce($1::json, meta), updated_at = now()
			where id = $2
		`, metaJSON, record.ID); err != nil {
			return DepositApplyResult{}, err
		}

		if err := tx.Commit(); err != nil {
			return DepositApplyResult{}, err
		}

		record.Meta = metaJSON
		return DepositApplyResult{Transaction: record, Applied: false, AlreadyDone: false}, nil
	}

	walletID, balanceBefore, err := r.lockWalletForUpdate(ctx, tx, record.WalletID)
	if err != nil {
		return DepositApplyResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		update wallets
		set balance = balance + $1::numeric(20,8), updated_at = now()
		where id = $2
	`, params.Amount, walletID); err != nil {
		return DepositApplyResult{}, err
	}

	bAmount, ok1 := new(big.Float).SetString(params.Amount)
	bBefore, ok2 := new(big.Float).SetString(balanceBefore)
	if !ok1 || !ok2 {
		return DepositApplyResult{}, fmt.Errorf("failed to parse balance or amount as numeric: amount=%q balance=%q", params.Amount, balanceBefore)
	}

	bAfter := new(big.Float).Add(bBefore, bAmount)
	balanceAfter := bAfter.Text('f', 8)

	if _, err := tx.ExecContext(ctx, `
		insert into wallet_ledger_entries (
			wallet_id, user_id, direction, amount, balance_before, balance_after,
			reference_type, reference_id, note, created_at
		)
		values ($1, $2, $3, $4, $5, $6,
		        $7, $8, $9, now())
	`, walletID, record.UserID, 1, params.Amount, balanceBefore, balanceAfter,
		"App\\Models\\Transaction\\Transaction", record.ID, "Nạp tiền thành công"); err != nil {
		return DepositApplyResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
		update transactions
		set status = $1,
		    provider_txn_id = coalesce(nullif($2, ''), provider_txn_id),
		    meta = coalesce($3::json, meta),
		    approved_at = $4,
		    updated_at = now()
		where id = $5
	`, 3, params.ProviderTxnID, metaJSON, now, record.ID); err != nil {
		return DepositApplyResult{}, err
	}

	if err := r.qualifyReferral(ctx, tx, record.UserID, record.ID, params.Amount, now); err != nil {
		return DepositApplyResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return DepositApplyResult{}, err
	}

	record.Status = 3
	record.ProviderTxnID = stringPtr(params.ProviderTxnID, record.ProviderTxnID)
	record.Meta = metaJSON
	record.ApprovedAt = &now

	return DepositApplyResult{
		Transaction:   record,
		Applied:       true,
		AlreadyDone:   false,
		WalletBalance: fmt.Sprintf("%s", params.Amount),
	}, nil
}

type CreateDepositIntentParams struct {
	UserID             int64
	WalletID           int64
	ClientRef          string
	Unit               int
	Type               int
	Amount             string
	Status             int
	Provider           string
	ProviderTxnID      *string
	ReceivingAccountID *int64
	Meta               map[string]any
}

type ApplyDepositParams struct {
	Provider       string
	ProviderStatus string
	ClientRef      string
	ProviderTxnID  string
	Amount         string
	Currency       string
	PaidAt         time.Time
	Raw            map[string]any
}

func (r *DepositRepository) findDepositIntentForUpdate(ctx context.Context, tx *sql.Tx, params ApplyDepositParams) (DepositTransactionRecord, error) {
	query := `
		select id, user_id, wallet_id, client_ref, unit, type, amount::text, net_amount::text,
		       status, provider, provider_txn_id, receiving_account_id, meta, reason_failed,
		       approved_by, approved_at, created_at, updated_at
		from transactions
		where client_ref = $1 and deleted_at is null
		limit 1
		for update
	`
	row := tx.QueryRowContext(ctx, query, params.ClientRef)

	var record DepositTransactionRecord
	if err := scanDepositTransaction(row, &record); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if strings.TrimSpace(params.ProviderTxnID) != "" {
				row = tx.QueryRowContext(ctx, `
					select id, user_id, wallet_id, client_ref, unit, type, amount::text, net_amount::text,
					       status, provider, provider_txn_id, receiving_account_id, meta, reason_failed,
					       approved_by, approved_at, created_at, updated_at
					from transactions
					where provider = $1 and provider_txn_id = $2 and deleted_at is null
					limit 1
					for update
				`, params.Provider, params.ProviderTxnID)
				if err := scanDepositTransaction(row, &record); err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						return DepositTransactionRecord{}, ErrDepositNotFound
					}
					return DepositTransactionRecord{}, err
				}
				return record, nil
			}

			return DepositTransactionRecord{}, ErrDepositNotFound
		}

		return DepositTransactionRecord{}, err
	}

	return record, nil
}

func (r *DepositRepository) lockWalletForUpdate(ctx context.Context, tx *sql.Tx, walletID int64) (int64, string, error) {
	row := tx.QueryRowContext(ctx, `
		select id, balance::text
		from wallets
		where id = $1
		limit 1
		for update
	`, walletID)

	var balance string
	if err := row.Scan(&walletID, &balance); err != nil {
		return 0, "", err
	}

	return walletID, balance, nil
}

func (r *DepositRepository) qualifyReferral(ctx context.Context, tx *sql.Tx, userID int64, transactionID int64, amount string, at time.Time) error {
	_, err := tx.ExecContext(ctx, `
		update affiliate_referrals
		set first_deposit_transaction_id = CAST($1 AS BIGINT),
		    first_deposit_amount = CAST($2 AS NUMERIC),
		    qualified_at = $3,
		    status = CAST($4 AS INT),
		    updated_at = $3
		where referred_user_id = CAST($5 AS BIGINT)
		  and first_deposit_transaction_id is null
		  and status = CAST($6 AS INT)
		  and CAST($2 AS NUMERIC) >= CAST($7 AS NUMERIC)
	`, transactionID, amount, at, 2, userID, 1, "50000")
	return err
}

func scanDepositTransaction(row *sql.Row, record *DepositTransactionRecord) error {
	var (
		providerTxnID      sql.NullString
		receivingAccountID sql.NullInt64
		metaJSON           []byte
		reasonFailed       sql.NullString
		approvedBy         sql.NullInt64
		approvedAt         sql.NullTime
	)

	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.WalletID,
		&record.ClientRef,
		&record.Unit,
		&record.Type,
		&record.Amount,
		&record.NetAmount,
		&record.Status,
		&record.Provider,
		&providerTxnID,
		&receivingAccountID,
		&metaJSON,
		&reasonFailed,
		&approvedBy,
		&approvedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return err
	}

	if providerTxnID.Valid {
		record.ProviderTxnID = &providerTxnID.String
	}
	if receivingAccountID.Valid {
		value := receivingAccountID.Int64
		record.ReceivingAccountID = &value
	}
	if len(metaJSON) > 0 {
		record.Meta = metaJSON
	}
	if reasonFailed.Valid {
		record.ReasonFailed = &reasonFailed.String
	}
	if approvedAt.Valid {
		record.ApprovedAt = &approvedAt.Time
	}
	if approvedBy.Valid {
		// approved_by intentionally ignored for deposit response payload
	}

	return nil
}

func scanDepositTransactionWithAccount(row *sql.Row, record *DepositTransactionRecord) error {
	var (
		account             ReceivingAccountRecord
		providerTxnID       sql.NullString
		receivingAccountID  sql.NullInt64
		metaJSON            []byte
		reasonFailed        sql.NullString
		approvedBy          sql.NullInt64
		approvedAt          sql.NullTime
		accountID           sql.NullInt64
		accountType         sql.NullInt64
		accountUnit         sql.NullInt64
		accountProviderCode sql.NullString
		accountName         sql.NullString
		accountNumber       sql.NullString
		accountStatus       sql.NullInt64
		accountIsDefault    sql.NullBool
		accountSortOrder    sql.NullInt64
	)

	if err := row.Scan(
		&record.ID,
		&record.UserID,
		&record.WalletID,
		&record.ClientRef,
		&record.Unit,
		&record.Type,
		&record.Amount,
		&record.NetAmount,
		&record.Status,
		&record.Provider,
		&providerTxnID,
		&receivingAccountID,
		&metaJSON,
		&reasonFailed,
		&approvedBy,
		&approvedAt,
		&record.CreatedAt,
		&record.UpdatedAt,
		&accountID,
		&accountType,
		&accountUnit,
		&accountProviderCode,
		&accountName,
		&accountNumber,
		&accountStatus,
		&accountIsDefault,
		&accountSortOrder,
	); err != nil {
		return err
	}

	if providerTxnID.Valid {
		record.ProviderTxnID = &providerTxnID.String
	}
	if receivingAccountID.Valid {
		value := receivingAccountID.Int64
		record.ReceivingAccountID = &value
	}
	if len(metaJSON) > 0 {
		record.Meta = metaJSON
	}
	if reasonFailed.Valid {
		record.ReasonFailed = &reasonFailed.String
	}
	if approvedAt.Valid {
		record.ApprovedAt = &approvedAt.Time
	}
	if accountID.Valid && accountID.Int64 > 0 {
		account.ID = accountID.Int64
		account.Type = int(accountType.Int64)
		account.Unit = int(accountUnit.Int64)
		account.ProviderCode = nullStringPtr(accountProviderCode)
		account.AccountName = nullStringPtr(accountName)
		account.AccountNumber = nullStringPtr(accountNumber)
		account.Status = int(accountStatus.Int64)
		account.IsDefault = accountIsDefault.Bool
		account.SortOrder = int(accountSortOrder.Int64)
		record.ReceivingAccount = &account
	}

	return nil
}

func marshalJSON(value map[string]any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	return json.Marshal(value)
}

func classifyProviderStatus(status string) (success bool, failed bool) {
	normalized := strings.ToLower(strings.TrimSpace(status))
	switch normalized {
	case "success", "succeeded", "paid", "completed", "confirmed", "done", "ok", "1", "00", "finished":
		return true, false
	case "failed", "canceled", "cancelled", "rejected", "error", "0", "expired", "refunded":
		return false, true
	default:
		return false, false
	}
}

func nullStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	v := value.String
	return &v
}

func stringPtr(value string, fallback *string) *string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	v := value
	return &v
}
