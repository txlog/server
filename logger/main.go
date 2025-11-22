package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/trace"
)

var (
	slogger      *slog.Logger
	otelLogger   log.Logger
	loggerIsInit bool
)

// InitLogger initializes a structured logger (slog) with a configurable log
// level. The log level can be set through the LOG_LEVEL environment variable
// with values: "DEBUG", "INFO", "WARN", or "ERROR". If LOG_LEVEL is not set or
// has an invalid value, it defaults to INFO level. The logger outputs to stdout
// using text format.
//
// This logger is enhanced with OpenTelemetry integration. If a LoggerProvider
// is available (configured via telemetry package), logs will also be sent to
// the OpenTelemetry backend with trace context correlation.
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

	slogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	loggerIsInit = true
}

// SetOTelLoggerProvider configures the OpenTelemetry logger provider
// This should be called after InitTelemetry() to enable log export
func SetOTelLoggerProvider(provider *sdklog.LoggerProvider) {
	if provider != nil {
		otelLogger = provider.Logger("txlog-server")
	}
}

// Error logs an error message using the logger instance. It accepts a string
// parameter that contains the message to be logged at the ERROR level.
// If OpenTelemetry is configured, the log will also be sent with trace context.
func Error(msg string) {
	if !loggerIsInit {
		return
	}
	slogger.Error(msg)
	emitOTelLog(context.Background(), log.SeverityError, msg)
}

// Info logs a message at the info level. It serves as a convenience wrapper
// around the underlying logger's Info method.
//
// Parameters:
//   - msg: The message string to be logged
//
// If OpenTelemetry is configured, the log will also be sent with trace context.
func Info(msg string) {
	if !loggerIsInit {
		return
	}
	slogger.Info(msg)
	emitOTelLog(context.Background(), log.SeverityInfo, msg)
}

// Debug logs a debug-level message. It forwards the message to the underlying
// logger implementation. This is useful for detailed debugging information that
// should typically only be enabled during development.
//
// Parameters:
//   - msg: The debug message to be logged
//
// If OpenTelemetry is configured, the log will also be sent with trace context.
func Debug(msg string) {
	if !loggerIsInit {
		return
	}
	slogger.Debug(msg)
	emitOTelLog(context.Background(), log.SeverityDebug, msg)
}

// Warn logs a message at the WARN level. It provides a convenient way to log
// warnings that should be noted but don't necessarily indicate an error
// condition. The message is passed directly to the underlying logger
// implementation.
//
// If OpenTelemetry is configured, the log will also be sent with trace context.
func Warn(msg string) {
	if !loggerIsInit {
		return
	}
	slogger.Warn(msg)
	emitOTelLog(context.Background(), log.SeverityWarn, msg)
}

// emitOTelLog sends a log record to OpenTelemetry if configured
// It automatically includes trace context if available
func emitOTelLog(ctx context.Context, severity log.Severity, message string) {
	if otelLogger == nil {
		return
	}

	var record log.Record
	record.SetSeverity(severity)
	record.SetBody(log.StringValue(message))

	// Add trace context if available
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		record.AddAttributes(
			log.String("trace_id", spanContext.TraceID().String()),
			log.String("span_id", spanContext.SpanID().String()),
		)
	}

	otelLogger.Emit(ctx, record)
}
