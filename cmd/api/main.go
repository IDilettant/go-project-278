package main

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"

	"code/internal/web"
)

const (
	defaultPort   = "8080"
	sentryTimeout = 2 * time.Second
)

func main() {
	cleanup, err := initSentry()
	if err != nil {
		log.Printf("Sentry disabled: %v", err)
	}
	defer cleanup()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	r := web.NewRouter()

	err = r.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}

func initSentry() (func(), error) {
	dsn := os.Getenv("SENTRY_DSN")
	if dsn == "" {
		return func() {}, errors.New("SENTRY_DSN is empty")
	}

	if err := sentry.Init(sentry.ClientOptions{Dsn: dsn}); err != nil {
		return func() {}, err
	}

	return func() { sentry.Flush(sentryTimeout) }, nil
}
