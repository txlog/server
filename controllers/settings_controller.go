package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/util"
)

// GetSettingsIndex handles requests to the settings page by rendering the settings.html template
// with environment variables related to PostgreSQL configuration and cron job settings.
//
// The function populates the template with:
// - PostgreSQL connection details (host, port, user, database name, password [masked], SSL mode)
// - Retention policy settings (days and cron expression)
// - Statistics generation cron expression
//
// Parameters:
//   - ctx: A gin.Context pointer containing the HTTP request context
//
// Returns:
//   - Renders the settings.html template with HTTP 200 OK status
func GetSettingsIndex(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "settings.html", gin.H{
		"Context":                  ctx,
		"title":                    "Server Settings",
		"pgsqlHost":                os.Getenv("PGSQL_HOST"),
		"pgsqlPort":                os.Getenv("PGSQL_PORT"),
		"pgsqlUser":                os.Getenv("PGSQL_USER"),
		"pgsqlDb":                  os.Getenv("PGSQL_DB"),
		"pgsqlPassword":            util.MaskString(os.Getenv("PGSQL_PASSWORD")),
		"pgsqlSslmode":             os.Getenv("PGSQL_SSLMODE"),
		"cronRetentionDays":        os.Getenv("CRON_RETENTION_DAYS"),
		"cronRetentionExpression":  os.Getenv("CRON_RETENTION_EXPRESSION"),
		"cronStatisticsExpression": os.Getenv("CRON_STATS_EXPRESSION"),
		"ignoreEmptyTransaction":   os.Getenv("IGNORE_EMPTY_TRANSACTION"),
	})
}
