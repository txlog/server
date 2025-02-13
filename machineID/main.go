package machineID

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type MachineID struct {
	Hostname  string     `json:"hostname"`
	MachineID string     `json:"machine_id"`
	BeginTime *time.Time `json:"begin_time"`
}

// GetMachineID List the machine_id of the given hostname
//
//	@Summary		List machine IDs
//	@Description	List machine IDs
//	@Tags			machine_id
//	@Accept			json
//	@Produce		json
//	@Param			hostname	query		string	false	"Hostname"
//	@Success		200			{object}	interface{}
//	@Failure		400			{string}	string	"Invalid execution data"
//	@Failure		400			{string}	string	"Invalid JSON input"
//	@Failure		500			{string}	string	"Database error"
//	@Router			/v1/machine_id [get]
func GetMachineID(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		hostname := c.Query("hostname")

		if hostname == "" {
			c.AbortWithStatusJSON(400, "hostname is required")
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
			fmt.Println("Error querying machine_id:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
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
				fmt.Println("Error iterating machine_id:", err)
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
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
