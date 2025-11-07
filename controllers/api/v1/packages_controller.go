package v1

import (
	"database/sql"
	"net/http"

	"github.com/txlog/server/models"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// GetAssetsUsingPackageVersion Get assets that are using a specific package version
// @summary		Get assets that are using a specific package version
// @description	Get assets that are using a specific package version
// @tags			packages
// @produce		json
// @param			name	path		string	true	"Package name"
// @param			version	path		string	true	"Package version"
// @param			release	path		string	true	"Package release"
// @success		200		{object}	models.Package
// @router			/v1/packages/{name}/{version}/{release}/assets [get]
func GetAssetsUsingPackageVersion(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pkg models.Package
		if err := c.ShouldBindUri(&pkg); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		query := `
WITH LatestPackageVersion AS (
  SELECT
    machine_id,
    version,
    release,
    action,
    ROW_NUMBER() OVER (
      PARTITION BY machine_id
      ORDER BY
        version DESC,
        release DESC,
        transaction_id DESC
    ) AS rn
  FROM
    public.transaction_items
  WHERE
    package = $1
)
SELECT
  DISTINCT t.hostname,
  lpv.machine_id
FROM
  LatestPackageVersion AS lpv
  INNER JOIN public.transactions AS t ON lpv.machine_id = t.machine_id
WHERE
  lpv.rn = 1
  AND lpv.version = $2
  AND lpv.release = $3
  AND lpv.action IN ('Install', 'Upgrade', 'Downgrade')
ORDER BY
  t.hostname ASC;`

		rows, err := database.Query(query, pkg.Name, pkg.Version, pkg.Release)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
			return
		}
		defer rows.Close()

		var assets []models.AssetInfo
		for rows.Next() {
			var asset models.AssetInfo
			if err := rows.Scan(&asset.Hostname, &asset.MachineID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
				return
			}
			assets = append(assets, asset)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating rows"})
			return
		}

		if len(assets) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "No assets found using this package version"})
			return
		}

		c.JSON(http.StatusOK, assets)
	}
}

func GetAssetsUsingPackageVersionWeb(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pkg models.Package
		if err := c.ShouldBindUri(&pkg); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		query := `
WITH LatestPackageVersion AS (
  SELECT
    machine_id,
    version,
    release,
    action,
    ROW_NUMBER() OVER (
      PARTITION BY machine_id
      ORDER BY
        version DESC,
        release DESC,
        transaction_id DESC
    ) AS rn
  FROM
    public.transaction_items
  WHERE
    package = $1
)
SELECT
  DISTINCT t.hostname,
  lpv.machine_id
FROM
  LatestPackageVersion AS lpv
  INNER JOIN public.transactions AS t ON lpv.machine_id = t.machine_id
WHERE
  lpv.rn = 1
  AND lpv.version = $2
  AND lpv.release = $3
  AND lpv.action IN ('Install', 'Upgrade', 'Downgrade')
ORDER BY
  t.hostname ASC;`

		rows, err := database.Query(query, pkg.Name, pkg.Version, pkg.Release)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
			return
		}
		defer rows.Close()

		assets := make([]models.AssetInfo, 0)
		for rows.Next() {
			var asset models.AssetInfo
			if err := rows.Scan(&asset.Hostname, &asset.MachineID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
				return
			}
			assets = append(assets, asset)
		}

		if err := rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating rows"})
			return
		}

		c.JSON(http.StatusOK, assets)
	}
}
