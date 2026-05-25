package fetch

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bklimczak/exchange-rates/internal/db"
	"github.com/bklimczak/exchange-rates/internal/fetcher"
	"github.com/bklimczak/exchange-rates/internal/rates"
)

func Run(ctx context.Context, cfg Config, logger *slog.Logger) error {
	logger.InfoContext(ctx, "fetch: starting", "source_url", cfg.FetchSourceURL, "currencies", len(fetcher.DefaultCurrencies))

	pool, err := db.Open(cfg.MySQLDSN)
	if err != nil {
		logger.ErrorContext(ctx, "fetch: open database failed", "error", err)
		return err
	}
	defer pool.Close()

	r := rates.New(pool, rates.WithLogger(logger))
	f := fetcher.NewECBRSSFetcher(cfg.FetchSourceURL, logger)

	if err := fetcher.FetchAll(ctx, f, r, fetcher.DefaultCurrencies, logger); err != nil {
		logger.ErrorContext(ctx, "fetch: fetch all failed", "error", err)
		return fmt.Errorf("fetch all: %w", err)
	}
	logger.InfoContext(ctx, "fetch: complete", "currencies", len(fetcher.DefaultCurrencies))
	return nil
}
