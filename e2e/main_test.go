//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	migmysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	"github.com/bklimczak/exchange-rates/internal/cli/serve"
	"github.com/bklimczak/exchange-rates/internal/db"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var (
	testDB  *sql.DB
	baseURL string
)

// there are a few ways we can handle the problem with running those tests.
// The biggest problem right now is we cannot safely run parallel tests right now.
// We'd have to run multiple MySQL instances what would let us run them in separate
// but it'd consume more resources.
//
// This is a homework so I didn't solve the problem here but at larger scale it'd be a problem to address
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	container, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("exchange_rates"),
		tcmysql.WithUsername("user"),
		tcmysql.WithPassword("pass"),
	)
	if err != nil {
		log.Fatalf("start mysql container: %v", err)
	}
	
	defer container.Terminate(context.Background())

	dsn, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		log.Fatalf("dsn: %v", err)
	}

	if err := applyMigrations(dsn); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	pool, err := db.Open(dsn)
	if err != nil {
		log.Fatalf("db.Open: %v", err)
	}
	defer pool.Close()
	testDB = pool

	cfg := serve.Config{HTTPAddr: ":0", MySQLDSN: dsn, LogLevel: "info"}
	srv := serve.New(cfg, slog.New(slog.NewTextHandler(io.Discard, nil)), pool)
	if err := srv.Start(); err != nil {
		log.Fatalf("serve.Start: %v", err)
	}
	baseURL = "http://" + srv.Addr()

	code := m.Run()

	shutdownCtx, c := context.WithTimeout(context.Background(), 10*time.Second)
	defer c()
	_ = srv.Shutdown(shutdownCtx)

	os.Exit(code)
}

func applyMigrations(dsn string) error {
	d, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("iofs source: %w", err)
	}

	// Use a raw *sql.DB (not the otelsql-wrapped pool) so migrate's MySQL
	// driver gets exactly the connection it expects.
	rawDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("sql.Open for migrate: %w", err)
	}
	defer rawDB.Close()

	drv, err := migmysql.WithInstance(rawDB, &migmysql.Config{})
	if err != nil {
		return fmt.Errorf("migrate mysql driver: %w", err)
	}

	mig, err := migrate.NewWithInstance("iofs", d, "mysql", drv)
	if err != nil {
		return fmt.Errorf("migrate.NewWithInstance: %w", err)
	}
	
	if err := mig.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}
