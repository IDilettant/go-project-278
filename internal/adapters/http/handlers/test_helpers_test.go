//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"code/internal/testutils"
)

const (
	apiLinksPath      = "/api/links"
	redirectPathPrefx = "/r/"
)

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

func requireProblem(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantType string) testutils.Problem {
	t.Helper()

	return testutils.RequireProblem(t, rec.Result(), wantStatus, wantType)
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
