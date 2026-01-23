//go:build integration

package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpg "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	httpapi "code/internal/adapters/httpapi"
	"code/internal/adapters/httpapi/handlers"
	"code/internal/adapters/httpapi/stack"
	pgrepo "code/internal/adapters/postgres"
	"code/internal/app/links"
	"code/internal/platform/config"
	"code/internal/platform/postgres"
	"code/internal/testing/dbtest"
)

var (
	tcCtx  = context.Background()
	pgC    *tcpg.PostgresContainer
	db     *sql.DB
	router *gin.Engine
)

var shortNameRe = regexp.MustCompile(`^[a-zA-Z0-9]{3,32}$`)

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
	rc := dbtest.DefaultDBRetryConfig()
	rc.Timeout = 10 * time.Second
	db, err = dbtest.OpenDBWithRetry(tcCtx, postgres.OpenConfig{
		DSN:             dsn,
		MaxOpenConns:    5,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}, rc)
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
	os.Setenv("HTTP_ADDR", "8080")
	os.Setenv("BASE_URL", "http://localhost:8080")
	os.Setenv("SENTRY_DSN", "https://public@o0.ingest.sentry.io/0")
	os.Setenv("DATABASE_URL", dsn)

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config load:", err)

		return 1
	}

	handlers.InitValidation()

	repo := pgrepo.NewRepo(db)
	visitsRepo := pgrepo.NewLinkVisitsRepo(db)
	svc := links.New(repo, visitsRepo, nil)

	router = httpapi.NewEngine(
		stack.Logger(),
		stack.Sentry(cfg.SentryMiddlewareTimeout),
		stack.Recovery(),
		stack.RequestTimeout(cfg.RequestBudget),
		stack.CORS(cfg.CORSAllowedOrigins),
	)

	httpapi.RegisterRoutes(router, httpapi.RouterDeps{
		Links:   svc,
		BaseURL: cfg.BaseURL,
	})

	return m.Run()
}

func TestAPI_CRUD_HappyPath(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/long-url", "exmpl")

	list := doJSONArray(t, http.MethodGet, "/api/links", nil, http.StatusOK)
	require.Len(t, list, 1)

	created := list[0]
	id := asInt64(t, created["id"])
	originalShort := asString(t, created["short_name"])
	require.Equal(t, "exmpl", originalShort)
	require.Equal(t, "http://localhost:8080/r/"+originalShort, asString(t, created["short_url"]))

	_ = doJSON(t, http.MethodGet, "/api/links/"+itoa(id), nil, http.StatusOK)

	updated := doJSON(t, http.MethodPut, "/api/links/"+itoa(id), map[string]any{
		"original_url": "https://example.com/updated",
		"short_name":   "updated",
	}, http.StatusOK)
	updatedShort := asString(t, updated["short_name"])
	require.Equal(t, "updated", updatedShort)
	require.Equal(t, "http://localhost:8080/r/"+updatedShort, asString(t, updated["short_url"]))

	doNoContent(t, http.MethodDelete, "/api/links/"+itoa(id), http.StatusNoContent)
	doJSONExpectError(t, http.MethodGet, "/api/links/"+itoa(id), nil, http.StatusNotFound)
}

func TestAPI_ListLinks_Range(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 11)

	rec := doRequest(t, http.MethodGet, "/api/links?range=[0,4]", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links 0-4/11", rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 5)
	require.Equal(t, int64(1), asInt64(t, list[0]["id"]))
	require.Equal(t, int64(5), asInt64(t, list[4]["id"]))

	rec2 := doRequest(t, http.MethodGet, "/api/links?range=[5,10]", nil)
	require.Equal(t, http.StatusOK, rec2.Code)
	require.Equal(t, "links 5-10/11", rec2.Header().Get("Content-Range"))

	var list2 []map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &list2))
	require.Len(t, list2, 6)
	require.Equal(t, int64(6), asInt64(t, list2[0]["id"]))
	require.Equal(t, int64(11), asInt64(t, list2[5]["id"]))
}

