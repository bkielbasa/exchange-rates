package rates

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// In real production code I'd create a separate `Rate` type only for displaying it on the frontend.
// Very often the type we have in our domain is different comapred to what we display to the user,
// but for simplicity I'm living it as it is right now.

type svcReader interface {
	Latest(ctx context.Context) ([]Rate, error)
	History(ctx context.Context, currency string) ([]RateHistoryEntry, error)
}

type handler struct {
	svc    svcReader
	logger *slog.Logger
	tracer trace.Tracer
}

func newHandler(svc svcReader) *handler {
	return &handler{
		svc:    svc,
		logger: slog.Default(),
		tracer: otel.Tracer("rates.handler"),
	}
}

func (h *handler) getLatest(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "rates.getLatest",
		trace.WithAttributes(attribute.String("http.route", "/api/v1/rates/latest")),
	)
	defer span.End()

	out, err := h.svc.Latest(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "get latest failed")
		h.logger.ErrorContext(ctx, "get latest failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	span.SetAttributes(attribute.Int("rates.count", len(out)))
	writeJSON(w, http.StatusOK, out)
}

func (h *handler) getHistory(w http.ResponseWriter, r *http.Request) {
	currency := r.PathValue("currency")
	ctx, span := h.tracer.Start(r.Context(), "rates.getHistory",
		trace.WithAttributes(
			attribute.String("http.route", "/api/v1/rates/history/{currency}"),
			attribute.String("currency", currency),
		),
	)
	defer span.End()

	out, err := h.svc.History(ctx, currency)
	switch {
	case errors.Is(err, errInvalidCurrency):
		span.SetAttributes(attribute.String("rates.result", "invalid_currency"))
		h.logger.WarnContext(ctx, "get history rejected: invalid currency", "currency", currency)
		writeJSONError(w, http.StatusBadRequest, "invalid currency code")
		return
	case err != nil:
		span.RecordError(err)
		span.SetStatus(codes.Error, "get history failed")
		h.logger.ErrorContext(ctx, "get history failed", "currency", currency, "error", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if len(out) == 0 {
		span.SetAttributes(attribute.String("rates.result", "not_found"))
		h.logger.InfoContext(ctx, "get history: no rows for currency", "currency", currency)
		writeJSONError(w, http.StatusNotFound, "no history for currency")
		return
	}

	span.SetAttributes(attribute.Int("rates.count", len(out)))
	writeJSON(w, http.StatusOK, out)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// in prod applications I try to follow https://www.rfc-editor.org/rfc/rfc9457.html but for sake of simplicity
// I skipped it here.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
