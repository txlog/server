package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/database"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
	"github.com/txlog/server/util"
)

// GetAdminIndex displays the admin panel
func GetAdminIndex(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := getAllUsers(db)
		if err != nil {
			logger.Error("Failed to get users: " + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"title": "Internal Server Error",
				"error": "Failed to load users",
			})
			return
		}

		// Get migration status
		migrationStatus, err := getMigrationStatus(db)
		if err != nil {
			logger.Error("Failed to get migration status: " + err.Error())
			// Don't fail the page, just show empty migration status
			migrationStatus = &models.MigrationStatus{}
		}

		// Get API keys
		apiKeys, err := getAllAPIKeys(db)
		if err != nil {
			logger.Error("Failed to get API keys: " + err.Error())
			// Don't fail the page, just show empty API keys list
			apiKeys = []models.ApiKey{}
		}

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"Context":    c,
			"title":      "Administration - Txlog Server",
			"users":      users,
			"migrations": migrationStatus,
			"apiKeys":    apiKeys,
		})
	}
}

// PostAdminUpdateUser updates user information
func PostAdminUpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.PostForm("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		isActive := c.PostForm("is_active") == "on"
		isAdmin := c.PostForm("is_admin") == "on"

		err = updateUser(db, userID, isActive, isAdmin)
		if err != nil {
			logger.Error("Failed to update user: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		logger.Info("User " + userIDStr + " updated successfully")
		c.Redirect(http.StatusSeeOther, "/admin")
	}
}

// PostAdminDeleteUser deactivates a user
func PostAdminDeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.PostForm("user_id")
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		err = deactivateUser(db, userID)
		if err != nil {
			logger.Error("Failed to deactivate user: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate user"})
			return
		}

		logger.Info("User " + userIDStr + " deactivated successfully")
		c.Redirect(http.StatusSeeOther, "/admin")
	}
}

// getAllUsers retrieves all users from the database
func getAllUsers(db *sql.DB) ([]models.User, error) {
	query := `
		SELECT id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin,
		       created_at, updated_at, last_login_at
		FROM users
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
			&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
			&user.LastLoginAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// updateUser updates user status in the database
func updateUser(db *sql.DB, userID int, isActive, isAdmin bool) error {
	query := `
		UPDATE users
		SET is_active = $1, is_admin = $2, updated_at = $3
		WHERE id = $4
	`

	_, err := db.Exec(query, isActive, isAdmin, time.Now(), userID)
	return err
}

// deactivateUser deactivates a user in the database
func deactivateUser(db *sql.DB, userID int) error {
	query := `
		UPDATE users
		SET is_active = false, updated_at = $1
		WHERE id = $2
	`

	_, err := db.Exec(query, time.Now(), userID)
	return err
}

// PostAdminRunMigrations runs all pending migrations
func PostAdminRunMigrations(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First check if database is dirty and force clean if needed
		if err := database.ForceCleanIfDirty(); err != nil {
			logger.Error("Failed to clean dirty state: " + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"title": "Migration Error",
				"error": "Failed to clean dirty state: " + err.Error(),
			})
			return
		}

		// Apply all pending migrations using database package function
		err := database.RunAllMigrations()
		if err != nil {
			logger.Error("Failed to run migrations: " + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"title": "Migration Error",
				"error": "Failed to apply migrations: " + err.Error(),
			})
			return
		}

		logger.Info("All pending migrations applied successfully via admin panel")
		c.Redirect(http.StatusSeeOther, "/admin?migration_success=1")
	}
}

// PostAdminForceCleanMigration forces the database migration state to clean
func PostAdminForceCleanMigration(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := database.ForceCleanIfDirty()
		if err != nil {
			logger.Error("Failed to force clean migration state: " + err.Error())
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"title": "Migration Error",
				"error": "Failed to force clean migration state: " + err.Error(),
			})
			return
		}

		logger.Info("Database migration forced to clean state via admin panel")
		c.Redirect(http.StatusSeeOther, "/admin?migration_success=1")
	}
}

// getMigrationStatus gets the current migration status
func getMigrationStatus(db *sql.DB) (*models.MigrationStatus, error) {
	status := &models.MigrationStatus{}

	// Check if schema_migrations table exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check schema_migrations table: %w", err)
	}

	if exists {
		// Get current migration version and dirty state
		err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&status.CurrentVersion, &status.IsDirty)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to get current migration version: %w", err)
		}
	}

	// Get all available migrations from database package
	allMigrations, err := database.GetAllAvailableMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get available migrations: %w", err)
	}

	// Separate applied and pending migrations
	for _, migration := range allMigrations {
		if migration.Version <= status.CurrentVersion {
			migration.Applied = true
			status.Applied = append(status.Applied, migration)
		} else {
			status.Pending = append(status.Pending, migration)
		}
	}

	status.TotalCount = len(allMigrations)
	return status, nil
}

// getAllAPIKeys retrieves all API keys from the database
func getAllAPIKeys(db *sql.DB) ([]models.ApiKey, error) {
	query := `
		SELECT
			ak.id,
			ak.name,
			ak.key_prefix,
			ak.created_at,
			ak.last_used_at,
			ak.is_active,
			ak.created_by,
			COALESCE(u.name, 'System') as creator_name
		FROM api_keys ak
		LEFT JOIN users u ON ak.created_by = u.id
		ORDER BY ak.created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apiKeys []models.ApiKey
	for rows.Next() {
		var key models.ApiKey
		err := rows.Scan(
			&key.ID,
			&key.Name,
			&key.KeyPrefix,
			&key.CreatedAt,
			&key.LastUsedAt,
			&key.IsActive,
			&key.CreatedBy,
			&key.CreatorName,
		)
		if err != nil {
			return nil, err
		}
		apiKeys = append(apiKeys, key)
	}

	return apiKeys, rows.Err()
}

