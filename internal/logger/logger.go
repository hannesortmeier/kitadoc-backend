package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

// Logger is the interface for structured logging.
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields logrus.Fields) Logger
}

// logrusLogger implements the Logger interface using logrus.
type logrusLogger struct {
	entry *logrus.Entry
}

// NewLogrusLogger creates a new logrusLogger instance.
func NewLogrusLogger(entry *logrus.Entry) Logger {
	return &logrusLogger{entry: entry}
}

func (l *logrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *logrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *logrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logrusLogger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *logrusLogger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *logrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *logrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func (l *logrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *logrusLogger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

func (l *logrusLogger) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}

func (l *logrusLogger) WithField(key string, value interface{}) Logger {
	return &logrusLogger{entry: l.entry.WithField(key, value)}
}

func (l *logrusLogger) WithFields(fields logrus.Fields) Logger {
	return &logrusLogger{entry: l.entry.WithFields(fields)}
}

var globalLogger Logger

// InitGlobalLogger initializes the global logger instance.
func InitGlobalLogger(level logrus.Level, format logrus.Formatter) {
	logrusLogger := logrus.New()
	logrusLogger.SetLevel(level)
	logrusLogger.SetFormatter(format)
	globalLogger = NewLogrusLogger(logrusLogger.WithFields(logrus.Fields{}))
}

// GetGlobalLogger returns the global logger instance.
func GetGlobalLogger() Logger {
	return globalLogger
}

type loggerContextKey struct{}

// WithLogger returns a new context with the provided logger.
func WithLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey{}, logger)
}

// GetLoggerFromContext retrieves the logger from the context.
// If no logger is found, it returns the global logger.
func GetLoggerFromContext(ctx context.Context) Logger {
	if ctx == nil {
		return globalLogger
	}
	if logger, ok := ctx.Value(loggerContextKey{}).(Logger); ok {
		return logger
	}
	return globalLogger
}
