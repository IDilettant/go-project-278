package testutils

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type Problem struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func DecodeProblem(t *testing.T, resp *http.Response) Problem {
	t.Helper()

	require.Equal(t, "application/problem+json", resp.Header.Get("Content-Type"))

	var p Problem
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&p))

	return p
}

func RequireProblem(t *testing.T, resp *http.Response, wantStatus int, wantType string) Problem {
	t.Helper()

	require.Equal(t, wantStatus, resp.StatusCode)

	p := DecodeProblem(t, resp)
	require.Equal(t, wantStatus, p.Status)
	require.Equal(t, wantType, p.Type)
	require.NotEmpty(t, p.Title)

	return p
}
