package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	// PORT is reserved for platform/Caddy. HTTP_ADDR controls the Go backend.
	defaultHTTPAddr = ":8080"

	// Sentry
	defaultSentryFlushTimeout      = 2 * time.Second
	defaultSentryMiddlewareTimeout = 2 * time.Second

	// DB pool
	defaultDBMaxOpenConns    = 10
	defaultDBMaxIdleConns    = 10
	defaultDBConnMaxLifetime = 30 * time.Minute

	// HTTP server timeouts
	defaultHTTPReadHeaderTimeout = 5 * time.Second
	defaultHTTPReadTimeout       = 15 * time.Second
	defaultHTTPWriteTimeout      = 15 * time.Second
	defaultHTTPIdleTimeout       = 60 * time.Second
	defaultHTTPShutdownTimeout   = 5 * time.Second

	// Request budget
	defaultRequestBudget = 2 * time.Second
)

type Config struct {
	// HTTPAddr is the Go backend listen address; PORT is reserved for platform/Caddy.
	HTTPAddr string
	BaseURL  string

	DatabaseURL string
	SentryDSN   string

	SentryFlushTimeout      time.Duration
	SentryMiddlewareTimeout time.Duration

	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration

	HTTPReadHeaderTimeout time.Duration
	HTTPReadTimeout       time.Duration
	HTTPWriteTimeout      time.Duration
	HTTPIdleTimeout       time.Duration
	HTTPShutdownTimeout   time.Duration
	RequestBudget         time.Duration

	CORSAllowedOrigins []string
}

type durationSpec struct {
	key string
	def time.Duration
	dst *time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr: normalizeListenAddr(getEnv("HTTP_ADDR", defaultHTTPAddr)),
	}

	baseURL, err := mustEnv("BASE_URL", ErrBaseURLEmpty)
	if err != nil {
		return Config{}, err
	}

	cfg.BaseURL = normalizeBaseURL(baseURL)
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return Config{}, err
	}

	dbURL, err := mustEnv("DATABASE_URL", ErrDatabaseURLEmpty)
	if err != nil {
		return Config{}, err
	}

	cfg.DatabaseURL = dbURL

	sentryDSN, err := mustEnv("SENTRY_DSN", ErrSentryDSNEmpty)
	if err != nil {
		return Config{}, err
	}

	cfg.SentryDSN = sentryDSN

	if err := loadSentry(&cfg); err != nil {
		return Config{}, err
	}

	if err := loadDBPool(&cfg); err != nil {
		return Config{}, err
	}

	if err := loadHTTPServer(&cfg); err != nil {
		return Config{}, err
	}

	if err := loadRequestBudget(&cfg); err != nil {
		return Config{}, err
	}

	loadCORS(&cfg)

	return cfg, nil
}

func normalizeListenAddr(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return defaultHTTPAddr
	}

	if strings.Contains(v, ":") {
		return v
	}

	return ":" + v
}

func normalizeBaseURL(v string) string {
	return strings.TrimRight(strings.TrimSpace(v), "/")
}

func validateBaseURL(v string) error {
	if v == "" {
		return ErrBaseURLEmpty
	}

	u, err := url.Parse(v)
	if err != nil {
		return ErrInvalidBaseURL
	}

	return validateBaseURLParts(u)
}

func validateBaseURLParts(u *url.URL) error {
	if !isAllowedScheme(u.Scheme) {
		return ErrInvalidBaseURL
	}

	if u.Hostname() == "" {
		return ErrInvalidBaseURL
	}

	if u.RawQuery != "" || u.Fragment != "" {
		return ErrInvalidBaseURL
	}

	if u.Path != "" && u.Path != "/" {
		return ErrInvalidBaseURL
	}

	return nil
}

func isAllowedScheme(scheme string) bool {
	return scheme == "http" || scheme == "https"
}

// env helpers

