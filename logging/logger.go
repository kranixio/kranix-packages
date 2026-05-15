package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides structured logging capabilities.
type Logger struct {
	logger *zap.Logger
	name   string
}

// New creates a new named logger.
func New(name string) *Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		logger: logger.Named(name),
		name:   name,
	}
}

// NewWithLevel creates a new logger with a specific log level.
func NewWithLevel(name string, level zapcore.Level) *Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &Logger{
		logger: logger.Named(name),
		name:   name,
	}
}

// Info logs an informational message with key-value pairs.
func (l *Logger) Info(msg string, fields ...interface{}) {
	l.logger.Sugar().Infow(msg, fields...)
}

// Error logs an error message with key-value pairs.
func (l *Logger) Error(msg string, fields ...interface{}) {
	l.logger.Sugar().Errorw(msg, fields...)
}

// Debug logs a debug message with key-value pairs.
func (l *Logger) Debug(msg string, fields ...interface{}) {
	l.logger.Sugar().Debugw(msg, fields...)
}

// Warn logs a warning message with key-value pairs.
func (l *Logger) Warn(msg string, fields ...interface{}) {
	l.logger.Sugar().Warnw(msg, fields...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, fields ...interface{}) {
	l.logger.Sugar().Fatalw(msg, fields...)
}

// With returns a new logger with additional fields.
func (l *Logger) With(fields ...interface{}) *Logger {
	return &Logger{
		logger: l.logger.Sugar().With(fields...).Desugar(),
		name:   l.name,
	}
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// Named returns a new logger with the specified name.
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		logger: l.logger.Named(name),
		name:   l.name + "." + name,
	}
}
