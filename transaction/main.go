package transaction

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Transaction struct {
	TransactionID   string            `json:"transaction_id"`
	MachineID       string            `json:"machine_id"`
	Hostname        string            `json:"hostname"`
	BeginTime       *time.Time        `json:"begin_time"`
	EndTime         *time.Time        `json:"end_time"`
	Actions         string            `json:"actions"`
	Altered         string            `json:"altered"`
	User            string            `json:"user"`
	ReturnCode      string            `json:"return_code"`
	ReleaseVersion  string            `json:"release_version"`
	CommandLine     string            `json:"command_line"`
	Comment         string            `json:"comment"`
	ScriptletOutput string            `json:"scriptlet_output"`
	Items           []TransactionItem `json:"items"`
}

type TransactionItem struct {
	Action   string `json:"action"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	Release  string `json:"release"`
	Epoch    string `json:"epoch"`
	Arch     string `json:"arch"`
	Repo     string `json:"repo"`
	FromRepo string `json:"from_repo,omitempty"`
}

// GetTransactionIDs Get the saved transactions IDs for a host
// @Summary		Get saved transactions IDs for a host
// @Description	Get saved transactions IDs for a host
// @Tags			transaction
// @Accept			json
// @Produce		json
// @Success		200	{object}	interface{}
// @Router			/v1/transaction_id [get]
func GetTransactionIDs(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := Transaction{}
		data, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid transaction data")
			return
		}
		err = json.Unmarshal(data, &body)
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid JSON input")
			fmt.Println("Invalid JSON input:", err)
			return
		}

		rows, err := database.Query(`
      SELECT transaction_id
      FROM public.transactions
      WHERE machine_id = $1
      AND hostname = $2
      ORDER BY transaction_id ASC`,
			body.MachineID,
			body.Hostname,
		)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(400, "Couldn't get saved transactions for this host.")
		} else {
			var transactions []int
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					fmt.Println(err)
					c.AbortWithStatusJSON(500, "Error scanning transactions")
					return
				}
				transactions = append(transactions, id)
			}
			defer rows.Close()

			if transactions == nil {
				c.JSON(http.StatusOK, []int{})
				return
			}
			c.JSON(http.StatusOK, transactions)
		}
	}
}

// PostTransaction Create a new transaction
// @Summary		Create a new transaction
// @Description	Create a new transaction
// @Tags			transaction
// @Accept			json
// @Produce		json
// @Param			Transaction	body		Transaction	true	"Transaction data"
// @Success		200			{string}	string		"Transaction created"
// @Failure		400			{string}	string		"Invalid transaction data"
// @Failure		400			{string}	string		"Invalid JSON input"
// @Failure		500			{string}	string		"Database error"
// @Router			/v1/transaction [post]
func PostTransaction(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := Transaction{}
		data, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid transaction data")
			return
		}
		err = json.Unmarshal(data, &body)
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid JSON input")
			fmt.Println("Invalid JSON input:", err)
			return
		}

		// Convert *time.Time to sql.NullTime
		var beginTime sql.NullTime
		if body.BeginTime != nil {
			beginTime.Time = *body.BeginTime
			beginTime.Valid = true
		}

		var endTime sql.NullTime
		if body.EndTime != nil {
			endTime.Time = *body.EndTime
			endTime.Valid = true
		}

		// Start database transaction
		tx, err := database.Begin()
		if err != nil {
			fmt.Println("Error beginning transaction:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "Database error"})
			return
		}

		// Insert the rpm transaction
		_, err = tx.Exec(`
      INSERT INTO transactions (
        transaction_id, machine_id, hostname, begin_time, end_time, actions, altered, "user",
        return_code, release_version, command_line, comment, scriptlet_output
      ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
      )`,
			body.TransactionID,
			body.MachineID,
			body.Hostname,
			beginTime,
			endTime,
			body.Actions,
			body.Altered,
			body.User,
			body.ReturnCode,
			body.ReleaseVersion,
			body.CommandLine,
			body.Comment,
			body.ScriptletOutput)

		if err != nil {
			tx.Rollback()
			fmt.Println("Error inserting transaction:", err)
			c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
			return
		}

		// Insert rpm transaction items
		for _, item := range body.Items {
			_, err = tx.Exec(`
      INSERT INTO transaction_items (
        transaction_id, machine_id, action, package, version, release, epoch, arch, repo, from_repo
      ) VALUES (
        $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
      )`,
				body.TransactionID,
				body.MachineID,
				item.Action,
				item.Name,
				item.Version,
				item.Release,
				item.Epoch,
				item.Arch,
				item.Repo,
				item.FromRepo)

			if err != nil {
				tx.Rollback()
				fmt.Println("Error inserting transaction item:", err)
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
		}

		// Commit the database transaction
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			fmt.Println("Error committing transaction:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transaction created"})
	}
}
