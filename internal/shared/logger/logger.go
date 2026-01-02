// Package logger provides structured logging functionality.
package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// contextKey is a type for context keys to avoid collisions
type contextKey struct{}

var requestIDKey = contextKey{}

// Logger wraps zerolog.Logger
type Logger struct {
	logger zerolog.Logger
}

// New creates a new logger instance
func New(level string, pretty bool) *Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(parseLevel(level))

	var logger zerolog.Logger
	if pretty {
		logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		logger = log.Logger
	}

	return &Logger{logger: logger}
}

// parseLevel parses log level string
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// getLogger returns a logger with request ID from context if available
func (l *Logger) getLogger(ctx context.Context) zerolog.Logger {
	log := l.logger
	if ctx != nil {
		if requestID := GetRequestIDFromContext(ctx); requestID != "" {
			log = log.With().Str("request_id", requestID).Logger()
		}
	}
	return log
}

// Debug logs a debug message (context-aware)
func (l *Logger) Debug(ctx context.Context, msg string) {
	log := l.getLogger(ctx)
	log.Debug().Msg(msg)
}

// Debugf logs a formatted debug message (context-aware)
func (l *Logger) Debugf(ctx context.Context, format string, v ...interface{}) {
	log := l.getLogger(ctx)
	log.Debug().Msgf(format, v...)
}

// Info logs an info message (context-aware)
func (l *Logger) Info(ctx context.Context, msg string) {
	log := l.getLogger(ctx)
	log.Info().Msg(msg)
}

// Infof logs a formatted info message (context-aware)
func (l *Logger) Infof(ctx context.Context, format string, v ...interface{}) {
	log := l.getLogger(ctx)
	log.Info().Msgf(format, v...)
}

// Warn logs a warning message (context-aware)
func (l *Logger) Warn(ctx context.Context, msg string) {
	log := l.getLogger(ctx)
	log.Warn().Msg(msg)
}

// Warnf logs a formatted warning message (context-aware)
func (l *Logger) Warnf(ctx context.Context, format string, v ...interface{}) {
	log := l.getLogger(ctx)
	log.Warn().Msgf(format, v...)
}

// Error logs an error message (context-aware)
func (l *Logger) Error(ctx context.Context, msg string) {
	log := l.getLogger(ctx)
	log.Error().Msg(msg)
}

// Errorf logs a formatted error message (context-aware)
func (l *Logger) Errorf(ctx context.Context, format string, v ...interface{}) {
	log := l.getLogger(ctx)
	log.Error().Msgf(format, v...)
}

// Err logs an error with error field (context-aware)
func (l *Logger) Err(ctx context.Context, err error) *zerolog.Event {
	log := l.getLogger(ctx)
	return log.Error().Err(err)
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{logger: l.logger.With().Interface(key, value).Logger()}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{logger: ctx.Logger()}
}

// WithRequestID adds request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return l.WithField("request_id", requestID)
}

// WithUserID adds user ID to the logger
func (l *Logger) WithUserID(userID string) *Logger {
	return l.WithField("user_id", userID)
}

// GetZerolog returns the underlying zerolog logger
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}

// WithRequestIDContext stores request ID in context for automatic propagation (OTel-like)
func WithRequestIDContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestIDFromContext extracts request ID from context
func GetRequestIDFromContext(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}
