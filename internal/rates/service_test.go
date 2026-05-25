package rates

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRepo struct {
	latest          []Rate
	latestErr       error
	history         []RateHistoryEntry
	historyErr      error
	historyCurrency string
}

func (f *fakeRepo) Latest(ctx context.Context) ([]Rate, error) {
	return f.latest, f.latestErr
}

func (f *fakeRepo) History(ctx context.Context, currency string) ([]RateHistoryEntry, error) {
	f.historyCurrency = currency
	return f.history, f.historyErr
}

func (f *fakeRepo) Save(ctx context.Context, r Rate) error { return nil }

func TestService_Latest_PassesThroughResults(t *testing.T) {
	want := []Rate{{Currency: "USD", Rate: 1.08, FetchedAt: time.Unix(1, 0)}}
	s := newService(&fakeRepo{latest: want})

	got, err := s.Latest(context.Background())
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if len(got) != 1 || got[0].Currency != "USD" {
		t.Fatalf("got %+v want %+v", got, want)
	}
}

func TestService_History_UppercasesCurrencyAndForwards(t *testing.T) {
	repo := &fakeRepo{history: []RateHistoryEntry{{Rate: 1.0, FetchedAt: time.Unix(1, 0)}}}
	s := newService(repo)

	_, err := s.History(context.Background(), "usd")
	if err != nil {
		t.Fatalf("History: %v", err)
	}
	if repo.historyCurrency != "USD" {
		t.Fatalf("history currency = %q, want %q", repo.historyCurrency, "USD")
	}
}

func TestService_History_RejectsInvalidCurrency(t *testing.T) {
	s := newService(&fakeRepo{})

	_, err := s.History(context.Background(), "us")
	if !errors.Is(err, errInvalidCurrency) {
		t.Fatalf("got %v, want errInvalidCurrency", err)
	}
}
