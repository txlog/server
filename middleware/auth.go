package middleware

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/auth"
	"github.com/txlog/server/models"
)

// AuthMiddleware checks for valid user sessions
// If OIDC is not configured, it allows all requests through
func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If OIDC is not configured, skip authentication entirely
		if !auth.IsConfigured() {
			c.Next()
			return
		}

		// Skip authentication for API endpoints and health checks
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/v1/") ||
			strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/auth/") ||
			strings.HasPrefix(path, "/images/") ||
			strings.HasPrefix(path, "/debug/") ||
			path == "/login" ||
			path == "/logout" {
			c.Next()
			return
		}

		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		user, err := getUserBySessionID(db, sessionID)
		if err != nil || !user.IsActive {
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

		// Set user in context for use in handlers
		c.Set("user", user)
		c.Next()
	}
}

// AdminMiddleware checks if user has admin privileges
// If OIDC is not configured, it allows all requests through
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If OIDC is not configured, skip admin check
		if !auth.IsConfigured() {
			c.Next()
			return
		}

		userInterface, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		user, ok := userInterface.(*models.User)
		if !ok || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin privileges required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func getUserBySessionID(db *sql.DB, sessionID string) (*models.User, error) {
	query := `
		SELECT u.id, u.sub, u.email, u.name, COALESCE(u.picture, '') as picture, u.is_active, u.is_admin,
		       u.created_at, u.updated_at, u.last_login_at
		FROM users u
		INNER JOIN user_sessions s ON u.id = s.user_id
		WHERE s.id = $1 AND s.is_active = true AND s.expires_at > NOW()
	`

	user := &models.User{}
	err := db.QueryRow(query, sessionID).Scan(
		&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt,
		&user.LastLoginAt,
	)

	return user, err
}
