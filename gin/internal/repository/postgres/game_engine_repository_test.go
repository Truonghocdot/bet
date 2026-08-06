package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListDrawnPeriodsForSettlementDoesNotCollapseRoomQueue(t *testing.T) {
	queryMatcher := sqlmock.QueryMatcherFunc(func(_ string, actualSQL string) error {
		normalized := strings.ToLower(strings.Join(strings.Fields(actualSQL), " "))
		if strings.Contains(normalized, "distinct on") {
			return fmt.Errorf("settlement queue must not collapse periods per room: %s", normalized)
		}

		for _, fragment := range []string{
			"from game_periods",
			"where status = $1",
			"result_payload is not null",
			"(settlement_next_retry_at is null or settlement_next_retry_at <= now())",
			"order by draw_at asc, id asc",
			"limit $2",
		} {
			if !strings.Contains(normalized, fragment) {
				return fmt.Errorf("settlement queue query is missing %q: %s", fragment, normalized)
			}
		}
		return nil
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(queryMatcher))
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	drawAt := time.Date(2026, time.August, 6, 15, 3, 0, 0, time.UTC)
	columns := []string{"id", "room_code", "period_no", "game_type", "result_payload", "draw_at"}
	rows := sqlmock.NewRows(columns).
		AddRow(int64(37090305), "wingo_3m", "WINGO_3M_1786003380", 1, []byte(`{"tags":["small"]}`), drawAt).
		AddRow(int64(37096785), "wingo_3m", "WINGO_3M_1786003560", 1, []byte(`{"tags":["big"]}`), drawAt.Add(3*time.Minute))

	mock.ExpectQuery("settlement queue").
		WithArgs(periodStatusDrawn, 100).
		WillReturnRows(rows)

	repository := NewGameRepository(db)
	periods, err := repository.ListDrawnPeriodsForSettlement(t.Context(), 100)
	if err != nil {
		t.Fatalf("list periods: %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected every drawn period from the room, got %d", len(periods))
	}
	if periods[0].RoomCode != "wingo_3m" || periods[1].RoomCode != "wingo_3m" {
		t.Fatalf("expected both periods to remain in the same room queue: %#v", periods)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestRecordPeriodSettlementFailureAppliesCappedExponentialBackoff(t *testing.T) {
	queryMatcher := sqlmock.QueryMatcherFunc(func(_ string, actualSQL string) error {
		normalized := strings.ToLower(strings.Join(strings.Fields(actualSQL), " "))
		for _, fragment := range []string{
			"update game_periods",
			"settlement_attempts = least(coalesce(settlement_attempts, 0), 2147483646) + 1",
			"settlement_last_error = $1",
			"settlement_next_retry_at = now() +",
			"power(2, least(greatest(coalesce(settlement_attempts, 0), 0), 9))::integer",
			"* interval '1 second'",
			"where id = $2 and status = $3",
		} {
			if !strings.Contains(normalized, fragment) {
				return fmt.Errorf("settlement failure update is missing %q: %s", fragment, normalized)
			}
		}
		if !strings.Contains(normalized, "300,") {
			return fmt.Errorf("settlement retry delay must cap at 300 seconds: %s", normalized)
		}
		return nil
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(queryMatcher))
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	periodID := int64(37090305)
	longMessage := strings.Repeat("界", settlementFailureMaxErrorRunes+100)
	wantMessage := strings.Repeat("界", settlementFailureMaxErrorRunes)
	mock.ExpectExec("record settlement failure").
		WithArgs(wantMessage, periodID, periodStatusDrawn).
		WillReturnResult(sqlmock.NewResult(0, 1))

	repository := NewGameRepository(db)
	if err := repository.RecordPeriodSettlementFailure(t.Context(), periodID, fmt.Errorf("  %s  ", longMessage)); err != nil {
		t.Fatalf("record settlement failure: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestListLockedPeriodsForDrawDoesNotCollapseRoomQueue(t *testing.T) {
	queryMatcher := sqlmock.QueryMatcherFunc(func(_ string, actualSQL string) error {
		normalized := strings.ToLower(strings.Join(strings.Fields(actualSQL), " "))
		if strings.Contains(normalized, "distinct on") {
			return fmt.Errorf("draw queue must not collapse periods per room: %s", normalized)
		}

		for _, fragment := range []string{
			"from game_periods",
			"where status = $1",
			"draw_at <= $2",
			"order by draw_at asc, id asc",
			"limit $3",
		} {
			if !strings.Contains(normalized, fragment) {
				return fmt.Errorf("draw queue query is missing %q: %s", fragment, normalized)
			}
		}
		return nil
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(queryMatcher))
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	drawAt := time.Date(2026, time.August, 6, 15, 3, 0, 0, time.UTC)
	openAt := drawAt.Add(-3 * time.Minute)
	lockAt := drawAt.Add(-30 * time.Second)
	columns := []string{
		"id", "room_code", "game_type", "period_no", "period_index",
		"open_at", "bet_lock_at", "draw_at", "status", "manual_result",
	}
	rows := sqlmock.NewRows(columns).
		AddRow(int64(37090305), "wingo_3m", 1, "WINGO_3M_1786003380", int64(1), openAt, lockAt, drawAt, periodStatusLocked, nil).
		AddRow(int64(37096785), "wingo_3m", 1, "WINGO_3M_1786003560", int64(2), openAt.Add(3*time.Minute), lockAt.Add(3*time.Minute), drawAt.Add(3*time.Minute), periodStatusLocked, nil)

	mock.ExpectQuery("draw queue").
		WithArgs(periodStatusLocked, drawAt.Add(10*time.Minute), 100).
		WillReturnRows(rows)

	repository := NewGameRepository(db)
	periods, err := repository.ListLockedPeriodsForDraw(t.Context(), drawAt.Add(10*time.Minute), 100)
	if err != nil {
		t.Fatalf("list periods: %v", err)
	}
	if len(periods) != 2 {
		t.Fatalf("expected every locked period from the room, got %d", len(periods))
	}
	if periods[0].RoomCode != "wingo_3m" || periods[1].RoomCode != "wingo_3m" {
		t.Fatalf("expected both periods to remain in the same room queue: %#v", periods)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestSettlementWalletBalanceProductionOverflowRegression(t *testing.T) {
	firstPayout := testWinningSmallPayout(t, "50000000000", "1000000000")
	if firstPayout != "99000000000.00000000" {
		t.Fatalf("unexpected first ticket payout: %s", firstPayout)
	}
	secondPayout := testWinningSmallPayout(t, "500000000000", "10000000000")
	if secondPayout != "990000000000.00000000" {
		t.Fatalf("unexpected second ticket payout: %s", secondPayout)
	}

	balanceAfterFirst, err := addNumeric("2334980.00000000", firstPayout)
	if err != nil {
		t.Fatalf("add first payout: %v", err)
	}
	if balanceAfterFirst != "99002334980.00000000" {
		t.Fatalf("unexpected balance after first ticket: %s", balanceAfterFirst)
	}

	balanceAfterSecond, err := addNumeric(balanceAfterFirst, secondPayout)
	if err != nil {
		t.Fatalf("add second payout: %v", err)
	}
	if balanceAfterSecond != "1089002334980.00000000" {
		t.Fatalf("unexpected balance after second ticket: %s", balanceAfterSecond)
	}
	if testValueFitsNumericPrecisionScale(balanceAfterSecond, 20, 8) {
		t.Fatalf("production balance must exceed the legacy numeric(20,8) capacity: %s", balanceAfterSecond)
	}
	if !testValueFitsNumericPrecisionScale(balanceAfterSecond, 30, 8) {
		t.Fatalf("production balance must fit the widened money capacity: %s", balanceAfterSecond)
	}

	balanceBeyondFixedPrecision, err := addNumeric("9999999999999999999999.99999999", "0.00000001")
	if err != nil {
		t.Fatalf("add cumulative balance beyond fixed precision: %v", err)
	}
	if balanceBeyondFixedPrecision != "10000000000000000000000.00000000" {
		t.Fatalf("unexpected unconstrained cumulative balance: %s", balanceBeyondFixedPrecision)
	}
	if testValueFitsNumericPrecisionScale(balanceBeyondFixedPrecision, 30, 8) {
		t.Fatalf("cumulative balance must exercise unconstrained NUMERIC storage: %s", balanceBeyondFixedPrecision)
	}
}

func TestSettlePeriodClearsRetryMetadataOnSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	periodID := int64(37090305)
	roomCode := "wingo_3m"
	periodNo := "WINGO_3M_1786003380"
	mock.ExpectBegin()
	mock.ExpectQuery(`select status\s+from game_periods\s+where id = \$1\s+for update`).
		WithArgs(periodID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(periodStatusDrawn))
	mock.ExpectQuery(`from bet_tickets\s+where period_id = \$1 and status = \$2`).
		WithArgs(periodID, betStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "wallet_id", "original_amount", "tax_amount", "net_amount", "total_stake", "items",
		}))
	mock.ExpectExec(`update game_periods\s+set status = \$1,\s+settled_at = \$2,\s+settlement_attempts = 0,\s+settlement_last_error = null,\s+settlement_next_retry_at = null,\s+updated_at = \$2\s+where id = \$3 and status = \$4`).
		WithArgs(periodStatusSettled, sqlmock.AnyArg(), periodID, periodStatusDrawn).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update game_round_histories\s+set status = 'SETTLED'`).
		WithArgs(sqlmock.AnyArg(), roomCode, periodNo).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repository := NewGameRepository(db)
	userIDs, err := repository.SettlePeriod(t.Context(), GamePeriodSettlementRecord{
		ID:            periodID,
		RoomCode:      roomCode,
		PeriodNo:      periodNo,
		ResultPayload: []byte(`{"tags":["small"]}`),
	})
	if err != nil {
		t.Fatalf("settle period: %v", err)
	}
	if len(userIDs) != 0 {
		t.Fatalf("period without tickets must not publish user updates: %#v", userIDs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestSettlePeriodSkipsPeriodThatAnotherWorkerAlreadySettled(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectBegin()
	mock.ExpectQuery(`select status\s+from game_periods\s+where id = \$1\s+for update`).
		WithArgs(int64(37090305)).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(periodStatusSettled))
	mock.ExpectRollback()

	repository := NewGameRepository(db)
	userIDs, err := repository.SettlePeriod(t.Context(), GamePeriodSettlementRecord{ID: 37090305})
	if err != nil {
		t.Fatalf("settle period: %v", err)
	}
	if len(userIDs) != 0 {
		t.Fatalf("already settled period must not publish user updates: %#v", userIDs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestSettlePeriodRejectsNegativeLockedBalanceInvariant(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	periodID := int64(37090305)
	walletID := int64(1163)
	items := []byte(`[{"stake":"1000","option_key":"small","option_type":"BIG_SMALL"}]`)
	mock.ExpectBegin()
	mock.ExpectQuery(`select status\s+from game_periods\s+where id = \$1\s+for update`).
		WithArgs(periodID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(periodStatusDrawn))
	mock.ExpectQuery(`from bet_tickets\s+where period_id = \$1 and status = \$2`).
		WithArgs(periodID, betStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "user_id", "wallet_id", "original_amount", "tax_amount", "net_amount", "total_stake", "items",
		}).AddRow(int64(4132), int64(642258), walletID, "1000.00000000", "20.00000000", "980.00000000", "1000.00000000", items))
	mock.ExpectQuery(`select balance::text, locked_balance::text\s+from wallets\s+where id = \$1\s+for update`).
		WithArgs(walletID).
		WillReturnRows(sqlmock.NewRows([]string{"balance", "locked_balance"}).AddRow("5000.00000000", "500.00000000"))
	mock.ExpectRollback()

	repository := NewGameRepository(db)
	_, err = repository.SettlePeriod(t.Context(), GamePeriodSettlementRecord{
		ID:            periodID,
		RoomCode:      "wingo_3m",
		PeriodNo:      "WINGO_3M_1786003380",
		GameType:      1,
		ResultPayload: []byte(`{"tags":["small"]}`),
	})
	if err == nil || !strings.Contains(err.Error(), "locked balance invariant violated") {
		t.Fatalf("SettlePeriod error = %v, want locked balance invariant violation", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func TestNormalizeMoneyForStorageEnforcesNumeric30Scale8(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "normalizes integer", input: "1000", want: "1000.00000000"},
		{name: "keeps exact scale", input: "1000.12345678", want: "1000.12345678"},
		{name: "accepts signed decimal", input: " +1000.25 ", want: "1000.25000000"},
		{name: "accepts negative database decimal", input: "-1000.25000000", want: "-1000.25000000"},
		{name: "accepts largest magnitude", input: "9999999999999999999999.99999999", want: "9999999999999999999999.99999999"},
		{name: "rejects fractional precision loss", input: "1000.123456789", wantErr: true},
		{name: "rejects positive capacity boundary", input: "10000000000000000000000", wantErr: true},
		{name: "rejects negative capacity boundary", input: "-10000000000000000000000", wantErr: true},
		{name: "rejects rational fraction", input: "1000/1", wantErr: true},
		{name: "rejects scientific notation", input: "1e3", wantErr: true},
		{name: "rejects empty input", input: "", wantErr: true},
		{name: "rejects decimal point only", input: ".", wantErr: true},
		{name: "rejects missing integer digits", input: ".5", wantErr: true},
		{name: "rejects missing fractional digits", input: "1.", wantErr: true},
		{name: "rejects NaN", input: "NaN", wantErr: true},
		{name: "rejects infinity", input: "Inf", wantErr: true},
		{name: "rejects invalid numeric", input: "not-a-number", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeMoneyForStorage(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeMoneyForStorage(%q) unexpectedly succeeded with %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeMoneyForStorage(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeMoneyForStorage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizePendingPayoutExposureUsesOnlyReservedPayouts(t *testing.T) {
	exposure, err := normalizePendingPayoutExposure("2000.00000000", "1980.00000000")
	if err != nil {
		t.Fatalf("normalize pending payout exposure: %v", err)
	}
	if exposure != "3980.00000000" {
		t.Fatalf("pending payout exposure = %q, want 3980.00000000", exposure)
	}

	_, err = normalizePendingPayoutExposure("9999999999999999999000.00000000", "1980.00000000")
	if err == nil {
		t.Fatal("pending payout exposure beyond numeric(30,8) unexpectedly succeeded")
	}
}

func TestNormalizeBetStakeEnforcesMinimumAndPositiveAmount(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "minimum", input: "1000", want: "1000.00000000"},
		{name: "above minimum", input: "1000.00000001", want: "1000.00000001"},
		{name: "below minimum", input: "999.99999999", wantErr: true},
		{name: "zero", input: "0", wantErr: true},
		{name: "negative", input: "-1000", wantErr: true},
		{name: "too many decimals", input: "1000.000000001", wantErr: true},
		{name: "storage overflow", input: "10000000000000000000000", wantErr: true},
		{name: "rational fraction", input: "1000/1", wantErr: true},
		{name: "scientific notation", input: "1e3", wantErr: true},
		{name: "empty input", input: "", wantErr: true},
		{name: "decimal point only", input: ".", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeBetStake(tt.input)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidBetAmount) {
					t.Fatalf("normalizeBetStake(%q) error = %v, want ErrInvalidBetAmount", tt.input, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeBetStake(%q): %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeBetStake(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeBetTicketItemsNormalizesAndChecksAggregateCapacity(t *testing.T) {
	items, total, err := normalizeBetTicketItems([]BetTicketItemRecord{
		{OptionType: " BIG_SMALL ", OptionKey: " small ", Stake: "1000"},
		{OptionType: "NUMBER", OptionKey: "number_2", Stake: "2000.12345678"},
	})
	if err != nil {
		t.Fatalf("normalize items: %v", err)
	}
	if total != "3000.12345678" {
		t.Fatalf("normalized total = %q", total)
	}
	if len(items) != 2 || items[0].Stake != "1000.00000000" || items[0].OptionType != "BIG_SMALL" || items[0].OptionKey != "small" {
		t.Fatalf("unexpected normalized items: %#v", items)
	}

	_, _, err = normalizeBetTicketItems([]BetTicketItemRecord{
		{OptionType: "BIG_SMALL", OptionKey: "small", Stake: "6000000000000000000000"},
		{OptionType: "BIG_SMALL", OptionKey: "big", Stake: "6000000000000000000000"},
	})
	if !errors.Is(err, ErrInvalidBetAmount) {
		t.Fatalf("aggregate overflow error = %v, want ErrInvalidBetAmount", err)
	}
}

func TestCreateBetTicketRejectsMismatchedItemTotalBeforeOpeningTransaction(t *testing.T) {
	repository := NewGameRepository(nil)
	_, err := repository.CreateBetTicket(t.Context(), CreateBetTicketParams{
		RoomCode:   "wingo_3m",
		TotalStake: "1001",
		Items: []BetTicketItemRecord{{
			OptionType: "BIG_SMALL",
			OptionKey:  "small",
			Stake:      "1000",
		}},
	})
	if !errors.Is(err, ErrInvalidBetAmount) {
		t.Fatalf("CreateBetTicket error = %v, want ErrInvalidBetAmount", err)
	}
}

func TestCreateBetTicketRejectsProjectedPendingPayoutExposure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	now := time.Now()
	periodID := int64(37090305)
	walletID := int64(1163)
	mock.ExpectBegin()
	mock.ExpectQuery(`select id, room_code, game_type, period_no, period_index, open_at, bet_lock_at, draw_at, status, manual_result\s+from game_periods`).
		WithArgs(periodID, "wingo_3m").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "room_code", "game_type", "period_no", "period_index",
			"open_at", "bet_lock_at", "draw_at", "status", "manual_result",
		}).AddRow(
			periodID, "wingo_3m", 1, "WINGO_3M_1786003380", int64(1),
			now.Add(-time.Minute), now.Add(time.Minute), now.Add(2*time.Minute), periodStatusOpen, nil,
		))
	mock.ExpectQuery(`select id, balance::text, locked_balance::text, status\s+from wallets`).
		WithArgs(int64(642258), 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "balance", "locked_balance", "status"}).
			AddRow(walletID, "1000000000000000000000000.00000000", "0.00000000", 1))
	mock.ExpectQuery(`select coalesce\(sum\(coalesce\(potential_payout, 0\)\), 0\)::text\s+from bet_tickets`).
		WithArgs(walletID, betStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{"pending_payout"}).AddRow("9999999999999999999000.00000000"))
	mock.ExpectRollback()

	repository := NewGameRepository(db)
	_, err = repository.CreateBetTicket(t.Context(), CreateBetTicketParams{
		UserID:     642258,
		RoomCode:   "wingo_3m",
		PeriodID:   periodID,
		TotalStake: "1000",
		Items: []BetTicketItemRecord{{
			OptionType: "BIG_SMALL",
			OptionKey:  "small",
			Stake:      "1000",
		}},
	})
	if !errors.Is(err, ErrInvalidBetAmount) {
		t.Fatalf("CreateBetTicket error = %v, want ErrInvalidBetAmount", err)
	}
	if !strings.Contains(err.Error(), "payout exposure") {
		t.Fatalf("CreateBetTicket error = %v, want payout exposure context", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}

func testWinningSmallPayout(t *testing.T, stake, tax string) string {
	t.Helper()

	items, err := json.Marshal([]BetTicketItemRecord{{
		OptionType: "BIG_SMALL",
		OptionKey:  "small",
		Stake:      stake,
	}})
	if err != nil {
		t.Fatalf("marshal items: %v", err)
	}

	outcomes, payoutTotal, err := settleTicketItems(
		items,
		map[string]struct{}{"small": {}},
		stake,
		tax,
	)
	if err != nil {
		t.Fatalf("settle ticket: %v", err)
	}
	if len(outcomes) != 1 || !outcomes[0].IsWin {
		t.Fatalf("unexpected item outcome: %#v", outcomes)
	}
	return payoutTotal
}

func testValueFitsNumericPrecisionScale(value string, precision, scale int) bool {
	parsed, ok := new(big.Rat).SetString(value)
	if !ok || precision <= 0 || scale < 0 || scale > precision {
		return false
	}

	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision-scale)), nil)
	abs := new(big.Rat).Abs(parsed)
	return abs.Cmp(new(big.Rat).SetInt(limit)) < 0
}
