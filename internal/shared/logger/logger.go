package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

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

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Msgf(format, v...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Info().Msgf(format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Msgf(format, v...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Msgf(format, v...)
}

// Err logs an error with error field
func (l *Logger) Err(err error) *zerolog.Event {
	return l.logger.Error().Err(err)
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
	return &Logger{logger: l.logger.With().Str("request_id", requestID).Logger()}
}

// WithUserID adds user ID to the logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{logger: l.logger.With().Str("user_id", userID).Logger()}
}

// GetZerolog returns the underlying zerolog logger
func (l *Logger) GetZerolog() zerolog.Logger {
	return l.logger
}

