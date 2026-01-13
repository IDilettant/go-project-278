// @title URL Shortener API
// @version 1.0
// @description URL shortener service API.
// @description
// @BasePath /
// @schemes http https
//
//go:generate sh -c "swag init -g ./main.go -o ../../docs/openapi --outputTypes json && mv ../../docs/openapi/swagger.json ../../docs/openapi/openapi.json"
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"code/internal/bootstrap/apiapp"
	"code/internal/config"
)

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

	app, err := apiapp.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		_ = app.Close()
	}()

	err = app.Run(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
