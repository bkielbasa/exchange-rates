package rates

import "time"

// Rate is the latest known exchange rate for a currency.
type Rate struct {
	Currency  string    `json:"currency"`
	Rate      float64   `json:"rate"`
	FetchedAt time.Time `json:"fetched_at"`
}

// RateHistoryEntry is a single historical data point for a currency.
type RateHistoryEntry struct {
	Rate      float64   `json:"rate"`
	FetchedAt time.Time `json:"fetched_at"`
}
