package rates

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
)

type Rates struct {
	repo    *repository
	svc     *service
	h       *handler
	logger  *slog.Logger
}

type Option func (r *Rates)

func WithLogger(logger *slog.Logger) func(r *Rates) {
	return func(r *Rates) {
		r.logger = logger
		r.h.logger = logger
	}
}

func New(db *sql.DB, options ...Option) *Rates {
	repo := newRepository(db)
	svc := newService(repo)
	h := newHandler(svc)
	r := &Rates{repo: repo, svc: svc, h: h, logger: slog.Default()}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(r)
	}

	return r
}

func (r *Rates) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/rates/latest", r.h.getLatest)
	mux.HandleFunc("GET /api/v1/rates/history/{currency}", r.h.getHistory)
}

func (r *Rates) Save(ctx context.Context, rate Rate) error {
	return r.repo.Save(ctx, rate)
}
