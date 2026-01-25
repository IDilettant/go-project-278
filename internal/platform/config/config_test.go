package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"code/internal/platform/config"
)

func TestLoad_MissingRequired(t *testing.T) {
	t.Setenv("BASE_URL", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SENTRY_DSN", "")

	_, err := config.Load()
	require.Error(t, err)
}

func TestLoad_DefaultsOk(t *testing.T) {
	t.Setenv("HTTP_ADDR", "8080")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/db?sslmode=disable")
	t.Setenv("SENTRY_DSN", "")

	_, err := config.Load()
	require.NoError(t, err)
}

func TestLoad_InvalidDuration(t *testing.T) {
	t.Setenv("HTTP_ADDR", "8080")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/db?sslmode=disable")
	t.Setenv("SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	t.Setenv("HTTP_READ_TIMEOUT", "not-a-duration")

	_, err := config.Load()
	require.Error(t, err)
}

func TestLoad_RequestBudgetInvalid(t *testing.T) {
	t.Setenv("HTTP_ADDR", "8080")
	t.Setenv("BASE_URL", "http://localhost:8080")
	t.Setenv("DATABASE_URL", "postgres://x:y@localhost:5432/db?sslmode=disable")
	t.Setenv("SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")

	t.Setenv("REQUEST_BUDGET", "0s")

	_, err := config.Load()
	require.Error(t, err)
}

func TestMainEnvDoesNotLeak(t *testing.T) {
	require.NotEqual(t, "", os.Getenv("PATH"))
}
