package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bklimczak/exchange-rates/internal/cli/fetch"
	"github.com/bklimczak/exchange-rates/internal/cli/serve"
	"github.com/bklimczak/exchange-rates/internal/db"
	"github.com/bklimczak/exchange-rates/internal/observability"
)

func main() {
	if err := dispatch(os.Args[1:]); err != nil {
		log.Fatalf("cli: %v", err)
	}
}

func dispatch(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no subcommand provided")
	}

	// in real world Cobra urfave is often better option but IMO not needed here
	sub := args[0]
	switch sub {
	case "serve":
		return runServe()
	case "fetch":
		return runFetch()
	case "-h", "--help", "help":
		return nil
	default:
		return fmt.Errorf("unknown subcommand: %s", sub)
	}
}

func runServe() error {
	// ctrl+c or sigterm frok k8s
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, obsShutdown, err := observability.Init(ctx, "serve")
	if err != nil {
		return fmt.Errorf("observability init: %w", err)
	}
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = obsShutdown(shutdownCtx)
	}()

	cfg, err := serve.LoadConfig()
	if err != nil {
		return err
	}

	pool, err := db.Open(cfg.MySQLDSN)
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := serve.New(cfg, logger, pool)
	if err := srv.Start(); err != nil {
		return err
	}

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, c := context.WithTimeout(context.Background(), 30*time.Second)
	defer c()
	return srv.Shutdown(shutdownCtx)
}

func runFetch() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger, obsShutdown, err := observability.Init(ctx, "fetch")
	if err != nil {
		return fmt.Errorf("observability init: %w", err)
	}
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = obsShutdown(shutdownCtx)
	}()

	cfg, err := fetch.LoadConfig()
	if err != nil {
		return err
	}

	return fetch.Run(ctx, cfg, logger)
}
