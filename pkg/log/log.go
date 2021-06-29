package log

import "go.uber.org/zap"

var defaultLogger *Logger

func init() {
	defaultLogger = New(&Config{})
}

// Default returns a default logger instance.
func Default() *Logger {
	return defaultLogger
}

func Init(config *Config) {
	defaultLogger = New(config)
}

// SetDefault sets the default logger instance.
func SetDefault(l *Logger) {
	defaultLogger = l
}

func With(args ...interface{}) *Logger {
	return defaultLogger.With(args...)
}

func WithOptions(opts ...zap.Option) *zap.Logger {
	return defaultLogger.WithOptions(opts...)
}

func WithError(err error) *Logger {
	return defaultLogger.With(zap.Error(err))
}

func WithField(key string, value interface{}) *Logger {
	return defaultLogger.With(zap.Any(key, value))
}

// Debugf prints a debug-level log with format by default logger instance.
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Infof prints a info-level log with format by default logger instance.
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warnf prints a warn-level log with format by default logger instance.
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Errorf prints a error-level log with format by default logger instance.
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatalf prints a fatal-level log with format by default logger instance.
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// Panicf prints a panic-level log with format by default logger instance.
func Panic(format string, args ...interface{}) {
	defaultLogger.Panic(format, args...)
}

// Debugw prints a debug-level log with json-format by default logger instance.
func Debugw(msg string, kvs ...interface{}) {
	defaultLogger.Debugw(msg, kvs...)
}

// Infow prints a info-level log with json-format by default logger instance.
func Infow(msg string, kvs ...interface{}) {
	defaultLogger.Infow(msg, kvs...)
}

// Warnw prints a warn-level log with json-format by default logger instance.
func Warnw(msg string, kvs ...interface{}) {
	defaultLogger.Warnw(msg, kvs...)
}

// Errorw prints a error-level log with json-format by default logger instance.
func Errorw(msg string, kvs ...interface{}) {
	defaultLogger.Errorw(msg, kvs...)
}

// Fatalw prints a fatal-level log with json-format by default logger instance.
func Fatalw(msg string, kvs ...interface{}) {
	defaultLogger.Fatalw(msg, kvs...)
}

// Panicw prints a panicw-level log with json-format by default logger instance.
func Panicw(msg string, kvs ...interface{}) {
	defaultLogger.Panicw(msg, kvs...)
}

// Debugw prints a debug-level log with json-format by default logger instance.
func Debugv(args ...interface{}) {
	defaultLogger.Debugv(args...)
}

// Infow prints a info-level log with json-format by default logger instance.
func Infov(args ...interface{}) {
	defaultLogger.Infov(args...)
}

// Warnw prints a warn-level log with json-format by default logger instance.
func Warnv(args ...interface{}) {
	defaultLogger.Warnv(args...)
}

// Errorw prints a error-level log with json-format by default logger instance.
func Errorv(args ...interface{}) {
	defaultLogger.Errorv(args...)
}

// Fatalw prints a fatal-level log with json-format by default logger instance.
func Fatalv(args ...interface{}) {
	defaultLogger.Fatalv(args...)
}

// Panicw prints a panicw-level log with json-format by default logger instance.
func Panicv(msg string, kvs ...interface{}) {
	defaultLogger.Panic(msg, kvs...)
}
