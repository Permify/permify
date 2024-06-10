package oidc

import (
	"log/slog"
)

// SlogAdapter adapts the slog.Logger to be compatible with retryablehttp.LeveledLogger.
type SlogAdapter struct {
	Logger *slog.Logger
}

// Error logs messages at error level.
func (a SlogAdapter) Error(msg string, keysAndValues ...interface{}) {
	a.Logger.Error(msg, keysAndValues...)
}

// Info logs messages at info level.
func (a SlogAdapter) Info(msg string, keysAndValues ...interface{}) {
	a.Logger.Info(msg, keysAndValues...)
}

// Debug logs messages at debug level.
func (a SlogAdapter) Debug(msg string, keysAndValues ...interface{}) {
	a.Logger.Info(msg, keysAndValues...)
}

// Warn logs messages at warn level.
func (a SlogAdapter) Warn(msg string, keysAndValues ...interface{}) {
	a.Logger.Warn(msg, keysAndValues...)
}
