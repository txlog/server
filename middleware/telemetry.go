package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// TelemetryMiddleware returns a Gin middleware that adds OpenTelemetry tracing
// to HTTP requests. It uses the otelgin instrumentation to automatically create
// spans for each request with details like:
// - HTTP method and route
// - Status code
// - Request duration
// - Client IP
//
// The middleware also propagates trace context to downstream services.
// If OpenTelemetry is not configured, the middleware still works but doesn't
// create traces.
func TelemetryMiddleware() gin.HandlerFunc {
	return otelgin.Middleware("txlog-server")
}
