package serve_test

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/bklimczak/exchange-rates/internal/cli/serve"
)

func TestServer_StartShutdown_BindsAndStopsCleanly(t *testing.T) {
	// We don't need a real DB for this smoke test — rates.New tolerates nil.
	var db *sql.DB
	cfg := serve.Config{HTTPAddr: ":0", MySQLDSN: "ignored"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	srv := serve.New(cfg, logger, db)
	if err := srv.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if srv.Addr() == "" {
		t.Fatalf("Addr returned empty after Start")
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://" + srv.Addr() + "/no/such/path")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
