//go:build integration

package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
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

func createLink(t *testing.T, originalURL, shortName string) int64 {
	t.Helper()

	payload := map[string]any{
		"original_url": originalURL,
	}
	if shortName != "" {
		payload["short_name"] = shortName
	}

	rec := doRequest(t, http.MethodPost, apiLinksPath, payload)
	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	location := rec.Header().Get("Location")
	require.NotEmpty(t, location)
	require.True(t, strings.HasPrefix(location, apiLinksPath+"/"))

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	id := asInt64(t, body["id"])
	require.NotZero(t, id)
	require.Equal(t, originalURL, asString(t, body["original_url"]))

	gotShortName := asString(t, body["short_name"])
	if shortName == "" {
		require.Regexp(t, shortNameRe, gotShortName)
	} else {
		require.Equal(t, shortName, gotShortName)
	}

	require.Equal(t, fmt.Sprintf("%s/r/%s", "http://localhost:8080", gotShortName), asString(t, body["short_url"]))

	idStr := strings.TrimPrefix(location, apiLinksPath+"/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/%d", apiLinksPath, id), location)
	require.Equal(t, id, asInt64(t, body["id"]))

	_ = doJSON(t, http.MethodGet, apiLinksPath+"/"+idStr, nil, http.StatusOK)

	return id
}

func getSingleLink(t *testing.T) map[string]any {
	t.Helper()

	list := doJSONArray(t, http.MethodGet, apiLinksPath, nil, http.StatusOK)
	require.Len(t, list, 1)

	return list[0]
}

func getLinkByShortName(t *testing.T, shortName string) map[string]any {
	t.Helper()

	list := doJSONArray(t, http.MethodGet, apiLinksPath, nil, http.StatusOK)
	for _, item := range list {
		if asString(t, item["short_name"]) == shortName {
			return item
		}
	}

	require.Failf(t, "link not found", "short_name=%s", shortName)

	return nil
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
