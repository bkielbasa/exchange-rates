package fetcher

import (
	"context"

	"github.com/bklimczak/exchange-rates/internal/rates"
)

type Fetcher interface {
	Fetch(ctx context.Context, currency string) (rates.Rate, error)
}

type Saver interface {
	Save(ctx context.Context, r rates.Rate) error
}

var DefaultCurrencies = []string{
	"USD", "GBP", "JPY", "CHF", "CAD",
	"AUD", "SEK", "NOK", "PLN", "DKK",
}
