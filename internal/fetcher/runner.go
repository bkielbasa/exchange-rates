package fetcher

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

func FetchAll(ctx context.Context, f Fetcher, saver Saver, currencies []string, logger *slog.Logger) error {
	if logger == nil {
		logger = slog.Default()
	}

	tracer := otel.Tracer("fetcher")
	meter := otel.Meter("fetcher")
	resultCounter, err := meter.Int64Counter(
		"fetcher.fetch.result",
		metric.WithDescription("Per-currency fetch outcomes"),
	)
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)
	for _, c := range currencies {
		currency := c
		g.Go(func() error {
			gctx, span := tracer.Start(gctx, "fetcher.fetch_one",
				trace.WithAttributes(attribute.String("currency", currency)),
			)
			defer span.End()

			rate, err := f.Fetch(gctx, currency)
			if err != nil {
				logger.ErrorContext(gctx, "fetch failed", "currency", currency, "error", err)
				resultCounter.Add(gctx, 1,
					metric.WithAttributes(
						attribute.String("currency", currency),
						attribute.String("outcome", "error"),
					))
				return err
			}

			if err := saver.Save(gctx, rate); err != nil {
				logger.ErrorContext(gctx, "save failed", "currency", currency, "error", err)
				resultCounter.Add(gctx, 1,
					metric.WithAttributes(
						attribute.String("currency", currency),
						attribute.String("outcome", "error"),
					))
				return err
			}

			resultCounter.Add(gctx, 1,
				metric.WithAttributes(
					attribute.String("currency", currency),
					attribute.String("outcome", "success"),
				))
			return nil
		})
	}
	return g.Wait()
}
