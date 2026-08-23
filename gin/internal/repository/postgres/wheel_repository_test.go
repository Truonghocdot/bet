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
	if wheelTotalRounds != 3 {
		t.Fatalf("wheel event must end after three rounds, got %d", wheelTotalRounds)
	}
	if positiveNumeric("0.00000000") || !positiveNumeric("50000000.00000000") {
		t.Fatal("positiveNumeric must distinguish zero and positive rewards")
	}
	value, err := randomUUID()
	if err != nil || len(value) != 36 || value[14] != '4' {
		t.Fatalf("unexpected UUID v4: %q, %v", value, err)
	}
}

func TestWheelLaunchActivatesFiveMinuteBotWindow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectExec("update chat_rooms as cr").
		WithArgs(int64(11), int64(203985), 300).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewWheelRepository(db)
	if err := repository.ActivateInvitationChat(t.Context(), 11, 203985, 300); err != nil {
		t.Fatalf("activate invitation chat: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestWheelCreateChatAllowsPendingInvitationRoom(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Now().UTC()
	mock.ExpectBegin()
	mock.ExpectQuery("select cr.id, wi.status, ws.id, ws.status, ws.ends_at").
		WithArgs(int64(11), int64(203985)).
		WillReturnRows(sqlmock.NewRows([]string{"room_id", "invitation_status", "session_id", "session_status", "ends_at"}).AddRow(int64(44), "pending", nil, nil, nil))
	mock.ExpectQuery("select exists\\(select 1 from chat_bans").
		WithArgs(int64(203985), int64(44)).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectQuery("select body from chat_messages").
		WithArgs(int64(44), int64(203985)).
		WillReturnRows(sqlmock.NewRows([]string{"body"}))
	mock.ExpectExec("insert into chat_user_profiles").
		WithArgs(int64(203985), "ID game #203985").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("insert into chat_messages").
		WithArgs(int64(44), int64(203985), "ID game #203985", "Chào phòng sự kiện", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow(int64(99), now))
	mock.ExpectQuery("insert into wheel_outbox_events").
		WithArgs("stream:wheel:invitation:11", "chat.message.created", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(501)))
	mock.ExpectCommit()

	repository := NewWheelRepository(db)
	message, sessionID, err := repository.CreateChat(t.Context(), 11, 203985, "Chào phòng sự kiện")
	if err != nil {
		t.Fatalf("create pending chat: %v", err)
	}
	if sessionID != 0 || message.ID != 99 || message.DisplayName != "ID game #203985" {
		t.Fatalf("unexpected result: session=%d message=%#v", sessionID, message)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
