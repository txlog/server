package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAnalyticsAnomalies renders the anomaly detection page. The data is fetched
// by the page itself from /v1/reports/anomalies, so no database access here.
func GetAnalyticsAnomalies(c *gin.Context) {
	c.HTML(http.StatusOK, "analytics_anomalies.html", gin.H{
		"Context": c,
		"title":   "Anomaly Detection",
	})
}

// GetAnalyticsSecurity renders the security analysis page. As above, the page
// fetches its own data from the API.
func GetAnalyticsSecurity(c *gin.Context) {
	c.HTML(http.StatusOK, "analytics_security.html", gin.H{
		"Context": c,
		"title":   "Security & Mitigations",
	})
}
