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
                REPLACE(package, 'Change ', '') AS package,
                ROW_NUMBER() OVER(PARTITION BY REPLACE(package, 'Change ', '') ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                REPLACE(package, 'Change ', '') AS package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT REPLACE(package, 'Change ', '') AS package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears (active assets only)
            SELECT
                REPLACE(ti.package, 'Change ', '') AS package,
                COUNT(DISTINCT ti.machine_id) as machine_count
            FROM
                public.transaction_items ti
            JOIN
                public.assets a ON ti.machine_id = a.machine_id
            WHERE
                a.is_active = TRUE
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
                REPLACE(package, 'Change ', '') AS package,
                ROW_NUMBER() OVER(PARTITION BY REPLACE(package, 'Change ', '') ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                REPLACE(package, 'Change ', '') AS package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT REPLACE(package, 'Change ', '') AS package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears (active assets only)
            SELECT
                REPLACE(ti.package, 'Change ', '') AS package,
                COUNT(DISTINCT ti.machine_id) as machine_count
            FROM
                public.transaction_items ti
            JOIN
                public.assets a ON ti.machine_id = a.machine_id
            WHERE
                a.is_active = TRUE
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
                REPLACE(package, 'Change ', '') AS package,
                version,
                release,
                arch,
                repo,
                ROW_NUMBER() OVER(PARTITION BY REPLACE(package, 'Change ', '') ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                REPLACE(package, 'Change ', '') AS package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT REPLACE(package, 'Change ', '') AS package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears (active assets only)
            SELECT
                REPLACE(ti.package, 'Change ', '') AS package,
                COUNT(DISTINCT ti.machine_id) as machine_count
            FROM
                public.transaction_items ti
            JOIN
                public.assets a ON ti.machine_id = a.machine_id
            WHERE
                a.is_active = TRUE
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
                REPLACE(package, 'Change ', '') AS package,
                version,
                release,
                arch,
                repo,
                ROW_NUMBER() OVER(PARTITION BY REPLACE(package, 'Change ', '') ORDER BY version DESC, release DESC) as rn
            FROM
                public.transaction_items
        ),
        VersionCounts AS (
            -- This part counts how many unique version/release combinations each package has
            SELECT
                REPLACE(package, 'Change ', '') AS package,
                COUNT(*) as total_versions
            FROM (
                SELECT DISTINCT REPLACE(package, 'Change ', '') AS package, version, release
                FROM public.transaction_items
            ) as distinct_versions
            GROUP BY
                package
        ),
        MachineCounts AS (
            -- This part counts on how many unique machines each package appears (active assets only)
            SELECT
                ti.package,
                COUNT(DISTINCT ti.machine_id) as machine_count
            FROM
                public.transaction_items ti
            JOIN
                public.assets a ON ti.machine_id = a.machine_id
            WHERE
                a.is_active = TRUE
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
      SELECT
        ti.package,
        ti.version,
        ti.release,
        ti.arch,
        ti.repo,
        MAX(t.end_time) AS last_seen
      FROM
        public.transaction_items AS ti
      JOIN
        public.transactions AS t ON ti.transaction_id = t.transaction_id AND ti.machine_id = t.machine_id
      WHERE
        ti.package = $1
      AND ti.repo != '@System'
      GROUP BY
        ti.package,
        ti.version,
        ti.release,
        ti.arch,
        ti.repo
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
				&packageName.LastSeen,
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

		if len(packageNames) == 0 {
			c.HTML(http.StatusNotFound, "404.html", gin.H{
				"title": "Not Found",
			})
			return
		}

		c.HTML(http.StatusOK, "package_name.html", gin.H{
			"Context":  c,
			"title":    "Packages",
			"packages": packageNames,
			"archList": extractUniqueArchitectures(packageNames),
			"repoList": extractUniqueRepositories(packageNames),
			"name":     pkg.Name,
		})
	}
}

// extractUniqueArchitectures takes a slice of PackageListing models and returns
// a slice containing all unique architecture strings found in the input packages.
// Duplicate architectures are automatically filtered out using a map-based approach.
//
// Parameters:
//   - packageNames: slice of models.PackageListing containing package information
//
// Returns:
//   - []string: slice of unique architecture strings (e.g., "amd64", "arm64", "i386")
//
// Example:
//
//	packages := []models.PackageListing{
//	  {Arch: "amd64"}, {Arch: "arm64"}, {Arch: "amd64"},
//	}
//	archs := extractUniqueArchitectures(packages) // Returns: ["amd64", "arm64"]
func extractUniqueArchitectures(packageNames []models.PackageListing) []string {
	// Create a map to track unique architectures
	archMap := make(map[string]bool)
	for _, pkg := range packageNames {
		archMap[pkg.Arch] = true
	}

	// Convert map keys to slice
	var archList []string
	for arch := range archMap {
		archList = append(archList, arch)
	}
	return archList
}

// extractUniqueRepositories takes a slice of PackageListing models and returns
// a slice containing all unique repository strings found in the input packages.
// Duplicate repositories are automatically filtered out using a map-based approach.
//
// Parameters:
//   - packageNames: slice of models.PackageListing containing package information
//
// Returns:
//   - []string: slice of unique repository strings (e.g., "fedora", "updates", "rpmfusion")
//
// Example:
//
//	packages := []models.PackageListing{
//	  {Repo: "fedora"}, {Repo: "updates"}, {Repo: "fedora"},
//	}
//	repos := extractUniqueRepositories(packages) // Returns: ["fedora", "updates"]
func extractUniqueRepositories(packageNames []models.PackageListing) []string {
	// Create a map to track unique repositories
	repoMap := make(map[string]bool)
	for _, pkg := range packageNames {
		repoMap[pkg.Repo] = true
	}

	// Convert map keys to slice
	var repoList []string
	for repo := range repoMap {
		repoList = append(repoList, repo)
	}
	return repoList
}
