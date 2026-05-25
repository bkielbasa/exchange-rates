package serve

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/bklimczak/exchange-rates/internal/rates"
)

// Server is the HTTP microservice. Construct with New, then Start to bind the
// listener and serve in a background goroutine, Shutdown to stop gracefully.
type Server struct {
	cfg    Config
	logger *slog.Logger
	http   *http.Server
	ln     net.Listener
	errCh  chan error
}

func New(cfg Config, logger *slog.Logger, db *sql.DB) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	mux := http.NewServeMux()
	rates.New(db, rates.WithLogger(logger)).Register(mux)

	handler := otelhttp.NewHandler(mux, "exchange-rates")

	httpSrv := &http.Server{
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		http:   httpSrv,
		errCh:  make(chan error, 1),
	}
}

// Start binds the listener (honouring ":0" so callers can ask for a free port)
// and runs Serve in a goroutine. Errors from Serve are surfaced via Shutdown.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.cfg.HTTPAddr)
	if err != nil {
		s.logger.Error("serve: listen failed", "addr", s.cfg.HTTPAddr, "error", err)
		return fmt.Errorf("listen %s: %w", s.cfg.HTTPAddr, err)
	}
	s.ln = ln
	s.logger.Info("serve: listening", "addr", ln.Addr().String())

	go func() {
		if err := s.http.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("serve: http serve exited with error", "error", err)
			s.errCh <- err
		}
		close(s.errCh)
	}()
	return nil
}

// Addr returns the bound address. Only valid after Start.
func (s *Server) Addr() string {
	if s.ln == nil {
		return ""
	}
	return s.ln.Addr().String()
}

// Shutdown gracefully stops the server. Returns the first error from either
// Shutdown itself or from Serve (if it died unexpectedly).
func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.http.Shutdown(ctx); err != nil {
		s.logger.Error("serve: http shutdown failed", "error", err)
		return fmt.Errorf("http shutdown: %w", err)
	}
	if err, ok := <-s.errCh; ok && err != nil {
		s.logger.Error("serve: shutdown surfaced serve error", "error", err)
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
