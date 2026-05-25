//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/bklimczak/exchange-rates/internal/rates"
)

func TestE2E_GetLatest_ReturnsLatestRowPerCurrency(t *testing.T) {
	truncateRates(t, testDB)

	// given
	older := time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)
	seedRates(t, testDB, []rateRow{
		{Currency: "USD", Rate: 1.07, FetchedAt: older},
		{Currency: "USD", Rate: 1.0823, FetchedAt: newer},
		{Currency: "GBP", Rate: 0.85, FetchedAt: older},
		{Currency: "GBP", Rate: 0.8541, FetchedAt: newer},
	})

	// when
	resp, err := http.Get(baseURL + "/api/v1/rates/latest")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	// then
	var got []rates.Rate
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d rows, want 2: %+v", len(got), got)
	}

	// a big ugly but for homework should work :) 
	byCur := map[string]rates.Rate{got[0].Currency: got[0], got[1].Currency: got[1]}
	if byCur["USD"].Rate != 1.0823 {
		t.Errorf("USD rate = %v, want 1.0823", byCur["USD"].Rate)
	}

	if byCur["GBP"].Rate != 0.8541 {
		t.Errorf("GBP rate = %v, want 0.8541", byCur["GBP"].Rate)
	}

	if !byCur["USD"].FetchedAt.Equal(newer) {
		t.Errorf("USD fetched_at = %v, want %v", byCur["USD"].FetchedAt, newer)
	}
}