func TestAPI_ListLinks_Range_Header(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 11)

	rec := doRequestWithHeaders(t, http.MethodGet, apiLinksPath, nil, map[string]string{
		"Range": "links=0-4",
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links 0-4/11", rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 5)
	require.Equal(t, int64(1), asInt64(t, list[0]["id"]))
	require.Equal(t, int64(5), asInt64(t, list[4]["id"]))
}

func TestAPI_ListLinks_Sort_RangeHeader(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 5)

	sortParam := sortJSON(t, string(links.SortFieldID), links.SortDesc)
	rec := doRequestWithHeaders(t, http.MethodGet, "/api/links?sort="+sortParam, nil, map[string]string{
		"Range": "links=0-1",
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links 0-1/5", rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 2)
	require.Equal(t, int64(5), asInt64(t, list[0]["id"]))
	require.Equal(t, int64(4), asInt64(t, list[1]["id"]))
}

func TestAPI_ListLinks_Range_WithFilter(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 11)

	rangeParam := url.QueryEscape(`[0,4]`)
	filterParam := url.QueryEscape(`{}`)

	rec := doRequest(t, http.MethodGet, "/api/links?range="+rangeParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links 0-4/11", rec.Header().Get("Content-Range"))

	var base []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &base))

	rec2 := doRequest(t, http.MethodGet, "/api/links?range="+rangeParam+"&filter="+filterParam, nil)
	require.Equal(t, http.StatusOK, rec2.Code)
	require.Equal(t, "links 0-4/11", rec2.Header().Get("Content-Range"))

	var withFilter []map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &withFilter))

	require.Equal(t, base, withFilter)
}

func TestAPI_ListLinks_SortByIDDesc(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 5)

	rangeParam := url.QueryEscape(`[0,4]`)
	sortParam := sortJSON(t, string(links.SortFieldID), links.SortDesc)
	rec := doRequest(t, http.MethodGet, "/api/links?range="+rangeParam+"&sort="+sortParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links 0-4/5", rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 5)
	require.Equal(t, int64(5), asInt64(t, list[0]["id"]))
	require.Equal(t, int64(1), asInt64(t, list[4]["id"]))
}

func TestAPI_ListLinks_SortByShortURL(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/one", "ccc")
	createLink(t, "https://example.com/two", "aaa")
	createLink(t, "https://example.com/three", "bbb")

	rangeParam := url.QueryEscape(`[0,2]`)
	sortParam := sortJSON(t, "short_url", links.SortAsc)
	rec := doRequest(t, http.MethodGet, "/api/links?range="+rangeParam+"&sort="+sortParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 3)
	require.Equal(t, "aaa", asString(t, list[0]["short_name"]))
	require.Equal(t, "ccc", asString(t, list[2]["short_name"]))
}

func TestAPI_ListLinks_Default(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 11)

	rec := doRequest(t, http.MethodGet, "/api/links", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Empty(t, rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 11)
	require.Equal(t, int64(1), asInt64(t, list[0]["id"]))
	require.Equal(t, int64(11), asInt64(t, list[10]["id"]))
}

func TestAPI_ListLinks_DefaultSort(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 3)

	rangeParam := url.QueryEscape(`[0,1]`)
	rec := doRequest(t, http.MethodGet, "/api/links?range="+rangeParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links 0-1/3", rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 2)
	require.Equal(t, int64(1), asInt64(t, list[0]["id"]))
	require.Equal(t, int64(2), asInt64(t, list[1]["id"]))
}

func TestAPI_ListLinks_EmptyRange(t *testing.T) {
	resetLinks(t)

	seedLinks(t, 11)

	rec := doRequest(t, http.MethodGet, "/api/links?range=[20,30]", nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "links */11", rec.Header().Get("Content-Range"))

	var list []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &list))
	require.Len(t, list, 0)
}

