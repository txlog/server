package telemetry

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	"github.com/txlog/server/util"
)

var defaultManager *TelemetryManager

// TelemetryManager encapsulates OpenTelemetry providers
type TelemetryManager struct {
	tracerProvider *sdktrace.TracerProvider
	loggerProvider *sdklog.LoggerProvider
}

// InitTelemetry initializes OpenTelemetry SDK with trace and log exporters.
// It configures OTLP exporters based on environment variables:
//
// Environment Variables:
//   - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP endpoint URL (default: none, telemetry disabled)
//   - OTEL_EXPORTER_OTLP_HEADERS: Additional headers for authentication (format: key1=value1,key2=value2)
//   - OTEL_SERVICE_NAME: Service name for telemetry (default: txlog-server)
//   - OTEL_SERVICE_VERSION: Service version (default: unknown)
//   - OTEL_RESOURCE_ATTRIBUTES: Additional resource attributes (format: key1=value1,key2=value2)
//   - OTEL_EXPORTER_OTLP_INSECURE: Set to "true" to use HTTP instead of HTTPS (default: false)
//   - OTEL_LOGS_EXPORTER: Set to "none" to disable log export (default: otlp)
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

	defaultManager = &TelemetryManager{}
	ctx := context.Background()

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(util.GetEnvOrDefault("OTEL_SERVICE_NAME", "txlog-server")),
			semconv.ServiceVersion(util.GetEnvOrDefault("OTEL_SERVICE_VERSION", "unknown")),
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
	if err := defaultManager.initTraceProvider(ctx, res); err != nil {
		return fmt.Errorf("failed to initialize trace provider: %w", err)
	}

	// Initialize log exporter
	if os.Getenv("OTEL_LOGS_EXPORTER") != "none" {
		if err := defaultManager.initLogProvider(ctx, res); err != nil {
			return fmt.Errorf("failed to initialize log provider: %w", err)
		}
	}

	logger.Info("OpenTelemetry: initialized successfully")
	logger.Info(fmt.Sprintf("OpenTelemetry: exporting to %s", endpoint))

	return nil
}

// initTraceProvider initializes the OpenTelemetry trace provider with OTLP HTTP exporter
func (tm *TelemetryManager) initTraceProvider(ctx context.Context, res *resource.Resource) error {
	endpoint, insecure := parseEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithHeaders(parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))),
	}

	if insecure || isInsecure() {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	// Create OTLP trace exporter
	traceExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	tm.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global trace provider
	otel.SetTracerProvider(tm.tracerProvider)

	// Set global propagator to tracecontext (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

// initLogProvider initializes the OpenTelemetry log provider with OTLP HTTP exporter
func (tm *TelemetryManager) initLogProvider(ctx context.Context, res *resource.Resource) error {
	endpoint, insecure := parseEndpoint(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithHeaders(parseHeaders(os.Getenv("OTEL_EXPORTER_OTLP_HEADERS"))),
	}

	if insecure || isInsecure() {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	// Create OTLP log exporter
	logExporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create log exporter: %w", err)
	}

	// Create log provider
	tm.loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)

	return nil
}

// Shutdown gracefully shuts down the telemetry providers
// It should be called when the application is shutting down to ensure
// all telemetry data is flushed to the backend
func Shutdown() error {
	if defaultManager == nil {
		return nil
	}
	return defaultManager.Shutdown()
}

// Shutdown gracefully shuts down the telemetry manager
func (tm *TelemetryManager) Shutdown() error {
	if tm.tracerProvider == nil && tm.loggerProvider == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var errs []error

	if tm.tracerProvider != nil {
		if err := tm.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("failed to shutdown trace provider: %w", err))
		}
	}

	if tm.loggerProvider != nil {
		if err := tm.loggerProvider.Shutdown(ctx); err != nil {
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
	if defaultManager == nil {
		return nil
	}
	return defaultManager.loggerProvider
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
	for _, part := range strings.Split(s, sep) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// isInsecure checks if the OTEL_EXPORTER_OTLP_INSECURE environment variable is set to "true"
func isInsecure() bool {
	val := os.Getenv("OTEL_EXPORTER_OTLP_INSECURE")
	b, _ := strconv.ParseBool(val)
	return b
}

// parseEndpoint parses the endpoint URL and returns the host:port and whether it is insecure (HTTP)
func parseEndpoint(endpoint string) (string, bool) {
	if strings.HasPrefix(endpoint, "http://") {
		return endpoint[7:], true
	}
	if strings.HasPrefix(endpoint, "https://") {
		return endpoint[8:], false
	}
	return endpoint, false
}