// PostAdminCreateAPIKey creates a new API key
func PostAdminCreateAPIKey(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.PostForm("name")
		if name == "" || len(name) < 3 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name must be at least 3 characters"})
			return
		}

		// Generate API key
		fullKey, keyHash, keyPrefix, err := util.GenerateAPIKey()
		if err != nil {
			logger.Error("Failed to generate API key: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate API key"})
			return
		}

		// Get user ID from context if available
		var createdBy *int
		if userInterface, exists := c.Get("user"); exists {
			if user, ok := userInterface.(*models.User); ok {
				createdBy = &user.ID
			}
		}

		// Insert into database
		var keyID int
		query := `
			INSERT INTO api_keys (name, key_hash, key_prefix, created_by, created_at, is_active)
			VALUES ($1, $2, $3, $4, $5, true)
			RETURNING id
		`
		err = db.QueryRow(query, name, keyHash, keyPrefix, createdBy, time.Now()).Scan(&keyID)
		if err != nil {
			logger.Error("Failed to insert API key: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
			return
		}

		logger.Info(fmt.Sprintf("API key created: ID=%d, Name=%s", keyID, name))

		// Return the full key (this is the only time it will be shown)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"id":      keyID,
			"name":    name,
			"key":     fullKey,
			"message": "API key created successfully. Save this key now - it won't be shown again!",
		})
	}
}

// PostAdminRevokeAPIKey revokes (deactivates) an API key
func PostAdminRevokeAPIKey(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyIDStr := c.PostForm("key_id")
		keyID, err := strconv.Atoi(keyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
			return
		}

		query := `UPDATE api_keys SET is_active = false WHERE id = $1`
		_, err = db.Exec(query, keyID)
		if err != nil {
			logger.Error("Failed to revoke API key: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke API key"})
			return
		}

		logger.Info(fmt.Sprintf("API key revoked: ID=%d", keyID))
		c.Redirect(http.StatusSeeOther, "/admin?apikey_revoked=1")
	}
}

// DeleteAdminAPIKey permanently deletes an API key
func DeleteAdminAPIKey(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyIDStr := c.PostForm("key_id")
		keyID, err := strconv.Atoi(keyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API key ID"})
			return
		}

		query := `DELETE FROM api_keys WHERE id = $1`
		_, err = db.Exec(query, keyID)
		if err != nil {
			logger.Error("Failed to delete API key: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
			return
		}

		logger.Info(fmt.Sprintf("API key deleted: ID=%d", keyID))
		c.Redirect(http.StatusSeeOther, "/admin?apikey_deleted=1")
	}
}
