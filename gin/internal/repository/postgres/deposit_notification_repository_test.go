package postgres

import (
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFindDepositForNotificationUsesReadOnlyJoinAndLookupKeys(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	createdAt := time.Date(2026, time.September, 1, 7, 30, 12, 0, time.UTC)
	mock.ExpectQuery("select t.id").WithArgs("sepay_vietqr", "FT-1", "DEP-abc").WillReturnRows(
		sqlmock.NewRows([]string{"id", "user_id", "name", "phone", "client_ref", "provider_txn_id", "provider", "amount", "status", "created_at", "provider_code", "account_name", "account_number"}).
			AddRow(int64(10), int64(123), "Nguyen Van A", "0900000000", "DEP-abc", "FT-1", "sepay_vietqr", "50000", 1, createdAt, "MBBank", "CONG TY A", "0123456789"),
	)

	repository := NewDepositRepository(db)
	record, err := repository.FindDepositForNotification(t.Context(), "sepay_vietqr", "FT-1", "DEP-abc")
	if err != nil {
		t.Fatalf("lookup deposit: %v", err)
	}
	if record.TransactionID != 10 || record.UserID != 123 || record.Amount != "50000" {
		t.Fatalf("unexpected lookup record: %#v", record)
	}
	if !strings.EqualFold(record.ReceivingBank, "MBBank") || record.ReceivingAccountName != "CONG TY A" || record.ReceivingAccount != "0123456789" {
		t.Fatalf("unexpected receiving account: %#v", record)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
