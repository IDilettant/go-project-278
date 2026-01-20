package apiapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	httpapi "code/internal/adapters/http"
	"code/internal/adapters/http/plugins"
	pgrepo "code/internal/adapters/postgres"
	"code/internal/app/links"
	"code/internal/platform/config"
	"code/internal/platform/postgres"
)

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

	repo := pgrepo.NewRepo(db)
	svc := links.New(repo)

	r := httpapi.NewEngine(
		plugins.Logger(),
		plugins.Sentry(cfg.SentryMiddlewareTimeout),
		plugins.Recovery(),
		plugins.RequestTimeout(cfg.RequestBudget),
		plugins.CORS(cfg.CORSAllowedOrigins),
	)

	httpapi.RegisterRoutes(r, httpapi.RouterDeps{
		Links:   svc,
		BaseURL: cfg.BaseURL,
	})

	return &App{cfg: cfg, db: db, router: r}, nil
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
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("http server: %w", err)

	case <-ctx.Done():
		return gracefulShutdown(ctx, srv, a.cfg.HTTPShutdownTimeout, errCh)
	}
}

func gracefulShutdown(ctx context.Context, srv *http.Server, timeout time.Duration, errCh <-chan error) error {
	srv.SetKeepAlivesEnabled(false)

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), timeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("http shutdown timed out; forced close: %w", err)
		}

		return fmt.Errorf("http shutdown failed; forced close: %w", err)
	}

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return fmt.Errorf("http server stopped with error: %w", err)
	default:
		return nil
	}
}
