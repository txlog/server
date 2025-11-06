package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/txlog/server/models"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
)

// GetTransactionIDs Get the saved transactions IDs for a host
//
//	@Summary		Get saved transactions IDs for a host
//	@Description	Get saved transactions IDs for a host
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/v1/transactions/ids [get]
func GetTransactionIDs(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := models.Transaction{}
		data, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid transaction data")
			return
		}
		err = json.Unmarshal(data, &body)
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid JSON input")
			logger.Error("Invalid JSON input: " + err.Error())
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
			logger.Error("Couldn't get saved transaction_ids for this host: " + err.Error())
			c.AbortWithStatusJSON(400, "Couldn't get saved transaction_ids for this host.")
		} else {
			var transactions []int
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					logger.Error("Error scanning transaction_ids: " + err.Error())
					c.AbortWithStatusJSON(500, "Error scanning transaction_ids")
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

// GetTransactions Get the saved transactions for a host
//
//	@Summary		Get saved transactions for a host
//	@Description	Get saved transactions for a host
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			machine_id	query		string	false	"Machine ID"
//	@Success		200			{object}	interface{}
//	@Security		ApiKeyAuth
//	@Router			/v1/transactions [get]
func GetTransactions(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		machineID := c.Query("machine_id")

		if machineID == "" {
			c.AbortWithStatusJSON(400, "machine_id is required")
			return
		}

		var rows *sql.Rows
		var err error

		rows, err = database.Query(`
      SELECT transaction_id, hostname, begin_time, end_time, actions, altered, "user", return_code,
           release_version, command_line, comment, scriptlet_output
      FROM public.transactions
      WHERE machine_id = $1
      ORDER BY transaction_id DESC`,
			machineID,
		)

		if err != nil {
			logger.Error("Error querying transactions: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		transactions := []models.Transaction{}
		for rows.Next() {
			var transaction models.Transaction
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
				logger.Error("Error iterating transactions: " + err.Error())
				c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
				return
			}
			if beginTime.Valid {
				transaction.BeginTime = &beginTime.Time
			}
			if endTime.Valid {
				transaction.EndTime = &endTime.Time
			}
			transactions = append(transactions, transaction)
		}

		c.JSON(http.StatusOK, transactions)
	}
}

// PostTransactions Create a new transaction
//
//	@Summary		Create a new transaction
//	@Description	Create a new transaction
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			Transaction	body		models.Transaction	true	"Transaction data"
//	@Success		200			{string}	string				"Transaction created"
//	@Failure		400			{string}	string				"Invalid transaction data"
//	@Failure		400			{string}	string				"Invalid JSON input"
//	@Failure		500			{string}	string				"Database error"
//	@Security		ApiKeyAuth
//	@Router			/v1/transactions [post]
func PostTransactions(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		body := models.Transaction{}
		data, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid transaction data")
			return
		}
		err = json.Unmarshal(data, &body)
		if err != nil {
			c.AbortWithStatusJSON(400, "Invalid JSON input")
			logger.Error("Invalid JSON input: " + err.Error())
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
			logger.Error("Error beginning transaction: " + err.Error())
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
			logger.Error("Error inserting transaction: " + err.Error())
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
				logger.Error("Error inserting transaction item: " + err.Error())
				c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
				return
			}
		}

		assetManager := models.NewAssetManager(database)
		timestamp := body.BeginTime
		if timestamp == nil {
			now := beginTime.Time
			timestamp = &now
		}
		err = assetManager.UpsertAsset(tx, body.Hostname, body.MachineID, *timestamp, sql.NullBool{}, sql.NullString{})
		if err != nil {
			tx.Rollback()
			logger.Error("Error upserting asset:" + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": "Failed to update asset registry"})
			return
		}

		// Commit the database transaction
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			logger.Error("Error committing transaction: " + err.Error())
			c.AbortWithStatusJSON(500, gin.H{"error": "Database error"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transaction created"})
	}
}
