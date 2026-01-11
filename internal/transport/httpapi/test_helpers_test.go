//go:build integration

package httpapi_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	pgrepo "code/internal/repository/postgres"
)

type problemResponse struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

func doRequest(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	return rec
}

func doJSON(t *testing.T, method, path string, body any, want int) map[string]any {
	t.Helper()

	rec := doRequest(t, method, path, body)
	require.Equal(t, want, rec.Code, rec.Body.String())

	if want == http.StatusNoContent {
		return nil
	}

	var out map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))

	return out
}

func doJSONArray(t *testing.T, method, path string, body any, want int) []map[string]any {
	t.Helper()

	rec := doRequest(t, method, path, body)
	require.Equal(t, want, rec.Code, rec.Body.String())

	var out []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &out))

	return out
}

func doNoContent(t *testing.T, method, path string, want int) {
	t.Helper()

	rec := doRequest(t, method, path, nil)
	require.Equal(t, want, rec.Code, rec.Body.String())
}

func doJSONExpectError(t *testing.T, method, path string, body any, want int) {
	t.Helper()

	_ = doJSON(t, method, path, body, want)
}

func requireProblem(
	t *testing.T,
	rec *httptest.ResponseRecorder,
	wantStatus int,
	wantType, wantTitle, wantDetail string,
) {
	t.Helper()

	require.Contains(t, rec.Header().Get("Content-Type"), "application/problem+json")

	var p problemResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &p))

	require.Equal(t, wantStatus, p.Status)
	require.Equal(t, wantType, p.Type)
	require.Equal(t, wantTitle, p.Title)
	require.Equal(t, wantDetail, p.Detail)
}

func openDBWithRetry(ctx context.Context, cfg pgrepo.OpenConfig, timeout time.Duration) (*sql.DB, error) {
	deadline := time.Now().Add(timeout)
	backoff := 200 * time.Millisecond
	var lastErr error
	for time.Now().Before(deadline) {
		attemptCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		db, err := pgrepo.Open(attemptCtx, cfg)
		cancel()
		if err == nil {
			return db, nil
		}
		lastErr = err
		time.Sleep(backoff + backoff/4)
		if backoff < time.Second {
			backoff *= 2
			if backoff > time.Second {
				backoff = time.Second
			}
		}
	}

	if lastErr == nil {
		lastErr = context.DeadlineExceeded
	}

	return nil, fmt.Errorf("open db with retry (timeout=%s): %w", timeout, lastErr)
}

func itoa(v int64) string { return fmt.Sprintf("%d", v) }

func asString(t *testing.T, v any) string {
	t.Helper()

	s, ok := v.(string)
	require.True(t, ok, "expected string, got %T (%v)", v, v)

	return s
}

func asInt64(t *testing.T, v any) int64 {
	t.Helper()

	f, ok := v.(float64)
	require.True(t, ok, "expected number(float64), got %T (%v)", v, v)

	return int64(f)
}
