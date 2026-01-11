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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	app, err := apiapp.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = app.Close()
	}()

	err = app.Run(ctx)
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
        	log.Print("server stopped")

        	return
    	}

		log.Fatal(err)
	}
}
