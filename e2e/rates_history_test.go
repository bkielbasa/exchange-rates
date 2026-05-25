//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/bklimczak/exchange-rates/internal/rates"
)

func TestE2E_GetHistory_ReturnsDescendingHistoryForCurrency(t *testing.T) {
	truncateRates(t, testDB)

	t1 := time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)
	seedRates(t, testDB, []rateRow{
		{Currency: "USD", Rate: 1.06, FetchedAt: t1},
		{Currency: "USD", Rate: 1.07, FetchedAt: t2},
		{Currency: "USD", Rate: 1.0823, FetchedAt: t3},
		{Currency: "GBP", Rate: 0.85, FetchedAt: t2},
	})

	resp, err := http.Get(baseURL + "/api/v1/rates/history/USD")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var got []rates.RateHistoryEntry
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3: %+v", len(got), got)
	}
	if !got[0].FetchedAt.Equal(t3) || !got[1].FetchedAt.Equal(t2) || !got[2].FetchedAt.Equal(t1) {
		t.Errorf("rows not in DESC fetched_at order: %+v", got)
	}
	if got[0].Rate != 1.0823 || got[1].Rate != 1.07 || got[2].Rate != 1.06 {
		t.Errorf("rates wrong: %+v", got)
	}
}

func TestE2E_GetHistory_UnknownCurrency_Returns404(t *testing.T) {
	truncateRates(t, testDB)

	// given (empty table)

	// when
	resp, err := http.Get(baseURL + "/api/v1/rates/history/USD")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()

	// then
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

func TestE2E_GetHistory_InvalidCurrency_Returns400(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/rates/history/us")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}
