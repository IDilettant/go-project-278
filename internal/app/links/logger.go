package links

// Logger is a minimal logging interface for the application layer.
type Logger interface {
	Warn(msg string, kv ...any)
}

// NopLogger drops all log messages.
type NopLogger struct{}

func (NopLogger) Warn(string, ...any) {}
