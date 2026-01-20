package links

// Logger is a minimal logging interface for the application layer.
type Logger interface {
	With(kv ...any) Logger
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
}

// NopLogger drops all log messages.
type NopLogger struct{}

func (NopLogger) With(...any) Logger   { return NopLogger{} }
func (NopLogger) Info(string, ...any)  {}
func (NopLogger) Warn(string, ...any)  {}
func (NopLogger) Error(string, ...any) {}