func TestAPI_CreateLink_ConcurrentConflict(t *testing.T) {
	resetLinks(t)

	const (
		workers   = 5
		shortName = "dupe"
	)

	payload, err := json.Marshal(map[string]any{
		"original_url": "https://example.com/long-url",
		"short_name":   shortName,
	})
	require.NoError(t, err)

	start := make(chan struct{})
	results := make(chan *httptest.ResponseRecorder, workers)

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
			results <- rec
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	var created int
	var conflicts int
	for rec := range results {
		switch rec.Code {
		case http.StatusCreated:
			created++
			require.NotEmpty(t, rec.Header().Get("Location"))
			var body map[string]any
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			require.Equal(t, shortName, body["short_name"])
			require.NotEmpty(t, body["short_url"])
		case http.StatusUnprocessableEntity:
			conflicts++
			errs := requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
			require.Equal(t, "short name already in use", errs["short_name"])
		default:
			t.Fatalf("unexpected status: %d body=%s", rec.Code, rec.Body.String())
		}
	}

	require.Equal(t, 1, created)
	require.Equal(t, workers-1, conflicts)
}

func TestAPI_Redirect(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/a", "good")
	short := "good"

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

func TestAPI_Redirect_ByShortName_StatusAndLocation(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/redirect", "redir")

	req := httptest.NewRequest(http.MethodGet, "/r/redir", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "https://example.com/redirect", rec.Header().Get("Location"))
}

func TestAPI_Redirect_WritesVisit(t *testing.T) {
	resetLinks(t)

	id := createLink(t, "https://example.com/long-url", "track")

	req := httptest.NewRequest(http.MethodGet, "/r/track", nil)
	req.Header.Set("User-Agent", "curl/8.5.0")
	req.Header.Set("Referer", "https://example.com")
	req.Header.Set("CF-Connecting-IP", "1.2.3.4")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusFound, rec.Code)
	require.Equal(t, "https://example.com/long-url", rec.Header().Get("Location"))

	var (
		linkID    int64
		ip        string
		userAgent string
		referer   string
		status    int
		createdAt time.Time
	)
	err := db.QueryRowContext(
		tcCtx,
		`SELECT link_id, ip, user_agent, referer, status, created_at
		FROM link_visits
		ORDER BY id DESC
		LIMIT 1`,
	).Scan(&linkID, &ip, &userAgent, &referer, &status, &createdAt)
	require.NoError(t, err)
	require.Equal(t, id, linkID)
	require.Equal(t, "1.2.3.4", ip)
	require.Equal(t, "curl/8.5.0", userAgent)
	require.Equal(t, "https://example.com", referer)
	require.Equal(t, http.StatusFound, status)
	require.False(t, createdAt.IsZero())
}

func TestAPI_ListLinkVisits_Range(t *testing.T) {
	resetLinks(t)

	id := createLink(t, "https://example.com/long-url", "visitlist")

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/r/visitlist", nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		req.Header.Set("CF-Connecting-IP", "1.2.3.4")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	rangeParam := url.QueryEscape(`[0,1]`)
	rec := doRequest(t, http.MethodGet, apiLinkVisitsPath+"?range="+rangeParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "link_visits 0-1/3", rec.Header().Get("Content-Range"))

	type visitResp struct {
		ID        int64     `json:"id"`
		LinkID    int64     `json:"link_id"`
		CreatedAt time.Time `json:"created_at"`
		IP        string    `json:"ip"`
		UserAgent string    `json:"user_agent"`
		Referer   string    `json:"referer"`
		Status    int       `json:"status"`
	}

	var visits []visitResp
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &visits))
	require.Len(t, visits, 2)
	require.Equal(t, id, visits[0].LinkID)
	require.Equal(t, "1.2.3.4", visits[0].IP)
	require.Equal(t, "curl/8.5.0", visits[0].UserAgent)
	require.Equal(t, http.StatusFound, visits[0].Status)
	require.False(t, visits[0].CreatedAt.IsZero())
}

