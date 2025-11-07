package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

func GetPackagesByWeekIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		graphData, err := getGraphData(database)
		if err != nil {
			logger.Error("Error getting statistics:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "packages_by_week.html", gin.H{
			"Context":   c,
			"title":     "Packages",
			"graphData": graphData,
		})
	}
}

func getGraphData(database *sql.DB) ([]models.PackageProgression, error) {
	rows, err := database.Query(`
    SELECT
      week,
      install,
      upgraded
    FROM (
      SELECT
        DATE_TRUNC('week', t.begin_time)::DATE AS week,
        COUNT(*) FILTER (WHERE ti.action = 'Install') AS install,
        COUNT(*) FILTER (WHERE ti.action = 'Upgraded') AS upgraded
      FROM
        transaction_items AS ti
      JOIN
        transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
      WHERE
        ti.action IN ('Install', 'Upgraded')
      GROUP BY
        week
      ORDER BY
        week DESC
      LIMIT 15
    ) AS recent_weeks
    ORDER BY
      week ASC;`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	progressions := []models.PackageProgression{}

	for rows.Next() {
		var progression = models.PackageProgression{}
		err := rows.Scan(
			&progression.Week,
			&progression.Install,
			&progression.Upgraded,
		)

		if err != nil {
			return nil, err
		}
		progressions = append(progressions, progression)
	}

	return progressions, nil
}
