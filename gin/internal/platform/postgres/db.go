package postgres

import (
	"database/sql"
	"fmt"
	"net/url"

	"gin/internal/support/message"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf(message.DatabaseURLRequired)
	}

	databaseURL = applyVietnamTimezone(databaseURL)

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func applyVietnamTimezone(databaseURL string) string {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return databaseURL
	}

	query := parsed.Query()
	if query.Get("options") == "" {
		query.Set("options", "-c timezone=Asia/Ho_Chi_Minh")
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}
