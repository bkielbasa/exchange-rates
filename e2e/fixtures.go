//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

func truncateRates(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), "TRUNCATE TABLE rates"); err != nil {
		t.Fatalf("truncate rates: %v", err)
	}
}

type rateRow struct {
	Currency  string
	Rate      float64
	FetchedAt time.Time
}

func seedRates(t *testing.T, db *sql.DB, rows []rateRow) {
	t.Helper()
	stmt, err := db.PrepareContext(context.Background(),
		`INSERT INTO rates (currency, rate, fetched_at) VALUES (?, ?, ?)`)
	if err != nil {
		t.Fatalf("prepare: %v", err)
	}
	defer stmt.Close()
	for _, r := range rows {
		if _, err := stmt.ExecContext(context.Background(), r.Currency, r.Rate, r.FetchedAt.UTC()); err != nil {
			t.Fatalf("insert (%s): %v", r.Currency, err)
		}
	}
}
