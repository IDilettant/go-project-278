//go:build integration

package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	httpapi "code/internal/adapters/http"
	"code/internal/adapters/http/plugins"
	"code/internal/app/links"
	"code/internal/domain"
	"code/internal/testutils"
)

type timeoutRepo struct{}

func (timeoutRepo) ListAll(ctx context.Context) ([]domain.Link, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (timeoutRepo) ListPage(ctx context.Context, _, _ int32) ([]domain.Link, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (timeoutRepo) Count(ctx context.Context) (int64, error) {
	<-ctx.Done()
	return 0, ctx.Err()
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
	router := httpapi.NewEngine(
		plugins.Recovery(),
		plugins.RequestTimeout(50*time.Millisecond),
	)
	httpapi.RegisterRoutes(router, httpapi.RouterDeps{
		Links:   svc,
		BaseURL: "http://localhost:8080",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusGatewayTimeout, rec.Code)
	p := testutils.RequireProblem(t, rec.Result(), http.StatusGatewayTimeout, "timeout")
	require.Equal(t, "Gateway Timeout", p.Title)
	require.Equal(t, "timeout", p.Detail)
}
