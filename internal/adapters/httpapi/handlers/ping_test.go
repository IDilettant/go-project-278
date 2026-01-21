package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"code/internal/adapters/httpapi/handlers"
)

func TestPing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	h := handlers.New(nil, "")
	router.GET("/ping", h.Ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "text/plain; charset=utf-8", rec.Header().Get("Content-Type"))
	require.Equal(t, "pong", rec.Body.String())
}
