package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"code/internal/httpserver"
)

const (
	defaultPort   = "8080"
	sentryTimeout = 2 * time.Second
)

func main() {
	err := initSentry()
	if err != nil {
		log.Fatal(err)
	}
	defer sentry.Flush(sentryTimeout)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r := httpserver.NewRouter()

	err = r.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}

func initSentry() error {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return errors.New("SENTRY_DSN is empty")
	}

	err := sentry.Init(sentry.ClientOptions{Dsn: dsn})
	if err != nil {
		return fmt.Errorf("init sentry: %w", err)
	}

	return nil
}

