package rates

import (
	"context"
	"database/sql"
	"fmt"
)

// repository is the persistence layer for rates. Unexported on purpose:
// the package's public API is Rates.Register and Rates.Save.
type repository struct {
	db *sql.DB
}

func newRepository(db *sql.DB) *repository {
	return &repository{db: db}
}

const latestQuery = `
SELECT r.currency, r.rate, r.fetched_at
FROM rates r
INNER JOIN (
    SELECT currency, MAX(fetched_at) AS max_fetched_at
    FROM rates
    GROUP BY currency
) latest
ON latest.currency = r.currency
AND latest.max_fetched_at = r.fetched_at
ORDER BY r.currency
`

func (r *repository) Latest(ctx context.Context) ([]Rate, error) {
	rows, err := r.db.QueryContext(ctx, latestQuery)
	if err != nil {
		return nil, fmt.Errorf("query latest: %w", err)
	}
	defer rows.Close()

	out := []Rate{}
	for rows.Next() {
		var rate Rate
		if err := rows.Scan(&rate.Currency, &rate.Rate, &rate.FetchedAt); err != nil {
			return nil, fmt.Errorf("scan latest: %w", err)
		}
		out = append(out, rate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate latest: %w", err)
	}
	return out, nil
}

const historyQuery = `
SELECT rate, fetched_at
FROM rates
WHERE currency = ?
ORDER BY fetched_at DESC
LIMIT 100
`

func (r *repository) History(ctx context.Context, currency string) ([]RateHistoryEntry, error) {
	rows, err := r.db.QueryContext(ctx, historyQuery, currency)
	if err != nil {
		return nil, fmt.Errorf("query history: %w", err)
	}
	defer rows.Close()

	out := []RateHistoryEntry{}
	for rows.Next() {
		var entry RateHistoryEntry
		if err := rows.Scan(&entry.Rate, &entry.FetchedAt); err != nil {
			return nil, fmt.Errorf("scan history: %w", err)
		}
		out = append(out, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate history: %w", err)
	}
	return out, nil
}

const saveQuery = `INSERT INTO rates (currency, rate, fetched_at) VALUES (?, ?, ?)`

func (r *repository) Save(ctx context.Context, rate Rate) error {
	if _, err := r.db.ExecContext(ctx, saveQuery, rate.Currency, rate.Rate, rate.FetchedAt); err != nil {
		return fmt.Errorf("insert rate: %w", err)
	}
	return nil
}
