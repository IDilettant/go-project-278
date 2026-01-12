//go:build integration

package middleware_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"code/internal/app/links"
	"code/internal/domain"
	"code/internal/testutils"
	"code/internal/transport/httpapi"
)

type slowRepo struct {
	db    *sql.DB
	errCh chan error
}

func (r slowRepo) ListAll(ctx context.Context) ([]domain.Link, error) {
	_, err := r.db.ExecContext(ctx, "SELECT pg_sleep($1)", 0.2)
	select {
	case r.errCh <- err:
	default:
	}
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (slowRepo) ListPage(ctx context.Context, _, _ int32) ([]domain.Link, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (slowRepo) Count(ctx context.Context) (int64, error) {
	<-ctx.Done()
	return 0, ctx.Err()
}

func (slowRepo) GetByID(_ context.Context, _ int64) (domain.Link, error) {
	return domain.Link{}, domain.ErrNotFound
}

func (slowRepo) GetByShortName(_ context.Context, _ string) (domain.Link, error) {
	return domain.Link{}, domain.ErrNotFound
}

func (slowRepo) Create(_ context.Context, _, _ string) (domain.Link, error) {
	return domain.Link{}, domain.ErrShortNameConflict
}

func (slowRepo) Update(_ context.Context, _ int64, _, _ string) (domain.Link, error) {
	return domain.Link{}, domain.ErrNotFound
}

func (slowRepo) Delete(_ context.Context, _ int64) error {
	return domain.ErrNotFound
}

func TestAPI_RequestTimeout_CancelsDBQuery(t *testing.T) {
	ctx := context.Background()

	pgC, err := tcpg.RunContainer(
		ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpg.WithDatabase("appdb"),
		tcpg.WithUsername("postgres"),
		tcpg.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = pgC.Terminate(ctx) })

	dsn, err := pgC.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	db, err := openDBWithRetry(ctx, dsn, 10*time.Second)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	errCh := make(chan error, 1)
	repo := slowRepo{db: db, errCh: errCh}
	svc := links.New(repo)
	router := httpapi.NewRouter(httpapi.RouterDeps{
		Links:                   svc,
		BaseURL:                 "http://localhost:8080",
		SentryMiddlewareTimeout: time.Second,
		RequestTimeout:          50 * time.Millisecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	p := testutils.RequireProblem(t, rec.Result(), http.StatusGatewayTimeout, "timeout")
	require.Equal(t, "Gateway Timeout", p.Title)
	require.Equal(t, "", p.Detail)

	select {
	case err := <-errCh:
		require.Error(t, err)
		require.True(t, errors.Is(err, context.DeadlineExceeded), fmt.Sprintf("unexpected error: %v", err))
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for db error")
	}
}

func openDBWithRetry(ctx context.Context, dsn string, timeout time.Duration) (*sql.DB, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("open db with retry (timeout=%s): %w", timeout, err)
		}

		db, err := sql.Open("pgx", dsn)
		if err != nil {
			lastErr = err
			if err := testutils.Sleep(ctx, 200*time.Millisecond); err != nil {
				return nil, fmt.Errorf("open db with retry (timeout=%s): %w", timeout, err)
			}

			continue
		}

		pctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		err = db.PingContext(pctx)
		cancel()
		if err == nil {
			return db, nil
		}

		lastErr = err
		_ = db.Close()
		if err := testutils.Sleep(ctx, 200*time.Millisecond); err != nil {
			return nil, fmt.Errorf("open db with retry (timeout=%s): %w", timeout, err)
		}
	}

	if lastErr == nil {
		lastErr = context.DeadlineExceeded
	}

	return nil, fmt.Errorf("open db with retry (timeout=%s): %w", timeout, lastErr)
}
