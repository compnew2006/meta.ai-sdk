package metaai

import (
	"log"
	"os"
)

// logger.go defines the minimal logging interface the SDK uses internally for
// best-effort operations whose errors must not abort a user-facing call but
// should still be observable by operators (e.g. markConversationSeen).
//
// Logger is intentionally narrow (Debugf only) to avoid pulling a logging
// dependency into the public API. Callers can adapt any structured logger
// (slog, logrus, zap) to it via a one-method shim.

// Logger is the internal debug-logging interface. The zero value of Client
// uses a no-op logger; supply a real one via WithLogger.
type Logger interface {
	Debugf(format string, args ...any)
}

// stdConsoleLogger prints to stderr.
type stdConsoleLogger struct {
	l *log.Logger
}

func (s stdConsoleLogger) Debugf(format string, args ...any) {
	s.l.Printf(format, args...)
}

// noopLogger discards all output.
type noopLogger struct{}

func (noopLogger) Debugf(format string, args ...any) {}

// NewDefaultLogger returns a logger based on environment variables.
// If META_AI_DEBUG=1 or true, it returns a console logger printing to stderr.
// Otherwise, it returns a no-op logger.
func NewDefaultLogger() Logger {
	if os.Getenv("META_AI_DEBUG") == "1" || os.Getenv("META_AI_DEBUG") == "true" {
		return stdConsoleLogger{l: log.New(os.Stderr, "[metaai] ", log.LstdFlags|log.Lmicroseconds)}
	}
	return noopLogger{}
}
