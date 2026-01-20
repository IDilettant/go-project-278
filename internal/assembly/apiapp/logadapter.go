package apiapp

import (
	"log/slog"

	"code/internal/app/links"
)

type linksSlogLogger struct {
	l *slog.Logger
}

func (l linksSlogLogger) With(kv ...any) links.Logger {
	if l.l == nil {
		return l
	}

	return linksSlogLogger{l: l.l.With(kv...)}
}

func (l linksSlogLogger) Info(msg string, kv ...any) {
	if l.l == nil {
		return
	}

	l.l.Info(msg, kv...)
}

func (l linksSlogLogger) Warn(msg string, kv ...any) {
	if l.l == nil {
		return
	}

	l.l.Warn(msg, kv...)
}

func (l linksSlogLogger) Error(msg string, kv ...any) {
	if l.l == nil {
		return
	}

	l.l.Error(msg, kv...)
}

var _ links.Logger = linksSlogLogger{}
