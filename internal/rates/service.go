package rates

import (
	"context"
	"errors"
	"strings"
)

// errInvalidCurrency is returned for malformed currency codes.
var errInvalidCurrency = errors.New("invalid currency code")

// reader is the read interface the service needs. Defined locally so the
// service is independent of the concrete repository implementation.
type reader interface {
	Latest(ctx context.Context) ([]Rate, error)
	History(ctx context.Context, currency string) ([]RateHistoryEntry, error)
}

// service holds the business logic over a reader.
type service struct {
	repo reader
}

func newService(repo reader) *service {
	return &service{repo: repo}
}

func (s *service) Latest(ctx context.Context) ([]Rate, error) {
	return s.repo.Latest(ctx)
}

func (s *service) History(ctx context.Context, currency string) ([]RateHistoryEntry, error) {
	c := strings.ToUpper(strings.TrimSpace(currency))
	if !isValidCurrencyCode(c) {
		return nil, errInvalidCurrency
	}
	return s.repo.History(ctx, c)
}

func isValidCurrencyCode(c string) bool {
	if len(c) != 3 {
		return false
	}
	for _, r := range c {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
