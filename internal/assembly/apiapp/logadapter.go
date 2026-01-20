package apiapp

import (
	"log/slog"

	"code/internal/app/links"
)

type linksSlogLogger struct {
	l *slog.Logger
}

func (l linksSlogLogger) Warn(msg string, kv ...any) {
	if l.l == nil {
		return
	}

	l.l.Warn(msg, kv...)
}

var _ links.Logger = linksSlogLogger{}
