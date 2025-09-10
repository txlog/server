package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
	"github.com/txlog/server/util"
)

func GetPackagesIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows *sql.Rows
		var err error

		search := c.Query("search")

		limit := 100
		page := 1

		if pageStr := c.DefaultQuery("page", "1"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}
		offset := (page - 1) * limit

		// First, get total package count
		var total int
		var query string

		if search != "" {
			query = `
        WITH RankedItems AS (
            -- This part identifies the most recent version/release for each package
            SELECT
                package,
                ROW_NUMBER() OVER(PARTITION BY package ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears
            SELECT
                package,
                COUNT(DISTINCT machine_id) as machine_count
            FROM
                public.transaction_items
            GROUP BY
                package
        )
        -- Count the rows that would be returned by the previous query
        SELECT
            COUNT(*) AS returned_records_count
        FROM
            RankedItems ri
        JOIN
            VersionCounts vc ON ri.package = vc.package
        JOIN
            MachineCounts mc ON ri.package = mc.package
        WHERE
            ri.rn = 1
            AND ri.package ILIKE $1;
      `
			err = database.QueryRow(query, util.FormatSearchTerm(search)).Scan(&total)
		} else {
			query = `
        WITH RankedItems AS (
            -- This part identifies the most recent version/release for each package
            SELECT
                package,
                ROW_NUMBER() OVER(PARTITION BY package ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears
            SELECT
                package,
                COUNT(DISTINCT machine_id) as machine_count
            FROM
                public.transaction_items
            GROUP BY
                package
        )
        -- Count the rows that would be returned by the previous query
        SELECT
            COUNT(*) AS returned_records_count
        FROM
            RankedItems ri
        JOIN
            VersionCounts vc ON ri.package = vc.package
        JOIN
            MachineCounts mc ON ri.package = mc.package
        WHERE
            ri.rn = 1;
      `
			err = database.QueryRow(query).Scan(&total)
		}

		if err != nil {
			logger.Error("Error counting packages:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}

		totalPages := (total + limit - 1) / limit

		if search != "" {
			query = `
        WITH RankedItems AS (
            -- This part identifies the most recent version/release for each package
            SELECT
                package,
                version,
                release,
                arch,
                repo,
                ROW_NUMBER() OVER(PARTITION BY package ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears
            SELECT
                package,
                COUNT(DISTINCT machine_id) as machine_count
            FROM
                public.transaction_items
            GROUP BY
                package
        )
        -- The final query joins all results and displays them
        SELECT
            ri.package,
            ri.version,
            ri.release,
            ri.arch,
            ri.repo,
            vc.total_versions - 1 as other_versions_count,
            mc.machine_count
        FROM
            RankedItems ri
        JOIN
            VersionCounts vc ON ri.package = vc.package
        JOIN
            MachineCounts mc ON ri.package = mc.package
        WHERE
            ri.rn = 1
            AND ri.package ILIKE $3
        ORDER BY
            ri.package
        LIMIT $1 OFFSET $2
      `
			rows, err = database.Query(query, limit, offset, util.FormatSearchTerm(search))
		} else {
			query = `
        WITH RankedItems AS (
            -- This part identifies the most recent version/release for each package
            SELECT
                package,
                version,
                release,
                arch,
                repo,
                ROW_NUMBER() OVER(PARTITION BY package ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears
            SELECT
                package,
                COUNT(DISTINCT machine_id) as machine_count
            FROM
                public.transaction_items
            GROUP BY
                package
        )
        -- The final query joins all results and displays them
        SELECT
            ri.package,
            ri.version,
            ri.release,
            ri.arch,
            ri.repo,
            vc.total_versions - 1 as other_versions_count,
            mc.machine_count
        FROM
            RankedItems ri
        JOIN
            VersionCounts vc ON ri.package = vc.package
        JOIN
            MachineCounts mc ON ri.package = mc.package
        WHERE
            ri.rn = 1
        ORDER BY
            ri.package
        LIMIT $1 OFFSET $2
      `
			rows, err = database.Query(query, limit, offset)
		}

		if err != nil {
			logger.Error("Error listing packages:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		packageNames := []models.PackageListing{}
		for rows.Next() {
			var packageName models.PackageListing
			err := rows.Scan(
				&packageName.Package,
				&packageName.Version,
				&packageName.Release,
				&packageName.Arch,
				&packageName.Repo,
				&packageName.TotalVersions,
				&packageName.MachineCount,
			)
			if err != nil {
				logger.Error("Error iterating packages:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			packageNames = append(packageNames, packageName)
		}

		c.HTML(http.StatusOK, "packages.html", gin.H{
			"Context":      c,
			"title":        "Packages",
			"packageNames": packageNames,
			"page":         page,
			"totalPages":   totalPages,
			"totalRecords": total,
			"limit":        limit,
			"offset":       offset,
			"search":       search,
		})
	}
}

func GetPackageByName(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pkg models.Package
		if err := c.ShouldBindUri(&pkg); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		query := `
      SELECT DISTINCT
          package,
          version,
          release,
          arch,
          repo
      FROM
          public.transaction_items
      WHERE
          package = $1
      ORDER BY
          version DESC,
          release DESC;
    `
		rows, err := database.Query(query, pkg.Name)

		if err != nil {
			logger.Error("Error listing packages:" + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": err.Error(),
			})
			return
		}
		defer rows.Close()

		packageNames := []models.PackageListing{}
		for rows.Next() {
			var packageName models.PackageListing
			err := rows.Scan(
				&packageName.Package,
				&packageName.Version,
				&packageName.Release,
				&packageName.Arch,
				&packageName.Repo,
			)
			if err != nil {
				logger.Error("Error iterating packages:" + err.Error())
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
			packageNames = append(packageNames, packageName)
		}

		c.HTML(http.StatusOK, "package_name.html", gin.H{
			"Context":  c,
			"title":    "Packages",
			"packages": packageNames,
			"name":     pkg.Name,
		})
	}
}
