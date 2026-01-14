//go:build integration

package apiapp_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"code/internal/assembly/apiapp"
	"code/internal/platform/config"
	"code/internal/platform/postgres"
	"code/internal/testutils"
)

func TestApp_New_Run_Close(t *testing.T) {
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

	db, err := testutils.OpenDBWithRetry(ctx, postgres.OpenConfig{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}, testutils.DBRetryConfig{
		Timeout: 10 * time.Second,
		Backoff: 200 * time.Millisecond,
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	goose.SetDialect("postgres")
	migrationsDir := filepath.Join(projectRoot(t), "db", "migrations")
	require.NoError(t, goose.Up(db, migrationsDir))

	// required envs for config.Load()
	t.Setenv("PORT", "8080")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")
	t.Setenv("DATABASE_URL", dsn)

	cfg, err := config.Load()
	require.NoError(t, err)

	app, err := apiapp.New(ctx, cfg)
	require.NoError(t, err)
	t.Cleanup(func() { _ = app.Close() })

	// Run should stop on ctx cancel
	runCtx, cancel := context.WithCancel(ctx)
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err = app.Run(runCtx)
	require.NoError(t, err)
}

func projectRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	require.NoError(t, err)

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod not found from working dir")
		}
		dir = parent
	}
}