func TestAPI_ListLinkVisits_DefaultSort(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/long-url", "visitlist")

	for range 2 {
		req := httptest.NewRequest(http.MethodGet, "/r/visitlist", nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	rangeParam := url.QueryEscape(`[0,1]`)
	rec := doRequest(t, http.MethodGet, apiLinkVisitsPath+"?range="+rangeParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var visits []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &visits))
	require.Len(t, visits, 2)
	// created_at can be equal, so the tie-breaker is id DESC.
	require.Greater(t, asInt64(t, visits[0]["id"]), asInt64(t, visits[1]["id"]))
}

func TestAPI_ListLinkVisits_Range_Header(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/long-url", "visitlist")

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/r/visitlist", nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	rec := doRequestWithHeaders(t, http.MethodGet, apiLinkVisitsPath, nil, map[string]string{
		"Range": "link_visits=0-1",
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "link_visits 0-1/3", rec.Header().Get("Content-Range"))

	var visits []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &visits))
	require.Len(t, visits, 2)
}

func TestAPI_ListLinkVisits_Sort_RangeHeader(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/long-url", "visitlist")

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/r/visitlist", nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	sortParam := sortJSON(t, string(links.SortFieldID), links.SortAsc)
	rec := doRequestWithHeaders(t, http.MethodGet, apiLinkVisitsPath+"?sort="+sortParam, nil, map[string]string{
		"Range": "link_visits=0-1",
	})
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "link_visits 0-1/3", rec.Header().Get("Content-Range"))

	var visits []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &visits))
	require.Len(t, visits, 2)
	require.Less(t, asInt64(t, visits[0]["id"]), asInt64(t, visits[1]["id"]))
}

func TestAPI_ListLinkVisits_Range_WithFilter(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/long-url", "visitlist")

	for range 3 {
		req := httptest.NewRequest(http.MethodGet, "/r/visitlist", nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	rangeParam := url.QueryEscape(`[0,1]`)
	filterParam := url.QueryEscape(`{}`)

	rec := doRequest(t, http.MethodGet, apiLinkVisitsPath+"?range="+rangeParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "link_visits 0-1/3", rec.Header().Get("Content-Range"))

	var base []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &base))

	rec2 := doRequest(t, http.MethodGet, apiLinkVisitsPath+"?range="+rangeParam+"&filter="+filterParam, nil)
	require.Equal(t, http.StatusOK, rec2.Code)
	require.Equal(t, "link_visits 0-1/3", rec2.Header().Get("Content-Range"))

	var withFilter []map[string]any
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &withFilter))

	require.Equal(t, base, withFilter)
}

func TestAPI_ListLinkVisits_SortByLinkIDAsc(t *testing.T) {
	resetLinks(t)

	firstID := createLink(t, "https://example.com/one", "visita")
	secondID := createLink(t, "https://example.com/two", "visitb")

	for _, code := range []string{"visitb", "visita"} {
		req := httptest.NewRequest(http.MethodGet, "/r/"+code, nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	rangeParam := url.QueryEscape(`[0,1]`)
	sortParam := sortJSON(t, string(links.SortFieldLinkID), links.SortAsc)
	rec := doRequest(t, http.MethodGet, apiLinkVisitsPath+"?range="+rangeParam+"&sort="+sortParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	type visitResp struct {
		LinkID int64 `json:"link_id"`
	}

	var visits []visitResp
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &visits))
	require.Len(t, visits, 2)
	require.Equal(t, firstID, visits[0].LinkID)
	require.Equal(t, secondID, visits[1].LinkID)
}

func TestAPI_ListLinkVisits_SortByRefferAlias(t *testing.T) {
	resetLinks(t)

	_ = createLink(t, "https://example.com/one", "reffer1")

	for _, ref := range []string{"https://b.example.com", "https://a.example.com"} {
		req := httptest.NewRequest(http.MethodGet, "/r/reffer1", nil)
		req.Header.Set("User-Agent", "curl/8.5.0")
		req.Header.Set("Referer", ref)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusFound, rec.Code)
	}

	rangeParam := url.QueryEscape(`[0,1]`)
	sortParam := sortJSON(t, "reffer", links.SortAsc)
	rec := doRequest(t, http.MethodGet, apiLinkVisitsPath+"?range="+rangeParam+"&sort="+sortParam, nil)
	require.Equal(t, http.StatusOK, rec.Code)

	var visits []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &visits))
	require.Len(t, visits, 2)
	require.Equal(t, "https://a.example.com", asString(t, visits[0]["reffer"]))
	require.Equal(t, "https://b.example.com", asString(t, visits[1]["reffer"]))
}

