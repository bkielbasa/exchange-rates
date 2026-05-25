package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Open opens a MySQL connection pool through otelsql so all queries are traced
// and metered, then pings it.
func Open(dsn string) (*sql.DB, error) {
	pool, err := otelsql.Open("mysql", dsn,
		otelsql.WithAttributes(semconv.DBSystemMySQL),
	)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	if _, err := otelsql.RegisterDBStatsMetrics(pool,
		otelsql.WithAttributes(semconv.DBSystemMySQL),
	); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("register db stats metrics: %w", err)
	}

	pool.SetMaxOpenConns(20)
	pool.SetMaxIdleConns(10)
	pool.SetConnMaxLifetime(5 * time.Minute)
	if err := pool.Ping(); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return pool, nil
}
