package config

import "errors"

var (
	ErrBaseURLEmpty   = errors.New("BASE_URL is empty")
	ErrInvalidBaseURL = errors.New("BASE_URL is invalid")

	ErrDatabaseURLEmpty = errors.New("DATABASE_URL is empty")

	ErrInvalidDuration = errors.New("invalid duration env")
	ErrInvalidInt      = errors.New("invalid int env")

	ErrInvalidDBPool = errors.New("invalid db pool config")
)
