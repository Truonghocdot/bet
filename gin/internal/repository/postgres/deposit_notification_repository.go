package postgres

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

type DepositNotificationRecord struct {
	TransactionID        int64
	UserID               int64
	UserName             string
	UserPhone            string
	ClientRef            string
	ProviderTxnID        string
	Provider             string
	Amount               string
	Status               int
	CreatedAt            time.Time
	ReceivingBank        string
	ReceivingAccountName string
	ReceivingAccount     string
}

func (r *DepositRepository) FindDepositForNotification(ctx context.Context, provider, providerTxnID, clientRef string) (DepositNotificationRecord, error) {
	var record DepositNotificationRecord
	var providerTxn, phone, bank, account sql.NullString
	row := r.db.QueryRowContext(ctx, `
		select t.id, t.user_id, coalesce(u.name, ''), coalesce(u.phone, ''),
		       t.client_ref, t.provider_txn_id, t.provider, t.amount::text, t.status,
		       t.created_at, coalesce(p.provider_code, ''), coalesce(p.account_name, ''), coalesce(p.account_number, '')
		from transactions t
		join users u on u.id = t.user_id
		left join payment_receiving_accounts p on p.id = t.receiving_account_id
		where t.type = 1
		  and t.deleted_at is null
		  and (
		      ($1 <> '' and t.provider = $1 and t.provider_txn_id = $2)
		      or ($3 <> '' and t.client_ref = $3 and t.provider = $1)
		  )
		order by case when $1 <> '' and t.provider = $1 and t.provider_txn_id = $2 then 0 else 1 end, t.id desc
		limit 1
	`, strings.TrimSpace(provider), strings.TrimSpace(providerTxnID), strings.TrimSpace(clientRef))

	if err := row.Scan(
		&record.TransactionID,
		&record.UserID,
		&record.UserName,
		&phone,
		&record.ClientRef,
		&providerTxn,
		&record.Provider,
		&record.Amount,
		&record.Status,
		&record.CreatedAt,
		&bank,
		&record.ReceivingAccountName,
		&account,
	); err != nil {
		return DepositNotificationRecord{}, err
	}
	record.UserPhone = phone.String
	record.ProviderTxnID = providerTxn.String
	record.ReceivingBank = bank.String
	record.ReceivingAccount = account.String

	return record, nil
}
