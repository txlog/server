package middleware

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	logger "github.com/txlog/server/logger"
)

// APIKeyMiddleware validates API keys for /v1 endpoints
// It also allows access for users authenticated via session cookie
func APIKeyMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First, try to get API key from header (no DB query needed yet)
		apiKey := c.GetHeader("X-API-Key")

		// Also check Authorization header as fallback (Bearer token format)
		if apiKey == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// If no API key provided, check for session authentication (for UI requests)
		// This avoids unnecessary session DB queries when API key is present
		if apiKey == "" {
			if sessionID, err := c.Cookie("session_id"); err == nil && sessionID != "" {
				if isValidSession(db, sessionID) {
					c.Next()
					return
				}
			}

			// No API key and no valid session
			logger.Warn("API request without API key from " + c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "API key required. Please provide X-API-Key header.",
			})
			c.Abort()
			return
		}

		// Validate API key format (should start with txlog_)
		if !strings.HasPrefix(apiKey, "txlog_") {
			logger.Warn("API request with invalid key format from " + c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key format.",
			})
			c.Abort()
			return
		}

		// Hash the provided API key
		hash := sha256.Sum256([]byte(apiKey))
		keyHash := hex.EncodeToString(hash[:])

		// Check if API key exists and is active
		var keyID int
		var isActive bool
		query := `
			SELECT id, is_active
			FROM api_keys
			WHERE key_hash = $1
		`
		err := db.QueryRow(query, keyHash).Scan(&keyID, &isActive)

		if err == sql.ErrNoRows {
			logger.Warn("API request with non-existent key from " + c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key.",
			})
			c.Abort()
			return
		}

		if err != nil {
			logger.Error("Database error validating API key: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error.",
			})
			c.Abort()
			return
		}

		if !isActive {
			logger.Warn(fmt.Sprintf("API request with inactive key (ID: %d) from %s", keyID, c.ClientIP()))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key.",
			})
			c.Abort()
			return
		}

		// Update last_used_at timestamp (async, don't block request)
		go func() {
			updateQuery := `UPDATE api_keys SET last_used_at = $1 WHERE id = $2`
			_, err := db.Exec(updateQuery, time.Now(), keyID)
			if err != nil {
				logger.Error("Failed to update last_used_at for API key: " + err.Error())
			}
		}()

		// Store API key ID in context for potential logging
		c.Set("api_key_id", keyID)

		c.Next()
	}
}

// isValidSession checks if a session ID corresponds to a valid, active session
func isValidSession(db *sql.DB, sessionID string) bool {
	query := `
		SELECT COUNT(*)
		FROM user_sessions s
		INNER JOIN users u ON s.user_id = u.id
		WHERE s.id = $1 AND s.is_active = true AND s.expires_at > NOW() AND u.is_active = true
	`
	var count int
	err := db.QueryRow(query, sessionID).Scan(&count)
	return err == nil && count > 0
}
