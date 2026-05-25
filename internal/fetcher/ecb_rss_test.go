package fetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// sampleFeed mirrors the real bank.lv ECB RSS shape: one item per day, all
// currencies in a single space-separated CDATA description, oldest first.
const sampleFeed = `<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
  <channel>
    <title>ECB rates</title>
    <item>
      <title>ECB EXCHANGE RATES.</title>
      <description><![CDATA[USD 1.0823 GBP 0.8541 PLN 4.2500]]></description>
      <pubDate>Tue, 20 May 2026 03:00:00 +0300</pubDate>
    </item>
    <item>
      <title>ECB EXCHANGE RATES.</title>
      <description><![CDATA[USD 1.1500 GBP 0.8700 PLN 4.2400]]></description>
      <pubDate>Wed, 21 May 2026 03:00:00 +0300</pubDate>
    </item>
  </channel>
</rss>`

func newSampleServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(sampleFeed))
	}))
}

func TestECBRSSFetcher_Fetch_PicksLatestRateForCurrency(t *testing.T) {
	srv := newSampleServer(t)
	defer srv.Close()

	f := NewECBRSSFetcher(srv.URL, nil)
	got, err := f.Fetch(context.Background(), "USD")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got.Currency != "USD" || got.Rate != 1.1500 {
		t.Fatalf("got %+v, want USD 1.1500", got)
	}
}

func TestECBRSSFetcher_Fetch_ExtractsRequestedCurrencyFromMultiCurrencyDescription(t *testing.T) {
	srv := newSampleServer(t)
	defer srv.Close()

	f := NewECBRSSFetcher(srv.URL, nil)
	got, err := f.Fetch(context.Background(), "GBP")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got.Rate != 0.8700 {
		t.Fatalf("rate = %v, want 0.8700", got.Rate)
	}
}

func TestECBRSSFetcher_Fetch_UnknownCurrency_Returns404Sentinel(t *testing.T) {
	srv := newSampleServer(t)
	defer srv.Close()

	f := NewECBRSSFetcher(srv.URL, nil)
	_, err := f.Fetch(context.Background(), "XYZ")
	if err == nil || !strings.Contains(err.Error(), "currency XYZ not found") {
		t.Fatalf("err = %v", err)
	}
}
