//go:build integration

package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"code/internal/app/links"
	"code/internal/domain"
	"code/internal/transport/httpapi"
)

type timeoutRepo struct{}

func (timeoutRepo) List(ctx context.Context) ([]domain.Link, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (timeoutRepo) GetByID(_ context.Context, _ int64) (domain.Link, error) {
	return domain.Link{}, domain.ErrNotFound
}

func (timeoutRepo) GetByShortName(_ context.Context, _ string) (domain.Link, error) {
	return domain.Link{}, domain.ErrNotFound
}

func (timeoutRepo) Create(_ context.Context, _, _ string) (domain.Link, error) {
	return domain.Link{}, domain.ErrShortNameConflict
}

func (timeoutRepo) Update(_ context.Context, _ int64, _, _ string) (domain.Link, error) {
	return domain.Link{}, domain.ErrNotFound
}

func (timeoutRepo) Delete(_ context.Context, _ int64) error {
	return domain.ErrNotFound
}

func TestAPI_RequestTimeout(t *testing.T) {
	svc := links.New(timeoutRepo{})
	r := httpapi.NewRouter(httpapi.RouterDeps{
		Links:                   svc,
		BaseURL:                 "http://localhost:8080",
		SentryMiddlewareTimeout: time.Second,
		RequestTimeout:          50 * time.Millisecond,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	requireProblem(t, rec, http.StatusGatewayTimeout, "timeout", "Gateway Timeout", "")
}
