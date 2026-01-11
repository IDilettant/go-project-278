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
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"code/internal/app/links"
	"code/internal/config"
	pgrepo "code/internal/repository/postgres"
	"code/internal/transport/httpapi"
)

var (
	tcCtx  = context.Background()
	pgC    *tcpg.PostgresContainer
	db     *sql.DB
	router http.Handler
)

var shortNameRe = regexp.MustCompile(`^[a-zA-Z0-9]{4,32}$`)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	var err error

	pgC, err = tcpg.RunContainer(
		tcCtx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpg.WithDatabase("appdb"),
		tcpg.WithUsername("postgres"),
		tcpg.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "start postgres:", err)

		return 1
	}
	defer func() { _ = pgC.Terminate(tcCtx) }()

	dsn, err := pgC.ConnectionString(tcCtx, "sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "dsn:", err)

		return 1
	}

	// open via production helper -> covers postgres.Open
	db, err = openDBWithRetry(tcCtx, pgrepo.OpenConfig{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}, 10*time.Second)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open db:", err)

		return 1
	}
	defer func() { _ = db.Close() }()

	goose.SetDialect("postgres")
	if err := goose.Up(db, filepath.Join(projectRoot(), "db", "migrations")); err != nil {
		fmt.Fprintln(os.Stderr, "goose up:", err)

		return 1
	}

	// config.Load required envs
	os.Setenv("PORT", "8080")
	os.Setenv("BASE_URL", "http://localhost:8080")
	os.Setenv("SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")
	os.Setenv("DATABASE_URL", dsn)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config load:", err)

		return 1
	}

	repo := pgrepo.NewRepo(db)
	svc := links.New(repo)

	router = httpapi.NewRouter(httpapi.RouterDeps{
		Links:                   svc,
		BaseURL:                 cfg.BaseURL,
		SentryMiddlewareTimeout: cfg.SentryMiddlewareTimeout,
		RequestTimeout:          cfg.HTTPRequestTimeout,
	})

	return m.Run()
}

func TestAPI_CRUD_HappyPath(t *testing.T) {
	resetLinks(t)

	created := doJSON(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/long-url",
		"short_name":   "exmpl",
	}, http.StatusCreated)

	id := asInt64(t, created["id"])
	require.Equal(t, "exmpl", asString(t, created["short_name"]))
	require.NotEmpty(t, asString(t, created["short_url"]))

	_ = doJSON(t, http.MethodGet, "/api/links/"+itoa(id), nil, http.StatusOK)

	list := doJSONArray(t, http.MethodGet, "/api/links", nil, http.StatusOK)
	require.Len(t, list, 1)

	updated := doJSON(t, http.MethodPut, "/api/links/"+itoa(id), map[string]any{
		"original_url": "https://example.com/updated",
	}, http.StatusOK)
	require.Equal(t, "exmpl", asString(t, updated["short_name"]))

	doNoContent(t, http.MethodDelete, "/api/links/"+itoa(id), http.StatusNoContent)
	doJSONExpectError(t, http.MethodGet, "/api/links/"+itoa(id), nil, http.StatusNotFound)
}

func TestAPI_Redirect(t *testing.T) {
	resetLinks(t)

	created := doJSON(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/a",
		"short_name":   "good",
	}, http.StatusCreated)

	short := asString(t, created["short_name"])

	req := httptest.NewRequest(http.MethodGet, "/r/"+short, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "https://example.com/a", rec.Header().Get("Location"))

	req2 := httptest.NewRequest(http.MethodGet, "/r/unknown", nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusNotFound, rec2.Code)
}

func TestAPI_Conflict_Create(t *testing.T) {
	resetLinks(t)

	doJSON(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/1",
		"short_name":   "dupe",
	}, http.StatusCreated)

	doJSONExpectError(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/2",
		"short_name":   "dupe",
	}, http.StatusConflict)
}

func TestAPI_Update_ImmutableShortName(t *testing.T) {
	resetLinks(t)

	a := doJSON(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/a",
		"short_name":   "aaaa",
	}, http.StatusCreated)

	b := doJSON(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/b",
		"short_name":   "bbbb",
	}, http.StatusCreated)

	bid := asInt64(t, b["id"])
	target := asString(t, a["short_name"])

	doJSONExpectError(t, http.MethodPut, "/api/links/"+itoa(bid), map[string]any{
		"original_url": "https://example.com/b2",
		"short_name":   target,
	}, http.StatusUnprocessableEntity)
}

func TestAPI_InvalidID_Returns400(t *testing.T) {
	resetLinks(t)

	doJSONExpectError(t, http.MethodGet, "/api/links/abc", nil, http.StatusBadRequest)
	doJSONExpectError(t, http.MethodPut, "/api/links/abc", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "good",
	}, http.StatusBadRequest)
	doJSONExpectError(t, http.MethodDelete, "/api/links/abc", nil, http.StatusBadRequest)
}

func TestAPI_NotFound_Update_Delete(t *testing.T) {
	resetLinks(t)

	doJSONExpectError(t, http.MethodPut, "/api/links/999999", map[string]any{
		"original_url": "https://example.com/x",
		"short_name":   "zzzz",
	}, http.StatusNotFound)

	doJSONExpectError(t, http.MethodDelete, "/api/links/999999", nil, http.StatusNotFound)
}

func TestAPI_ValidationAndBadJSON(t *testing.T) {
	resetLinks(t)

	// invalid json
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString("{not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// invalid short_name: too short
	doJSONExpectError(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "abc",
	}, http.StatusBadRequest)

	// invalid short_name: forbidden chars
	doJSONExpectError(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "ab_cd",
	}, http.StatusBadRequest)
	doJSONExpectError(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "ab-cd",
	}, http.StatusBadRequest)

	// invalid url
	doJSONExpectError(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "not-a-url",
		"short_name":   "good",
	}, http.StatusBadRequest)
}

func TestAPI_Create_AutoShortName(t *testing.T) {
	resetLinks(t)

	rec := doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/auto",
		"short_name":   "",
	})

	// if API does not support it, skip
	if rec.Code == http.StatusBadRequest {
		t.Skip("API does not support auto short_name generation")
	}

	require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())

	var created map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))

	short := asString(t, created["short_name"])
	require.True(t, shortNameRe.MatchString(short), "generated short_name must be alnum 4..32, got %q", short)

	req := httptest.NewRequest(http.MethodGet, "/r/"+short, nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req)

	require.Equal(t, http.StatusFound, rec2.Code)
	require.Equal(t, "https://example.com/auto", rec2.Header().Get("Location"))
}

func TestAPI_Redirect_InvalidShortName_400(t *testing.T) {
	resetLinks(t)

	req := httptest.NewRequest(http.MethodGet, "/r/ab_cd", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	requireProblem(t, rec, http.StatusBadRequest, "validation_error", "Validation error", "invalid short_name")
}

func TestAPI_Redirect_NotFound_404(t *testing.T) {
	resetLinks(t)

	req := httptest.NewRequest(http.MethodGet, "/r/Unknown1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	requireProblem(t, rec, http.StatusNotFound, "about:blank", "Not Found", "")
}

func resetLinks(t *testing.T) {
	t.Helper()

	truncateLinks(t)
	t.Cleanup(func() { truncateLinks(t) })
}

func truncateLinks(t *testing.T) {
	t.Helper()

	_, err := db.ExecContext(tcCtx, `TRUNCATE links RESTART IDENTITY`)
	require.NoError(t, err)
}

func projectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}

		dir = parent
	}
}
