package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS daily_kline_tab (
    code       TEXT    NOT NULL,
    trade_date TEXT    NOT NULL,
    open       REAL,
    high       REAL,
    low        REAL,
    close      REAL,
    volume     REAL,
    amount     REAL,
    PRIMARY KEY (code, trade_date)
);
`

func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

type Stats struct {
	SymbolCount int
	TotalRows   int
	MinDate     string
	MaxDate     string
}

func GetStats(db *sql.DB) (*Stats, error) {
	s := &Stats{}

	row := db.QueryRow("SELECT COUNT(DISTINCT code) FROM daily_kline_tab")
	if err := row.Scan(&s.SymbolCount); err != nil {
		return nil, fmt.Errorf("count symbols: %w", err)
	}

	row = db.QueryRow("SELECT COUNT(*) FROM daily_kline_tab")
	if err := row.Scan(&s.TotalRows); err != nil {
		return nil, fmt.Errorf("count rows: %w", err)
	}

	row = db.QueryRow("SELECT MIN(trade_date), MAX(trade_date) FROM daily_kline_tab")
	var minDate, maxDate sql.NullString
	if err := row.Scan(&minDate, &maxDate); err != nil {
		return nil, fmt.Errorf("date range: %w", err)
	}
	if minDate.Valid {
		s.MinDate = minDate.String
	}
	if maxDate.Valid {
		s.MaxDate = maxDate.String
	}

	return s, nil
}
