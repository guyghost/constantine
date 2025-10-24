package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with convenience methods
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      slog.Level
	Format     string // "json" or "text"
	AddSource  bool
	OutputPath string // empty means stdout
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:     slog.LevelInfo,
		Format:    "json",
		AddSource: false,
	}
}

// New creates a new structured logger
func New(config *Config) *Logger {
	if config == nil {
		config = DefaultConfig()
	}

	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	output := os.Stdout
	if config.OutputPath != "" {
		file, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			output = file
		}
	}

	if config.Format == "text" {
		handler = slog.NewTextHandler(output, opts)
	} else {
		handler = slog.NewJSONHandler(output, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger.With(),
	}
}

// WithFields returns a logger with additional fields
func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// WithField returns a logger with an additional field
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		Logger: l.Logger.With(key, value),
	}
}

// WithError returns a logger with an error field
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
	}
}

// Component returns a logger for a specific component
func (l *Logger) Component(name string) *Logger {
	return &Logger{
		Logger: l.Logger.With("component", name),
	}
}

// Exchange returns a logger for a specific exchange
func (l *Logger) Exchange(name string) *Logger {
	return &Logger{
		Logger: l.Logger.With("exchange", name),
	}
}

// Symbol returns a logger for a specific trading symbol
func (l *Logger) Symbol(symbol string) *Logger {
	return &Logger{
		Logger: l.Logger.With("symbol", symbol),
	}
}

// Trade logs trade-related information
func (l *Logger) Trade(fields map[string]any) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Logger.Info("trade", args...)
}

// Order logs order-related information
func (l *Logger) Order(fields map[string]any) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Logger.Info("order", args...)
}

// Risk logs risk management information
func (l *Logger) Risk(fields map[string]any) {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Logger.Warn("risk_event", args...)
}

// Global logger instance
var defaultLogger *Logger

func init() {
	defaultLogger = New(DefaultConfig())
}

// SetDefault sets the default global logger
func SetDefault(l *Logger) {
	if l != nil {
		defaultLogger = l
	}
}

// Default returns the default global logger
func Default() *Logger {
	return defaultLogger
}

// Convenience functions using default logger

// Debug logs a debug message
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

// WithFields returns a logger with fields
func WithFields(fields map[string]any) *Logger {
	return defaultLogger.WithFields(fields)
}

// WithField returns a logger with a field
func WithField(key string, value any) *Logger {
	return defaultLogger.WithField(key, value)
}

// WithError returns a logger with an error
func WithError(err error) *Logger {
	return defaultLogger.WithError(err)
}

// Component returns a component logger
func Component(name string) *Logger {
	return defaultLogger.Component(name)
}

// Exchange returns an exchange logger
func Exchange(name string) *Logger {
	return defaultLogger.Exchange(name)
}

// Symbol returns a symbol logger
func Symbol(symbol string) *Logger {
	return defaultLogger.Symbol(symbol)
}
