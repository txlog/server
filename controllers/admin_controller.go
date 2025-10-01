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

		c.HTML(http.StatusOK, "admin.html", gin.H{
			"Context":    c,
			"title":      "Administration - Txlog Server",
			"users":      users,
			"migrations": migrationStatus,
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
		c.Redirect(http.StatusTemporaryRedirect, "/admin")
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
		c.Redirect(http.StatusTemporaryRedirect, "/admin")
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
//
//	@Summary		Run Database Migrations
//	@Description	Applies all pending database migrations (Admin only)
//	@Tags			admin
//	@Accept			x-www-form-urlencoded
//	@Produce		json
//	@Success		302	{string}	string	"Redirect to admin panel"
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/admin/migrations/run [post]
func PostAdminRunMigrations(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		c.Redirect(http.StatusTemporaryRedirect, "/admin?migration_success=1")
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
