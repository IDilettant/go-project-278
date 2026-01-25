package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"code/internal/assembly/apiapp"
	"code/internal/platform/config"
)

// main duplicates cmd/api entrypoint so `go run .` works in CI containers.
func main() {
	if err := run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	app, err := apiapp.New(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer func() {
		_ = app.Close()
	}()

	if err := app.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
