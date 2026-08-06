package service

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	repopg "gin/internal/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	goredis "github.com/redis/go-redis/v9"
)

func TestRunTickSafelyConvertsPanicToError(t *testing.T) {
	testDiscardRoomEngineLogs(t)
	service := &RoomEngineService{}

	err := service.runTickSafely(t.Context())
	if err == nil {
		t.Fatal("runTickSafely must return an error after a tick panic")
	}
	if !strings.Contains(err.Error(), "room engine tick panic") {
		t.Fatalf("unexpected panic error: %v", err)
	}
}

func TestRoomEngineSettlementKillSwitchSkipsRepository(t *testing.T) {
	service := &RoomEngineService{settlementEnabled: false}

	if err := service.settleDrawnPeriods(t.Context()); err != nil {
		t.Fatalf("disabled settlement returned an error: %v", err)
	}
}

func TestRoomEngineSettlementFailureIsRecordedAndDoesNotBlockNextPeriod(t *testing.T) {
	testDiscardRoomEngineLogs(t)
	redisClient := testRoomEngineRedisClient(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const (
		periodStatusDrawn   = 4
		periodStatusSettled = 5
	)
	firstPeriodID := int64(37090305)
	secondPeriodID := int64(37096785)
	drawAt := time.Date(2026, time.August, 6, 15, 3, 0, 0, time.UTC)
	mock.ExpectQuery(`select id, room_code, period_no, game_type, result_payload, draw_at\s+from game_periods`).
		WithArgs(periodStatusDrawn, 200).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_code", "period_no", "game_type", "result_payload", "draw_at"}).
			AddRow(firstPeriodID, "wingo_3m", "WINGO_3M_1786003380", 1, []byte(`not-json`), drawAt).
			AddRow(secondPeriodID, "wingo_3m", "WINGO_3M_1786003560", 1, []byte(`{"tags":["small"]}`), drawAt.Add(3*time.Minute)))

	mock.ExpectBegin()
	mock.ExpectQuery(`select status\s+from game_periods\s+where id = \$1\s+for update`).
		WithArgs(firstPeriodID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(periodStatusDrawn))
	mock.ExpectRollback()
	mock.ExpectExec(`update game_periods\s+set settlement_attempts`).
		WithArgs(sqlmock.AnyArg(), firstPeriodID, periodStatusDrawn).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectBegin()
	mock.ExpectQuery(`select status\s+from game_periods\s+where id = \$1\s+for update`).
		WithArgs(secondPeriodID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(periodStatusSettled))
	mock.ExpectRollback()

	service := &RoomEngineService{
		gameRepository:    repopg.NewGameRepository(db),
		redis:             redisClient,
		settlementEnabled: true,
	}
	if err := service.settleDrawnPeriods(t.Context()); err != nil {
		t.Fatalf("settle drawn periods: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestRoomEngineSettlementDeadlinePropagatesWithoutFailureRecord(t *testing.T) {
	testDiscardRoomEngineLogs(t)
	redisClient := testRoomEngineRedisClient(t)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	const periodStatusDrawn = 4
	periodID := int64(37090305)
	mock.ExpectQuery(`select id, room_code, period_no, game_type, result_payload, draw_at\s+from game_periods`).
		WithArgs(periodStatusDrawn, 200).
		WillReturnRows(sqlmock.NewRows([]string{"id", "room_code", "period_no", "game_type", "result_payload", "draw_at"}).
			AddRow(periodID, "wingo_3m", "WINGO_3M_1786003380", 1, []byte(`{"tags":["small"]}`), time.Now()))
	mock.ExpectBegin()
	mock.ExpectQuery(`select status\s+from game_periods\s+where id = \$1\s+for update`).
		WithArgs(periodID).
		WillReturnError(context.DeadlineExceeded)
	mock.ExpectRollback()

	service := &RoomEngineService{
		gameRepository:    repopg.NewGameRepository(db),
		redis:             redisClient,
		settlementEnabled: true,
	}
	err = service.settleDrawnPeriods(t.Context())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected settlement deadline to propagate, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestRoomEngineRunContinuesAfterTickPanicUntilCanceled(t *testing.T) {
	testDiscardRoomEngineLogs(t)

	service := &RoomEngineService{tickInterval: 5 * time.Millisecond}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() {
		done <- service.Run(ctx)
	}()

	select {
	case err := <-done:
		t.Fatalf("engine exited after a recovered tick panic: %v", err)
	case <-time.After(30 * time.Millisecond):
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("engine returned an error after cancellation: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("engine did not stop after cancellation")
	}
}

func testDiscardRoomEngineLogs(t *testing.T) {
	t.Helper()
	previousLogOutput := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(previousLogOutput) })
}

func TestRoomEngineLockReleaseRequiresCurrentOwnerToken(t *testing.T) {
	client := testRoomEngineRedisClient(t)
	service := &RoomEngineService{redis: client}
	ctx := t.Context()
	key := "test:engine:period:settle:37090305"

	ownerAToken, acquired, err := service.acquireLock(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("owner A acquire lock: %v", err)
	}
	if !acquired {
		t.Fatal("owner A must acquire an empty lock")
	}

	competingToken, acquired, err := service.acquireLock(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("competing acquire: %v", err)
	}
	if acquired {
		t.Fatal("a competing owner must not acquire a held lock")
	}
	if competingToken == "" || competingToken == ownerAToken {
		t.Fatalf("lock attempts must use unique non-empty tokens: owner=%q competing=%q", ownerAToken, competingToken)
	}

	const ownerBToken = "replacement-owner-token"
	if err := client.Set(ctx, key, ownerBToken, time.Minute).Err(); err != nil {
		t.Fatalf("replace expired lock with owner B: %v", err)
	}
	service.releaseLock(ctx, key, ownerAToken)

	currentOwner, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("read lock after stale release: %v", err)
	}
	if currentOwner != ownerBToken {
		t.Fatalf("stale owner A deleted owner B lock: got %q", currentOwner)
	}

	service.releaseLock(ctx, key, ownerBToken)
	if exists, err := client.Exists(ctx, key).Result(); err != nil {
		t.Fatalf("check released lock: %v", err)
	} else if exists != 0 {
		t.Fatalf("current owner release left %d lock keys", exists)
	}
}

func testRoomEngineRedisClient(t *testing.T) *goredis.Client {
	t.Helper()

	serverBinary, err := exec.LookPath("redis-server")
	if err != nil {
		t.Skip("redis-server is not installed")
	}

	tempDir := t.TempDir()
	socketPath := filepath.Join(tempDir, "redis.sock")
	command := exec.Command(
		serverBinary,
		"--port", "0",
		"--save", "",
		"--appendonly", "no",
		"--unixsocket", socketPath,
		"--unixsocketperm", "700",
	)
	command.Stdout = io.Discard
	command.Stderr = io.Discard
	if err := command.Start(); err != nil {
		t.Fatalf("start isolated redis-server: %v", err)
	}

	waitDone := make(chan struct{})
	var waitErr error
	go func() {
		waitErr = command.Wait()
		close(waitDone)
	}()
	t.Cleanup(func() {
		if command.Process != nil {
			_ = command.Process.Signal(os.Interrupt)
		}
		select {
		case <-waitDone:
		case <-time.After(time.Second):
			if command.Process != nil {
				_ = command.Process.Kill()
			}
			<-waitDone
		}
	})

	client := goredis.NewClient(&goredis.Options{Network: "unix", Addr: socketPath})
	t.Cleanup(func() { _ = client.Close() })

	deadline := time.Now().Add(2 * time.Second)
	for {
		if err := client.Ping(t.Context()).Err(); err == nil {
			return client
		}
		select {
		case <-waitDone:
			t.Fatalf("isolated redis-server exited before readiness: %v", waitErr)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("isolated redis-server did not become ready")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
