// Package logger provides context-aware and structured logging capabilities.
// This module is (mostly) copied from the Tharsis API.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// Logger is a logger that supports log levels, context and structured logging.
type Logger interface {
	// With returns a logger based off the root logger and decorates it with the given arguments.
	With(args ...interface{}) Logger

	// Debug uses fmt.Sprint to construct and log a message at DEBUG level
	Debug(args ...interface{})
	// Info uses fmt.Sprint to construct and log a message at INFO level
	Info(args ...interface{})

	// Error uses fmt.Sprint to construct and log a message at ERROR level
	Error(args ...interface{})

	// Debugf uses fmt.Sprintf to construct and log a message at DEBUG level
	Debugf(format string, args ...interface{})
	// Infof uses fmt.Sprintf to construct and log a message at INFO level
	Infof(format string, args ...interface{})
	// Errorf uses fmt.Sprintf to construct and log a message at ERROR level
	Errorf(format string, args ...interface{})

	// Debugw logs a message with some additional context
	Debugw(msg string, keysAndValues ...interface{})
	// Infow logs a message with some additional context
	Infow(msg string, keysAndValues ...interface{})
	// Errorw logs a message with some additional context
	Errorw(msg string, keysAndValues ...interface{})
}

type logger struct {
	*zap.SugaredLogger
}

// New creates a new logger using the default configuration.
func New() Logger {
	l, _ := zap.NewProduction()
	return NewWithZap(l)
}

// NewAtLevel creates a new logger at a level specified by the string argument,
// defaulting to info level if the string argument is not a valid level specifier.
func NewAtLevel(wantLevel string) Logger {

	// Set the level.
	atom := zap.NewAtomicLevel()
	parsed, err := zapcore.ParseLevel(wantLevel)
	// Ignore error and leave level set at info.
	if err == nil {
		atom.SetLevel(parsed)
	}

	// Set the encoder configuration.
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = nil  // Remove timestamp.
	encoderCfg.EncodeLevel = nil // Remove error level.
	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	// Finally, create the logger.
	zapLogger := zap.New(zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), atom))
	return NewWithZap(zapLogger)
}

// NewWithZap creates a new logger using the pre-configured zap logger.
func NewWithZap(l *zap.Logger) Logger {
	return &logger{l.Sugar()}
}

// NewForTest returns a new logger and the corresponding observed logs which can be used in unit tests to verify log entries.
func NewForTest() (Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	return NewWithZap(zap.New(core)), recorded
}

// With returns a logger based off the root logger and decorates it with the given arguments.
//
// The arguments should be specified as a sequence of name, value pairs with names being strings.
// The arguments will also be added to every log message generated by the logger.
func (l *logger) With(args ...interface{}) Logger {
	if len(args) > 0 {
		return &logger{l.SugaredLogger.With(args...)}
	}
	return l
}
