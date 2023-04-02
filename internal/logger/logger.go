package logger

import (
	"context"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// global is a global logger instance.
	global       *zap.SugaredLogger
	defaultLevel = zap.NewAtomicLevelAt(zap.ErrorLevel)
)

func init() { //nolint: gochecknoinits // it must be installed otherwise the application will not have logs
	SetLogger(New(defaultLevel))
}

// New creates a new instance of *zap.SugaredLogger with standard JSON output.
// If the logging level is not passed, the default level (zap.ErrorLevel) will be used.
func New(level zapcore.LevelEnabler, options ...zap.Option) *zap.SugaredLogger {
	return NewWithSink(level, os.Stdout, options...)
}

// NewWithSink creates a new instance of *zap.SugaredLogger with standard JSON output.
// If the logging level is not passed, the default level (zap.ErrorLevel) will be used.
// Sink is used for log output.
func NewWithSink(level zapcore.LevelEnabler, sink io.Writer, options ...zap.Option) *zap.SugaredLogger {
	if level == nil {
		level = defaultLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}),
		zapcore.AddSync(sink),
		level,
	)

	return zap.New(core, options...).Sugar()
}

// Level returns the current logging level of the global logger.
func Level() zapcore.Level {
	return defaultLevel.Level()
}

// SetLevel sets the logging level of the global logger.
func SetLevel(l string) {
	zl := getLogLevel(l)
	defaultLevel.SetLevel(zl)
}

// Logger returns the global logger.
func Logger() *zap.SugaredLogger {
	return global
}

// SetLogger sets the global logger. The function is not thread-safe.
func SetLogger(l *zap.SugaredLogger) {
	global = l
}

// Debug logs the given message at the debug level using the logger from the context.
func Debug(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Debug(args...)
}

// Debugf logs the formatted message at the debug level using the logger from the context.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Debugf(format, args...)
}

// DebugKV logs the message and key-value pairs at the debug level using the logger from the context.
func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Debugw(message, kvs...)
}

// Info logs the given message at the info level using the logger from the context.
func Info(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Info(args...)
}

// Infof logs the formatted message at the info level using the logger from the context.
func Infof(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Infof(format, args...)
}

// InfoKV logs the message and key-value pairs at the info level using the logger from the context.
func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Infow(message, kvs...)
}

// Warn logs the given message at the warn level using the logger from the context.
func Warn(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Warn(args...)
}

// Warnf logs the formatted message at the warn level using the logger from the context.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Warnf(format, args...)
}

// WarnKV logs the message and key-value pairs at the warn level using the logger from the context.
func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Warnw(message, kvs...)
}

// Error logs the given message at the error level using the logger from the context.
func Error(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Error(args...)
}

// Errorf logs the formatted message at the error level using the logger from the context.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Errorf(format, args...)
}

// ErrorKV logs the message and key-value pairs at the error level using the logger from the context.
func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Errorw(message, kvs...)
}

// Fatal logs the given message at the fatal level using the logger from the context
// and then calls os.Exit(1).
func Fatal(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Fatal(args...)
}

// Fatalf logs the formatted message at the fatal level using the logger from the context
// and then calls os.Exit(1).
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Fatalf(format, args...)
}

// FatalKV logs the message and key-value pairs at the fatal level using the logger from the context
// and then calls os.Exit(1).
func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Fatalw(message, kvs...)
}

// Panic logs the given message at the panic level using the logger from the context
// and then calls panic().
func Panic(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Panic(args...)
}

// Panicf logs the formatted message at the panic level using the logger from the context
// and then calls panic().
func Panicf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Panicf(format, args...)
}

// PanicKV logs the message and key-value pairs at the fatal level using the logger from the context
// and then calls panic().
func PanicKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Panicw(message, kvs...)
}

func getLogLevel(l string) zapcore.Level {
	switch strings.ToLower(l) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.ErrorLevel
	}
}
