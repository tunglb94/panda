package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// New creates a production-ready structured JSON logger.
// In development (non-JSON) environments pass pretty=true for human-readable output.
func New(level, serviceName string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	lvl := parseLevel(level)

	return zerolog.New(os.Stdout).
		Level(lvl).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// NewPretty creates a logger with coloured console output for local development.
func NewPretty(level, serviceName string) zerolog.Logger {
	zerolog.TimeFieldFormat = time.Kitchen

	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Kitchen}).
		Level(parseLevel(level)).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()
}

// FromConfig creates a logger based on environment: pretty in development, JSON in production.
func FromConfig(level, serviceName, environment string) zerolog.Logger {
	if environment == "development" {
		return NewPretty(level, serviceName)
	}
	return New(level, serviceName)
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
