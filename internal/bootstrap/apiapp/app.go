package apiapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"

	"code/internal/app/links"
	"code/internal/config"
	"code/internal/repository/postgres"
	"code/internal/transport/httpapi"
)

// App is a composition root for the HTTP API.
type App struct {
	cfg    config.Config
	db     *sql.DB
	router http.Handler
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	if err := sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN}); err != nil {
		return nil, fmt.Errorf("init sentry: %w", err)
	}

	db, err := postgres.Open(ctx, postgres.OpenConfig{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	})
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	repo := postgres.NewRepo(db)
	svc := links.New(repo)

	router := httpapi.NewRouter(httpapi.RouterDeps{
		Links:                   svc,
		BaseURL:                 cfg.BaseURL,
		SentryMiddlewareTimeout: cfg.SentryMiddlewareTimeout,
	})

	return &App{cfg: cfg, db: db, router: router}, nil
}

func (a *App) Close() error {
	sentry.Flush(a.cfg.SentryFlushTimeout)

	if a.db == nil {
		return nil
	}
	
	return a.db.Close()
}

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:              a.cfg.Port,
		Handler:           a.router,
		ReadHeaderTimeout: a.cfg.HTTPReadHeaderTimeout,
		ReadTimeout:       a.cfg.HTTPReadTimeout,
		WriteTimeout:      a.cfg.HTTPWriteTimeout,
		IdleTimeout:       a.cfg.HTTPIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTPShutdownTimeout)
		defer cancel()

		_ = srv.Shutdown(shutdownCtx)

		// Wait for ListenAndServe to return.
		err := <-errCh
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		
		return err

	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		
		return err
	}
}