func TestAPI_Conflict_Create(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/1", "dupe")

	rec := doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/2",
		"short_name":   "dupe",
	})

	errs := requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Equal(t, "short name already in use", errs["short_name"])
}

func TestAPI_Update_ShortNameAllowed(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/a", "aaaa")

	created := getLinkByShortName(t, "aaaa")
	id := asInt64(t, created["id"])
	updated := doJSON(t, http.MethodPut, "/api/links/"+itoa(id), map[string]any{
		"original_url": "https://example.com/a2",
		"short_name":   "bbbb",
	}, http.StatusOK)

	require.Equal(t, "bbbb", asString(t, updated["short_name"]))
	require.Equal(t, "http://localhost:8080/r/bbbb", asString(t, updated["short_url"]))
	require.Equal(t, "https://example.com/a2", asString(t, updated["original_url"]))
}

func TestAPI_Update_ShortNameMissing_Generates(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/a", "aaaa")

	created := getLinkByShortName(t, "aaaa")
	id := asInt64(t, created["id"])

	updated := doJSON(t, http.MethodPut, "/api/links/"+itoa(id), map[string]any{
		"original_url": "https://example.com/a2",
	}, http.StatusOK)

	newShort := asString(t, updated["short_name"])
	require.NotEqual(t, "aaaa", newShort)
	require.True(t, shortNameRe.MatchString(newShort))
	require.Equal(t, "https://example.com/a2", asString(t, updated["original_url"]))
}

func TestAPI_Update_ShortNameEmpty_Generates(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/a", "aaaa")

	created := getLinkByShortName(t, "aaaa")
	id := asInt64(t, created["id"])

	updated := doJSON(t, http.MethodPut, "/api/links/"+itoa(id), map[string]any{
		"original_url": "https://example.com/a2",
		"short_name":   "",
	}, http.StatusOK)

	newShort := asString(t, updated["short_name"])
	require.NotEqual(t, "aaaa", newShort)
	require.True(t, shortNameRe.MatchString(newShort))
	require.Equal(t, "https://example.com/a2", asString(t, updated["original_url"]))
}

func TestAPI_Update_ShortNameConflict(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/a", "aaaa")

	createLink(t, "https://example.com/b", "bbbb")

	first := getLinkByShortName(t, "aaaa")
	second := getLinkByShortName(t, "bbbb")
	sid := asInt64(t, second["id"])
	target := asString(t, first["short_name"])

	rec := doRequest(t, http.MethodPut, "/api/links/"+itoa(sid), map[string]any{
		"original_url": "https://example.com/b2",
		"short_name":   target,
	})

	errs := requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Equal(t, "short name already in use", errs["short_name"])
}

