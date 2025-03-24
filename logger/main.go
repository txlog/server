package logger

import (
	"log/slog"
	"os"
	"strings"
)

var logger *slog.Logger

// InitLogger initializes a structured logger (slog) with a configurable log
// level. The log level can be set through the LOG_LEVEL environment variable
// with values: "DEBUG", "INFO", "WARN", or "ERROR". If LOG_LEVEL is not set or
// has an invalid value, it defaults to INFO level. The logger outputs to stdout
// using text format.
func InitLogger() {
	levelMap := map[string]slog.Level{
		"DEBUG": slog.LevelDebug,
		"INFO":  slog.LevelInfo,
		"WARN":  slog.LevelWarn,
		"ERROR": slog.LevelError,
	}

	levelStr := os.Getenv("LOG_LEVEL")
	level, ok := levelMap[strings.ToUpper(levelStr)]
	if !ok {
		level = slog.LevelInfo
	}

	logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}

// Error logs an error message using the logger instance. It accepts a string
// parameter that contains the message to be logged at the ERROR level.
func Error(msg string) {
	logger.Error(msg)
}

// Info logs a message at the info level. It serves as a convenience wrapper
// around the underlying logger's Info method.
//
// Parameters:
//   - msg: The message string to be logged
func Info(msg string) {
	logger.Info(msg)
}

// Debug logs a debug-level message. It forwards the message to the underlying
// logger implementation. This is useful for detailed debugging information that
// should typically only be enabled during development.
//
// Parameters:
//   - msg: The debug message to be logged
func Debug(msg string) {
	logger.Debug(msg)
}

// Warn logs a message at the WARN level. It provides a convenient way to log
// warnings that should be noted but don't necessarily indicate an error
// condition. The message is passed directly to the underlying logger
// implementation.
func Warn(msg string) {
	logger.Warn(msg)
}
