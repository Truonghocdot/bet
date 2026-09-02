package postgres

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gin/internal/domain/telegram"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpsertTelegramGroupEventUsesExplicitPostgresParameterTypes(t *testing.T) {
	matcher := sqlmock.QueryMatcherFunc(func(_ string, actualSQL string) error {
		normalized := strings.ToLower(strings.Join(strings.Fields(actualSQL), " "))
		for _, fragment := range []string{
			"$1::varchar(32)",
			"$2::bigint",
			"$3::varchar(20)",
			"$6::varchar(32)",
			"case when $6::text in ('left', 'kicked') then $7::timestamp else null::timestamp end",
		} {
			if !strings.Contains(normalized, fragment) {
				return fmt.Errorf("telegram upsert is missing explicit parameter type %q: %s", fragment, normalized)
			}
		}
		return nil
	})
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(matcher))
	if err != nil {
		t.Fatalf("create SQL mock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	when := time.Date(2026, time.September, 2, 6, 49, 46, 0, time.UTC)
	mock.ExpectExec("telegram group upsert").
		WithArgs("fh88u", int64(-1001234567890), "supergroup", "Group test", "", "member", when).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repository := NewTelegramRepository(db)
	err = repository.UpsertGroupEvent(t.Context(), telegram.GroupEvent{
		SiteCode:   "fh88u",
		ChatID:     -1001234567890,
		ChatType:   "supergroup",
		Title:      "Group test",
		BotStatus:  "member",
		OccurredAt: when,
	})
	if err != nil {
		t.Fatalf("upsert telegram group event: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("SQL expectations: %v", err)
	}
}
