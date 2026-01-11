package config

import "errors"

var (
	ErrPortEmpty = errors.New("PORT is empty")

	ErrBaseURLEmpty   = errors.New("BASE_URL is empty")
	ErrInvalidBaseURL = errors.New("BASE_URL is invalid")

	ErrDatabaseURLEmpty = errors.New("DATABASE_URL is empty")
	ErrSentryDSNEmpty   = errors.New("SENTRY_DSN is empty")

	ErrInvalidDuration = errors.New("invalid duration env")
	ErrInvalidInt      = errors.New("invalid int env")

	ErrInvalidDBPool = errors.New("invalid db pool config")
)
