package observability_test

import (
	"context"
	"testing"

	"github.com/bklimczak/exchange-rates/internal/observability"
)

func TestInit_DisabledReturnsNoopLoggerAndShutdown(t *testing.T) {
	t.Setenv("OTEL_SDK_DISABLED", "true")

	logger, shutdown, err := observability.Init(context.Background(), "test")
	if err != nil {
		t.Fatalf("Init returned error: %v", err)
	}
	if logger == nil {
		t.Fatalf("Init returned nil logger")
	}
	if shutdown == nil {
		t.Fatalf("Init returned nil shutdown")
	}

	logger.Info("hello", "k", "v")

	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}
}
