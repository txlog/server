package transactionItem

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type Transaction struct {
	TransactionID   string            `json:"transaction_id"`
	MachineID       string            `json:"machine_id,omitempty"`
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
	FromRepo string `json:"from_repo"`
}

// GetItemIDs Get the saved item IDs for a transaction
//
//	@Summary		Get saved item IDs for a transaction
//	@Description	Get saved item IDs for a transaction
//	@Tags			item
//	@Accept			json
//	@Produce		json
//
//	@Param			machine_id		query		string	false	"Machine ID"
//	@Param			transaction_id	query		string	false	"Transaction ID"
//
//	@Success		200				{object}	interface{}
//	@Router			/v1/item_id [get]
func GetItemIDs(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		transactionID := c.Query("transaction_id")

		if machineID == "" {
			c.AbortWithStatusJSON(400, "machine_id is required")
			return
		}

		if transactionID == "" {
			c.AbortWithStatusJSON(400, "transaction_id is required")
			return
		}

		rows, err := database.Query(`
      SELECT item_id
      FROM public.transaction_items
      WHERE machine_id = $1
      AND transaction_id = $2
      ORDER BY item_id ASC`,
			machineID,
			transactionID,
		)
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(400, "Couldn't get saved item_ids for this transaction.")
		} else {
			var items []int
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					fmt.Println(err)
					c.AbortWithStatusJSON(500, "Error scanning transaction_ids")
					return
				}
				items = append(items, id)
			}
			defer rows.Close()

			if items == nil {
				c.JSON(http.StatusOK, []int{})
				return
			}
			c.JSON(http.StatusOK, items)
		}
	}
}

// GetItems Get the saved items for a transaction
//
//	@Summary		Get saved items for a transaction
//	@Description	Get saved items for a transaction
//	@Tags			item
//	@Accept			json
//	@Produce		json
//	@Param			machine_id		query		string	false	"Machine ID"
//	@Param			transaction_id	query		string	false	"Transaction ID"
//	@Success		200				{object}	interface{}
//	@Router			/v1/item [get]
func GetItems(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		transactionID := c.Query("transaction_id")

		if machineID == "" {
			c.AbortWithStatusJSON(400, "machine_id is required")
			return
		}

		if transactionID == "" {
			c.AbortWithStatusJSON(400, "transaction_id is required")
			return
		}

		var transaction Transaction
		var rows *sql.Rows
		var err error

		// details about the transaction

		rows, err = database.Query(`
      SELECT
        transaction_id,
        hostname,
        begin_time,
        end_time,
        actions,
        altered,
        "user",
        return_code,
        release_version,
        command_line,
        comment,
        scriptlet_output
      FROM public.transactions
      WHERE machine_id = $1
      AND transaction_id = $2`,
			machineID,
			transactionID,
		)

		if err != nil {
			fmt.Println("Error querying transaction:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		if rows.Next() {
			var beginTime sql.NullTime
			var endTime sql.NullTime
			err := rows.Scan(
				&transaction.TransactionID,
				&transaction.Hostname,
				&beginTime,
				&endTime,
				&transaction.Actions,
				&transaction.Altered,
				&transaction.User,
				&transaction.ReturnCode,
				&transaction.ReleaseVersion,
				&transaction.CommandLine,
				&transaction.Comment,
				&transaction.ScriptletOutput,
			)

			if err != nil {
				fmt.Println("Error reading transaction:", err)
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			if beginTime.Valid {
				transaction.BeginTime = &beginTime.Time
			}
			if endTime.Valid {
				transaction.EndTime = &endTime.Time
			}
		} else {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		// details about the items

		rows, err = database.Query(`
    SELECT
      action,
      package,
      version,
      release,
      epoch,
      arch,
      repo,
      from_repo
    FROM public.transaction_items
    WHERE machine_id = $1
    AND transaction_id = $2
    ORDER BY item_id ASC`,
			machineID,
			transactionID,
		)

		if err != nil {
			fmt.Println("Error querying items:", err)
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var transactionItem TransactionItem
			err := rows.Scan(
				&transactionItem.Action,
				&transactionItem.Name,
				&transactionItem.Version,
				&transactionItem.Release,
				&transactionItem.Epoch,
				&transactionItem.Arch,
				&transactionItem.Repo,
				&transactionItem.FromRepo,
			)

			if err != nil {
				fmt.Println("Error reading transaction:", err)
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}

			transaction.Items = append(transaction.Items, transactionItem)
		}

		c.JSON(http.StatusOK, transaction)
	}
}
