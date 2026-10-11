package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus for consistent structured logging throughout the application
type Logger struct {
	*logrus.Logger
	name string
}

// New creates a new logger with the given component name
// The name is added as a field to all log entries for traceability
func New(name string) *Logger {
	l := logrus.New()
	l.SetOutput(os.Stdout)
	l.SetLevel(logrus.InfoLevel)

	// Use text formatter for human-readable output
	// Change to &logrus.JSONFormatter{} for production JSON logs
	l.SetFormatter(&logrus.TextFormatter{
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return &Logger{
		Logger: l,
		name:   name,
	}
}

// SetLevel sets the logging level (DEBUG, INFO, WARN, ERROR)
func (l *Logger) SetLevel(level string) {
	switch level {
	case "debug":
		l.Logger.SetLevel(logrus.DebugLevel)
	case "warn":
		l.Logger.SetLevel(logrus.WarnLevel)
	case "error":
		l.Logger.SetLevel(logrus.ErrorLevel)
	default:
		l.Logger.SetLevel(logrus.InfoLevel)
	}
}

// Info logs an info level message with optional key-value pairs
// Usage: lg.Info("User logged in", "user_id", 123, "username", "john")
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.WithField("component", l.name).WithFields(l.parseFields(args)).Info(msg)
}

// Error logs an error level message with optional key-value pairs
// Usage: lg.Error("Failed to process", "error", err.Error(), "request_id", reqID)
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.WithField("component", l.name).WithFields(l.parseFields(args)).Error(msg)
}

// Debug logs a debug level message with optional key-value pairs
// Usage: lg.Debug("Processing request", "request_id", reqID)
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.WithField("component", l.name).WithFields(l.parseFields(args)).Debug(msg)
}

// Warn logs a warning level message with optional key-value pairs
// Usage: lg.Warn("Rate limit approaching", "current", 90, "limit", 100)
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.WithField("component", l.name).WithFields(l.parseFields(args)).Warn(msg)
}

// parseFields converts variadic args to logrus.Fields map
// Expects even number of arguments as key-value pairs
func (l *Logger) parseFields(args []any) logrus.Fields {
	if len(args)%2 != 0 {
		// Odd number of fields - return empty map, don't crash
		return logrus.Fields{}
	}

	fieldMap := make(logrus.Fields, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			// Key is not a string, skip this pair
			continue
		}
		fieldMap[key] = args[i+1]
	}
	return fieldMap
}
