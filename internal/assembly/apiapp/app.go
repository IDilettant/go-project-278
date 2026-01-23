package apiapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"

	httpapi "code/internal/adapters/httpapi"
	"code/internal/adapters/httpapi/handlers"
	"code/internal/adapters/httpapi/stack"
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

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	handlers.InitValidation()

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
	visitsRepo := pgrepo.NewLinkVisitsRepo(db)

	appLogger := linksSlogLogger{l: logger}
	svc := links.New(repo, visitsRepo, appLogger)

	r := httpapi.NewEngine(
		stack.Logger(),
		stack.RequestID(),
		stack.Sentry(cfg.SentryMiddlewareTimeout),
		stack.Recovery(),
		stack.RequestTimeout(cfg.RequestBudget),
		stack.CORS(cfg.CORSAllowedOrigins),
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
		Addr:              a.cfg.HTTPAddr,
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
