//go:build integration

package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	httpapi "code/internal/adapters/http"
	"code/internal/adapters/http/plugins"
	"code/internal/app/links"
	"code/internal/testutils"
)

func TestAPI_PanicRecovery_ReturnsProblemJSON(t *testing.T) {
	svc := links.New(timeoutRepo{}, nil, nil)
	router := httpapi.NewEngine(
		plugins.Recovery(),
		plugins.RequestTimeout(50*time.Millisecond),
	)
	httpapi.RegisterRoutes(router, httpapi.RouterDeps{
		Links:   svc,
		BaseURL: "http://localhost:8080",
	})

	router.GET("/panic", func(c *gin.Context) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	p := testutils.RequireProblem(t, rec.Result(), http.StatusInternalServerError, "internal_error")
	require.Equal(t, "Internal Server Error", p.Title)
	require.Equal(t, "internal error", p.Detail)
}
