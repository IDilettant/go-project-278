package httptest

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func NewJSONRequest(t *testing.T, method, url string, body any) *http.Request {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req, err := http.NewRequest(method, url, &buf)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	return req
}

func DecodeJSON[T any](t *testing.T, r io.Reader) T {
	t.Helper()

	var v T
	require.NoError(t, json.NewDecoder(r).Decode(&v))

	return v
}