func TestAPI_Validation(t *testing.T) {
	resetLinks(t)

	// invalid short_name: too short
	rec := doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "ab",
	})
	errs := requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Contains(t, errs["short_name"], "failed on the 'min' tag")

	// invalid short_name: forbidden chars
	rec = doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "ab_cd",
	})
	errs = requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Equal(t, "invalid short_name", errs["short_name"])

	// invalid short_name: too long
	rec = doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   strings.Repeat("a", 33),
	})
	errs = requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Contains(t, errs["short_name"], "failed on the 'max' tag")

	rec = doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com",
		"short_name":   "ab-cd",
	})
	errs = requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Equal(t, "invalid short_name", errs["short_name"])

	// invalid url
	rec = doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "not-a-url",
		"short_name":   "good",
	})
	errs = requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Equal(t, "invalid url", errs["original_url"])
}

func TestAPI_Create_MissingOriginalURL_Required(t *testing.T) {
	resetLinks(t)

	rec := doRequest(t, http.MethodPost, "/api/links", map[string]any{
		"short_name": "good",
	})

	errs := requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Contains(t, errs["original_url"], "failed on the 'required' tag")
}

func TestAPI_Create_AutoShortName(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/auto", "")

	created := getSingleLink(t)
	short := asString(t, created["short_name"])
	require.True(t, shortNameRe.MatchString(short), "generated short_name must be alnum 3..32, got %q", short)

	req := httptest.NewRequest(http.MethodGet, "/r/"+short, nil)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req)

	require.Equal(t, http.StatusFound, rec2.Code)
	require.Equal(t, "https://example.com/auto", rec2.Header().Get("Location"))
}

func TestAPI_Create_ShortNameSpacesOnly_Generates(t *testing.T) {
	resetLinks(t)

	created := doJSON(t, http.MethodPost, "/api/links", map[string]any{
		"original_url": "https://example.com/space",
		"short_name":   "   ",
	}, http.StatusCreated)

	short := asString(t, created["short_name"])
	require.True(t, shortNameRe.MatchString(short))
}

func TestAPI_Redirect_InvalidShortName_400(t *testing.T) {
	resetLinks(t)

	req := httptest.NewRequest(http.MethodGet, "/r/ab_cd", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	errs := requireValidationErrors(t, rec, http.StatusUnprocessableEntity)
	require.Equal(t, "invalid short_name", errs["short_name"])
}

func TestAPI_Update_ShortNameSpacesOnly_Generates(t *testing.T) {
	resetLinks(t)

	createLink(t, "https://example.com/a", "aaaa")

	created := getLinkByShortName(t, "aaaa")
	id := asInt64(t, created["id"])

	updated := doJSON(t, http.MethodPut, "/api/links/"+itoa(id), map[string]any{
		"original_url": "https://example.com/a2",
		"short_name":   "   ",
	}, http.StatusOK)

	newShort := asString(t, updated["short_name"])
	require.NotEqual(t, "aaaa", newShort)
	require.True(t, shortNameRe.MatchString(newShort))
	require.Equal(t, "https://example.com/a2", asString(t, updated["original_url"]))
}

func TestAPI_Redirect_NotFound_404(t *testing.T) {
	resetLinks(t)

	req := httptest.NewRequest(http.MethodGet, "/r/Unknown1", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	p := requireProblem(t, rec, http.StatusNotFound, "about:blank")
	require.Equal(t, "Not Found", p.Title)
	require.Equal(t, "not found", p.Detail)
}

func resetLinks(t *testing.T) {
	t.Helper()

	truncateLinks(t)
	t.Cleanup(func() { truncateLinks(t) })
}

func seedLinks(t *testing.T, count int) {
	t.Helper()

	for i := 0; i < count; i++ {
		shortName := fmt.Sprintf("lnk%03d", i)
		originalURL := fmt.Sprintf("https://example.com/%d", i)
		createLink(t, originalURL, shortName)
	}
}

func truncateLinks(t *testing.T) {
	t.Helper()

	_, err := db.ExecContext(tcCtx, `TRUNCATE link_visits, links RESTART IDENTITY`)
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