func env(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func getEnv(key, def string) string {
	v := env(key)
	if v == "" {
		return def
	}

	return v
}

func mustEnv(key string, errEmpty error) (string, error) {
	v := env(key)
	if v == "" {
		return "", errEmpty
	}

	return v, nil
}

// typed parsers

func parseDurationEnv(key string, def time.Duration) (time.Duration, error) {
	raw := env(key)
	if raw == "" {
		return def, nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s=%q", ErrInvalidDuration, key, raw)
	}

	return d, nil
}

func parseIntEnv(key string, def int) (int, error) {
	raw := env(key)
	if raw == "" {
		return def, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%w: %s=%q", ErrInvalidInt, key, raw)
	}

	return n, nil
}

// grouped loaders

func loadSentry(cfg *Config) error {
	flush, err := parseDurationEnv("SENTRY_FLUSH_TIMEOUT", defaultSentryFlushTimeout)
	if err != nil {
		return err
	}

	mw, err := parseDurationEnv("SENTRY_MIDDLEWARE_TIMEOUT", defaultSentryMiddlewareTimeout)
	if err != nil {
		return err
	}

	if flush <= 0 || mw <= 0 {
		return fmt.Errorf("%w: sentry flush=%s middleware=%s", ErrInvalidDuration, flush, mw)
	}

	cfg.SentryFlushTimeout = flush
	cfg.SentryMiddlewareTimeout = mw

	return nil
}

func loadDBPool(cfg *Config) error {
	maxOpen, err := parseIntEnv("DB_MAX_OPEN_CONNS", defaultDBMaxOpenConns)
	if err != nil {
		return err
	}

	maxIdle, err := parseIntEnv("DB_MAX_IDLE_CONNS", defaultDBMaxIdleConns)
	if err != nil {
		return err
	}

	connLife, err := parseDurationEnv("DB_CONN_MAX_LIFETIME", defaultDBConnMaxLifetime)
	if err != nil {
		return err
	}

	if maxOpen <= 0 || maxIdle <= 0 || connLife <= 0 {
		return fmt.Errorf("%w: open=%d idle=%d lifetime=%s", ErrInvalidDBPool, maxOpen, maxIdle, connLife)
	}

	if maxIdle > maxOpen {
		return fmt.Errorf("%w: idle=%d > open=%d", ErrInvalidDBPool, maxIdle, maxOpen)
	}

	cfg.DBMaxOpenConns = maxOpen
	cfg.DBMaxIdleConns = maxIdle
	cfg.DBConnMaxLifetime = connLife

	return nil
}

func loadHTTPServer(cfg *Config) error {
	specs := []durationSpec{
		{
			key: "HTTP_READ_HEADER_TIMEOUT",
			def: defaultHTTPReadHeaderTimeout,
			dst: &cfg.HTTPReadHeaderTimeout,
		},
		{
			key: "HTTP_READ_TIMEOUT",
			def: defaultHTTPReadTimeout,
			dst: &cfg.HTTPReadTimeout,
		},
		{
			key: "HTTP_WRITE_TIMEOUT",
			def: defaultHTTPWriteTimeout,
			dst: &cfg.HTTPWriteTimeout,
		},
		{
			key: "HTTP_IDLE_TIMEOUT",
			def: defaultHTTPIdleTimeout,
			dst: &cfg.HTTPIdleTimeout,
		},
		{
			key: "HTTP_SHUTDOWN_TIMEOUT",
			def: defaultHTTPShutdownTimeout,
			dst: &cfg.HTTPShutdownTimeout,
		},
	}

	for _, spec := range specs {
		d, err := parseDurationEnv(spec.key, spec.def)
		if err != nil {
			return err
		}

		if d <= 0 {
			return fmt.Errorf("%w: %s=%s", ErrInvalidDuration, spec.key, d)
		}

		*spec.dst = d
	}

	return nil
}

func loadRequestBudget(cfg *Config) error {
	budget, err := parseDurationEnv("REQUEST_BUDGET", defaultRequestBudget)
	if err != nil {
		return err
	}

	if budget <= 0 {
		return fmt.Errorf("%w: REQUEST_BUDGET=%s", ErrInvalidDuration, budget)
	}

	cfg.RequestBudget = budget

	return nil
}

func loadCORS(cfg *Config) {
	raw := env("CORS_ALLOWED_ORIGINS")
	if raw == "" {
		return
	}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		
		cfg.CORSAllowedOrigins = append(cfg.CORSAllowedOrigins, origin)
	}
}
