package v1

import (
	"database/sql"
	"net/http"

	"github.com/txlog/server/models"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
)

// GetItemIDs Get the saved item IDs for a transaction
//
//	@Summary		Get saved item IDs for a transaction
//	@Description	Get saved item IDs for a transaction
//	@Tags			items
//	@Accept			json
//	@Produce		json
//
//	@Param			machine_id		query		string	false	"Machine ID"
//	@Param			transaction_id	query		string	false	"Transaction ID. If not provided, the last transaction will be used."
//
//	@Success		200				{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/v1/items/ids [get]
func GetItemIDs(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		transactionID := c.Query("transaction_id")

		if machineID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, "machine_id is required")
			return
		}

		if transactionID == "" {
			// If no transaction_id provided, get the latest one
			row := database.QueryRow(`
        SELECT transaction_id
        FROM public.transaction_items
        WHERE machine_id = $1
        ORDER BY transaction_id DESC
        LIMIT 1`, machineID)
			if err := row.Scan(&transactionID); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, "No transactions found for this machine")
				return
			}
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
			logger.Error("Couldn't get saved item_ids for this transaction: " + err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, "Couldn't get saved item_ids for this transaction.")
			return
		}
		defer rows.Close()

		var items []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				logger.Error("Error scanning transaction_ids: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, "Error scanning transaction_ids")
				return
			}
			items = append(items, id)
		}

		if items == nil {
			c.JSON(http.StatusOK, []int{})
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

// GetItems Get the saved items for a transaction
//
//	@Summary		Get saved items for a transaction
//	@Description	Get saved items for a transaction
//	@Tags			items
//	@Accept			json
//	@Produce		json
//	@Param			machine_id		query		string	false	"Machine ID"
//	@Param			transaction_id	query		string	false	"Transaction ID. If not provided, the last transaction will be used."
//	@Success		200				{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/v1/items [get]
func GetItems(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")
		transactionID := c.Query("transaction_id")

		if machineID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, "machine_id is required")
			return
		}

		if transactionID == "" {
			// If no transaction_id provided, get the latest one
			row := database.QueryRow(`
        SELECT transaction_id
        FROM public.transaction_items
        WHERE machine_id = $1
        ORDER BY transaction_id DESC
        LIMIT 1`, machineID)
			if err := row.Scan(&transactionID); err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, "No transactions found for this machine")
				return
			}
		}

		var transaction models.Transaction
		var beginTime sql.NullTime
		var endTime sql.NullTime

		// Query for transaction details using QueryRow since we expect only one row
		row := database.QueryRow(`
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

		err := row.Scan(
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

		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		if err != nil {
			logger.Error("Error querying transaction: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if beginTime.Valid {
			transaction.BeginTime = &beginTime.Time
		}
		if endTime.Valid {
			transaction.EndTime = &endTime.Time
		}

		// details about the items

		rows, err := database.Query(`
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
			logger.Error("Error querying items: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		for rows.Next() {
			var transactionItem models.TransactionItem
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
				logger.Error("Error reading transaction item: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			transaction.Items = append(transaction.Items, transactionItem)
		}

		c.JSON(http.StatusOK, transaction)
	}
}
