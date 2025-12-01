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

var defaultLogger *Logger

// Logger encapsulates the logging state
type Logger struct {
	slogger    *slog.Logger
	otelLogger log.Logger
	isInit     bool
}

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

	defaultLogger = &Logger{
		slogger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})),
		isInit:  true,
	}
}

// SetOTelLoggerProvider configures the OpenTelemetry logger provider
// This should be called after InitTelemetry() to enable log export
func SetOTelLoggerProvider(provider *sdklog.LoggerProvider) {
	if provider != nil && defaultLogger != nil {
		defaultLogger.otelLogger = provider.Logger("txlog-server")
	}
}

// Error logs an error message using the logger instance.
func Error(msg string) {
	ErrorCtx(context.Background(), msg)
}

// ErrorCtx logs an error message with context.
func ErrorCtx(ctx context.Context, msg string) {
	if defaultLogger == nil || !defaultLogger.isInit {
		return
	}
	defaultLogger.slogger.ErrorContext(ctx, msg)
	defaultLogger.emitOTelLog(ctx, log.SeverityError, msg)
}

// Info logs a message at the info level.
func Info(msg string) {
	InfoCtx(context.Background(), msg)
}

// InfoCtx logs a message at the info level with context.
func InfoCtx(ctx context.Context, msg string) {
	if defaultLogger == nil || !defaultLogger.isInit {
		return
	}
	defaultLogger.slogger.InfoContext(ctx, msg)
	defaultLogger.emitOTelLog(ctx, log.SeverityInfo, msg)
}

// Debug logs a debug-level message.
func Debug(msg string) {
	DebugCtx(context.Background(), msg)
}

// DebugCtx logs a debug-level message with context.
func DebugCtx(ctx context.Context, msg string) {
	if defaultLogger == nil || !defaultLogger.isInit {
		return
	}
	defaultLogger.slogger.DebugContext(ctx, msg)
	defaultLogger.emitOTelLog(ctx, log.SeverityDebug, msg)
}

// Warn logs a message at the WARN level.
func Warn(msg string) {
	WarnCtx(context.Background(), msg)
}

// WarnCtx logs a message at the WARN level with context.
func WarnCtx(ctx context.Context, msg string) {
	if defaultLogger == nil || !defaultLogger.isInit {
		return
	}
	defaultLogger.slogger.WarnContext(ctx, msg)
	defaultLogger.emitOTelLog(ctx, log.SeverityWarn, msg)
}

// emitOTelLog sends a log record to OpenTelemetry if configured
// It automatically includes trace context if available
func (l *Logger) emitOTelLog(ctx context.Context, severity log.Severity, message string) {
	if l.otelLogger == nil {
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

	l.otelLogger.Emit(ctx, record)
}
