package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	logger "github.com/txlog/server/logger"
)

var (
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
)

// InitTelemetry initializes OpenTelemetry SDK with trace and log exporters.
// It configures OTLP exporters based on environment variables:
//
// Environment Variables:
//   - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP endpoint URL (default: none, telemetry disabled)
//   - OTEL_EXPORTER_OTLP_HEADERS: Additional headers for authentication (format: key1=value1,key2=value2)
//   - OTEL_SERVICE_NAME: Service name for telemetry (default: txlog-server)
//   - OTEL_SERVICE_VERSION: Service version (default: unknown)
//   - OTEL_RESOURCE_ATTRIBUTES: Additional resource attributes (format: key1=value1,key2=value2)
//
// If OTEL_EXPORTER_OTLP_ENDPOINT is not set, telemetry is disabled and the function returns nil.
// This allows the application to run normally without OpenTelemetry configuration.
func InitTelemetry() error {
	// Check if OpenTelemetry is configured
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		logger.Info("OpenTelemetry: disabled (OTEL_EXPORTER_OTLP_ENDPOINT not set)")
		return nil
	}

	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(getEnvOrDefault("OTEL_SERVICE_NAME", "txlog-server")),
			semconv.ServiceVersion(getEnvOrDefault("OTEL_SERVICE_VERSION", "unknown")),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithContainer(),
		resource.WithHost(),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize trace exporter
	if err := initTraceProvider(ctx, res); err != nil {
		return fmt.Errorf("failed to initialize trace provider: %w", err)
	}

	// Initialize log exporter
	if err := initLogProvider(ctx, res); err != nil {
		return fmt.Errorf("failed to initialize log provider: %w", err)
	}

	logger.Info("OpenTelemetry: initialized successfully")
	logger.Info(fmt.Sprintf("OpenTelemetry: exporting to %s", endpoint))

	return nil
}

// initTraceProvider initializes the OpenTelemetry trace provider with OTLP HTTP exporter
func initTraceProvider(ctx context.Context, res *resource.Resource) error {
	// Create OTLP trace exporter
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlptracehttp.WithHeaders(parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tracerProvider)

	// Set global propagator to tracecontext (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// initLogProvider initializes the OpenTelemetry log provider with OTLP HTTP exporter
func initLogProvider(ctx context.Context, res *resource.Resource) error {
	// Create OTLP log exporter
	logExporter, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")),
		otlploghttp.WithHeaders(parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))),
	)
	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	// Create log provider
	loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)

	return nil
}

// Shutdown gracefully shuts down the telemetry providers
// It should be called when the application is shutting down to ensure
// all telemetry data is flushed to the backend
func Shutdown() error {
	if tracerProvider == nil && loggerProvider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var errs []error

	if tracerProvider != nil {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown trace provider: %w", err))
		}
	}

	if loggerProvider != nil {
		if err := loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown log provider: %w", err))
		}
	}

	if len(errs) > 0 {
		logger.Error(fmt.Sprintf("OpenTelemetry shutdown errors: %v", errs))
		return fmt.Errorf("shutdown errors: %v", errs)
	}

	logger.Info("OpenTelemetry: shutdown completed")
	return nil
}

// GetLoggerProvider returns the global logger provider for use in the logger package
func GetLoggerProvider() *sdklog.LoggerProvider {
	return loggerProvider
}

// parseHeaders parses the OTEL_EXPORTER_OTLP_HEADERS environment variable
// Expected format: key1=value1,key2=value2
func parseHeaders(headersStr string) map[string]string {
	headers := make(map[string]string)
	if headersStr == "" {
		return headers
	}

	// Simple parsing of comma-separated key=value pairs
	pairs := splitAndTrim(headersStr, ",")
	for _, pair := range pairs {
		kv := splitAndTrim(pair, "=")
		if len(kv) == 2 {
			headers[kv[0]] = kv[1]
		}
	}

	return headers
}

// splitAndTrim splits a string by a delimiter and trims whitespace
func splitAndTrim(s, sep string) []string {
	parts := []string{}
	for _, part := range split(s, sep) {
		trimmed := trim(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// split is a simple string split function
func split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trim removes leading and trailing whitespace
func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// getEnvOrDefault returns the value of an environment variable or a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
