package domain

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrInvalidURL         = errors.New("invalid url")
	ErrInvalidShortName   = errors.New("invalid short name")
	ErrShortNameConflict  = errors.New("short name already exists")
	ErrShortNameImmutable = errors.New("short name is immutable")
)
