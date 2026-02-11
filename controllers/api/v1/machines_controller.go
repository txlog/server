package v1

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
)

type MachineID struct {
	Hostname  string     `json:"hostname"`
	MachineID string     `json:"machine_id"`
	BeginTime *time.Time `json:"begin_time,omitempty"`
}

// GetMachines queries the database for unique machines based on hostname,
// filtering by OS and agent version if provided.
//
//	@Summary		List machine IDs
//	@Description	List machine IDs
//	@Tags			assets
//	@Accept			json
//	@Produce		json
//	@Param			os				query		string	false	"Operating System"
//	@Param			agent_version	query		string	false	"Agent Version"
//	@Success		200				{object}	interface{}
//	@Failure		400				{string}	string	"Invalid execution data"
//	@Failure		400				{string}	string	"Invalid JSON input"
//	@Failure		500				{string}	string	"Database error"
//	@Security		ApiKeyAuth
//	@Router			/v1/machines [get]
func GetMachines(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		os := c.Query("os")
		agentVersion := c.Query("agent_version")

		var rows *sql.Rows
		var err error

		query := `
    SELECT
      a.hostname,
      a.machine_id
    FROM assets a
    WHERE a.is_active = TRUE`

		var params []interface{}
		var paramCount int

		if os != "" {
			logger.Debug("os: " + os)
			if os == "Undefined OS" {
				os = ""
			}
			paramCount++
			query += ` AND a.os = $` + strconv.Itoa(paramCount)
			params = append(params, os)
		}

		if agentVersion != "" {
			logger.Debug("agent_version: " + agentVersion)
			if agentVersion == "with undefined version" {
				agentVersion = ""
			}
			paramCount++
			query += ` AND a.agent_version = $` + strconv.Itoa(paramCount)
			params = append(params, agentVersion)
		}

		if len(params) > 0 {
			rows, err = database.Query(query+" ORDER BY hostname", params...)
		} else {
			rows, err = database.Query(query + " ORDER BY hostname")
		}

		if err != nil {
			logger.Error("Error querying assets: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		machines := []MachineID{}
		for rows.Next() {
			var machine MachineID
			err := rows.Scan(
				&machine.Hostname,
				&machine.MachineID,
			)
			if err != nil {
				logger.Error("Error iterating assets: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			machines = append(machines, machine)
		}

		c.JSON(http.StatusOK, machines)
	}
}

// GetAssetsRequiringRestart queries the database for assets that require a restart
//
//	@Summary		List assets requiring restart
//	@Description	This endpoint retrieves assets that have the latest execution data indicating a restart is required.
//	@Tags			assets
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	interface{}
//	@Failure		400	{string}	string	"Invalid execution data"
//	@Failure		400	{string}	string	"Invalid JSON input"
//	@Failure		500	{string}	string	"Database error"
//	@Security		ApiKeyAuth
//	@Router			/v1/assets/requiring-restart [get]
func GetAssetsRequiringRestart(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rows *sql.Rows
		var err error

		query := `
      SELECT
        a.hostname,
        a.machine_id
      FROM assets a
      WHERE a.is_active = TRUE
      AND a.needs_restarting IS TRUE
      ORDER BY a.hostname ASC;`

		rows, err = database.Query(query)

		if err != nil {
			logger.Error("Error querying assets that require a restart: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		machines := []MachineID{}
		for rows.Next() {
			var machine MachineID
			err := rows.Scan(
				&machine.Hostname,
				&machine.MachineID,
			)
			if err != nil {
				logger.Error("Error iterating assets: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			machines = append(machines, machine)
		}

		c.JSON(http.StatusOK, machines)
	}
}

func GetMachineIDs(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostname := c.Query("hostname")

		if hostname == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, "hostname is required")
			return
		}

		var rows *sql.Rows
		var err error

		rows, err = database.Query(`
      SELECT hostname, machine_id, begin_time
      FROM transactions
      WHERE transaction_id IN (
        SELECT MIN(transaction_id)
        FROM transactions
        GROUP BY machine_id
      ) AND hostname = $1
      ORDER BY begin_time DESC`,
			hostname,
		)

		if err != nil {
			logger.Error("Error querying machine_id: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		machines := []MachineID{}
		for rows.Next() {
			var machine MachineID
			var beginTime sql.NullTime
			err := rows.Scan(
				&machine.Hostname,
				&machine.MachineID,
				&beginTime,
			)
			if err != nil {
				logger.Error("Error iterating machine_id: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if beginTime.Valid {
				machine.BeginTime = &beginTime.Time
			}
			machines = append(machines, machine)
		}

		c.JSON(http.StatusOK, machines)
	}
}
