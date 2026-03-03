package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAnalyticsAnomalies returns the anomaly detection page
func GetAnalyticsAnomalies(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "analytics_anomalies.html", gin.H{
			"Context": c,
			"title":   "Anomaly Detection",
		})
	}
}

// GetAnalyticsSecurity returns the security analysis page
func GetAnalyticsSecurity(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "analytics_security.html", gin.H{
			"Context": c,
			"title":   "Security & Mitigations",
		})
	}
}
