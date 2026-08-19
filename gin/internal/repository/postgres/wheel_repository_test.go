package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWheelSpinRejectsNextRoundBeforeFiveSecondGate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery("select id, status, current_round, ends_at from wheel_sessions").
		WithArgs(int64(11), int64(203985)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "current_round", "ends_at"}).AddRow(int64(81), "active", 2, time.Now().Add(5*time.Minute)))
	mock.ExpectQuery("select id, status, segment_key, result_label, prize_amount::text, spun_at from wheel_invitation_rounds").
		WithArgs(int64(11), 2).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "segment_key", "result_label", "prize_amount", "spun_at"}).AddRow(int64(102), "pending", "jackpot_50m", "Giải 50 triệu", "50000000", nil))
	previousSpunAt := time.Now().Add(-2 * time.Second)
	mock.ExpectQuery("select spun_at from wheel_invitation_rounds").
		WithArgs(int64(11), 1).
		WillReturnRows(sqlmock.NewRows([]string{"spun_at"}).AddRow(previousSpunAt))
	mock.ExpectRollback()

	repository := NewWheelRepository(db)
	_, err = repository.Spin(t.Context(), 11, 203985, 2, 5, "fh88u")
	if !errors.Is(err, ErrWheelRoundNotReady) {
		t.Fatalf("expected round-not-ready error, got %v", err)
	}
	var notReady *WheelNotReadyError
	if !errors.As(err, &notReady) || !notReady.AvailableAt.Equal(previousSpunAt.Add(5*time.Second)) {
		t.Fatalf("expected exact available_at, got %#v", notReady)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestWheelSpinRejectsSkippedRound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery("select id, status, current_round, ends_at from wheel_sessions").
		WithArgs(int64(12), int64(203986)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "current_round", "ends_at"}).AddRow(int64(82), "active", 2, time.Now().Add(5*time.Minute)))
	mock.ExpectQuery("select id, status, segment_key, result_label, prize_amount::text, spun_at from wheel_invitation_rounds").
		WithArgs(int64(12), 3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "segment_key", "result_label", "prize_amount", "spun_at"}).AddRow(int64(203), "pending", "try_again", "May mắn lần sau", "0", nil))
	mock.ExpectRollback()

	repository := NewWheelRepository(db)
	_, err = repository.Spin(t.Context(), 12, 203986, 3, 5, "fh88u")
	if !errors.Is(err, ErrWheelRoundOrder) {
		t.Fatalf("expected invalid-order error, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestWheelHelpersPreserveMoneyAndUUIDShape(t *testing.T) {
	if positiveNumeric("0.00000000") || !positiveNumeric("50000000.00000000") {
		t.Fatal("positiveNumeric must distinguish zero and positive rewards")
	}
	value, err := randomUUID()
	if err != nil || len(value) != 36 || value[14] != '4' {
		t.Fatalf("unexpected UUID v4: %q, %v", value, err)
	}
}
