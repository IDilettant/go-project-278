//go:build integration

package handlers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"code/internal/app/links"
	testhttp "code/internal/testing/httptest"
)

func TestLinksAPI_Errors(t *testing.T) {
	resetLinks(t)

	invalidRequestTests := []struct {
		name    string
		method  string
		path    string
		headers map[string]string
		body    any
		rawBody string
	}{
		{
			name:   "invalid_json",
			method: http.MethodPost,
			path:   apiLinksPath,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			rawBody: "{not-json",
		},
		{
			name:   "strict_json_unknown_field",
			method: http.MethodPost,
			path:   apiLinksPath,
			body: map[string]any{
				"original_url": "https://example.com",
				"extra":        "nope",
			},
		},
		{
			name:   "strict_json_extra_object",
			method: http.MethodPost,
			path:   apiLinksPath,
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			rawBody: `{"original_url":"https://example.com"}{"short_name":"good"}`,
		},
	}

	for _, tc := range invalidRequestTests {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			switch {
			case tc.rawBody != "":
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.rawBody))
			case tc.body != nil:
				req = testhttp.NewJSONRequest(t, tc.method, tc.path, tc.body)
			default:
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			requireInvalidRequest(t, rec)
		})
	}

	problemTests := []struct {
		name    string
		method  string
		path    string
		headers map[string]string
		body    any
		rawBody string
		status  int
		typeID  string
		title   string
		detail  string
	}{
		{
			name:   "invalid_range",
			method: http.MethodGet,
			path:   apiLinksPath + "?range=bad",
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid range",
		},
		{
			name:   "invalid_range_limit",
			method: http.MethodGet,
			path:   apiLinksPath + "?range=[0,1001]",
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid range",
		},
		{
			name:   "invalid_sort",
			method: http.MethodGet,
			path:   apiLinksPath + "?sort=bad",
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid sort",
		},
		{
			name:   "invalid_sort_unknown_field",
			method: http.MethodGet,
			path:   apiLinksPath + "?sort=" + sortJSONRaw(t, "nope", string(links.SortAsc)),
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid sort",
		},
		{
			name:   "invalid_sort_order",
			method: http.MethodGet,
			path:   apiLinksPath + "?sort=" + sortJSONRaw(t, string(links.SortFieldID), "DOWN"),
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid sort",
		},
		{
			name:   "invalid_id",
			method: http.MethodGet,
			path:   apiLinksPath + "/abc",
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid id",
		},
		{
			name:   "invalid_id_update",
			method: http.MethodPut,
			path:   apiLinksPath + "/abc",
			body: map[string]any{
				"original_url": "https://example.com",
				"short_name":   "good",
			},
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid id",
		},
		{
			name:   "invalid_id_delete",
			method: http.MethodDelete,
			path:   apiLinksPath + "/abc",
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid id",
		},
		{
			name:   "not_found",
			method: http.MethodPut,
			path:   apiLinksPath + "/999999",
			body: map[string]any{
				"original_url": "https://example.com/x",
				"short_name":   "zzzz",
			},
			status: http.StatusNotFound,
			typeID: "about:blank",
			title:  "Not Found",
			detail: "not found",
		},
		{
			name:   "not_found_delete",
			method: http.MethodDelete,
			path:   apiLinksPath + "/999999",
			status: http.StatusNotFound,
			typeID: "about:blank",
			title:  "Not Found",
			detail: "not found",
		},
	}

	for _, tc := range problemTests {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request

			switch {
			case tc.rawBody != "":
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString(tc.rawBody))
			case tc.body != nil:
				req = testhttp.NewJSONRequest(t, tc.method, tc.path, tc.body)
			default:
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, tc.status, rec.Body.String())
			}

			p := requireProblem(t, rec, tc.status, tc.typeID)
			require.Equal(t, tc.title, p.Title)
			require.Equal(t, tc.detail, p.Detail)
		})
	}
}

func TestLinkVisitsAPI_Errors(t *testing.T) {
	resetLinks(t)

	problemTests := []struct {
		name   string
		method string
		path   string
		status int
		typeID string
		title  string
		detail string
	}{
		{
			name:   "invalid_sort",
			method: http.MethodGet,
			path:   apiLinkVisitsPath + "?sort=bad",
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid sort",
		},
		{
			name:   "invalid_sort_unknown_field",
			method: http.MethodGet,
			path:   apiLinkVisitsPath + "?sort=" + sortJSONRaw(t, "nope", string(links.SortAsc)),
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid sort",
		},
		{
			name:   "invalid_sort_order",
			method: http.MethodGet,
			path:   apiLinkVisitsPath + "?sort=" + sortJSONRaw(t, string(links.SortFieldID), "DOWN"),
			status: http.StatusBadRequest,
			typeID: "validation_error",
			title:  "Validation error",
			detail: "invalid sort",
		},
	}

	for _, tc := range problemTests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != tc.status {
				t.Fatalf("unexpected status: got=%d want=%d body=%s", rec.Code, tc.status, rec.Body.String())
			}

			p := requireProblem(t, rec, tc.status, tc.typeID)
			require.Equal(t, tc.title, p.Title)
			require.Equal(t, tc.detail, p.Detail)
		})
	}
}
