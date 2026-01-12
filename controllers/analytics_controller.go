package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
)

// GetAnalyticsCompare returns the package comparison page
func GetAnalyticsCompare(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get list of active assets for selection
		assets, err := getActiveAssetsForComparison(database)
		if err != nil {
			logger.Error("Error getting assets for comparison: " + err.Error())
		}

		c.HTML(http.StatusOK, "analytics_compare.html", gin.H{
			"Context": c,
			"title":   "Package Comparison",
			"assets":  assets,
		})
	}
}

// GetAnalyticsFreshness returns the package freshness page
func GetAnalyticsFreshness(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "analytics_freshness.html", gin.H{
			"Context": c,
			"title":   "Package Freshness",
		})
	}
}

// GetAnalyticsAdoption returns the package adoption page
func GetAnalyticsAdoption(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "analytics_adoption.html", gin.H{
			"Context": c,
			"title":   "Package Adoption",
		})
	}
}

// GetAnalyticsAnomalies returns the anomaly detection page
func GetAnalyticsAnomalies(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "analytics_anomalies.html", gin.H{
			"Context": c,
			"title":   "Anomaly Detection",
		})
	}
}

// AssetForComparison represents an asset available for comparison
type AssetForComparison struct {
	MachineID string
	Hostname  string
}

func getActiveAssetsForComparison(database *sql.DB) ([]AssetForComparison, error) {
	query := `
		SELECT machine_id, hostname
		FROM assets
		WHERE is_active = TRUE
		ORDER BY hostname
		LIMIT 100
	`

	rows, err := database.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]AssetForComparison, 0)
	for rows.Next() {
		var asset AssetForComparison
		if err := rows.Scan(&asset.MachineID, &asset.Hostname); err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	return assets, nil
}
