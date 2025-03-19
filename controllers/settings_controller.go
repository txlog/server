package controllers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/util"
)

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
	})
}
