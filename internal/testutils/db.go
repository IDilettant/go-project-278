package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"code/internal/platform/postgres"
)

type DBRetryConfig struct {
	Timeout time.Duration
	Backoff time.Duration
}

func DefaultDBRetryConfig() DBRetryConfig {
	return DBRetryConfig{
		Timeout: 10 * time.Second,
		Backoff: 200 * time.Millisecond,
	}
}

func OpenDBWithRetry(ctx context.Context, cfg postgres.OpenConfig, rc DBRetryConfig) (*sql.DB, error) {
	deadline := time.Now().Add(rc.Timeout)

	var lastErr error

	for time.Now().Before(deadline) {
		db, err := postgres.Open(ctx, cfg)
		if err == nil {
			return db, nil
		}

		lastErr = err
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("open db with retry: %w", ctx.Err())
		case <-time.After(rc.Backoff):
		}
	}

	return nil, fmt.Errorf("open db with retry (timeout=%s): %w", rc.Timeout, lastErr)
}
