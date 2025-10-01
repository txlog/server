package controllers

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/auth"
	logger "github.com/txlog/server/logger"
)

// GetDebugOIDC provides diagnostic information about OIDC configuration
// This endpoint should be removed or secured before production deployment
//
//	@Summary		OIDC Debug Information
//	@Description	Returns diagnostic information about OIDC configuration and connectivity
//	@Tags			debug
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/debug/oidc [get]
func GetDebugOIDC(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		debug := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		}

		// Basic OIDC configuration check
		debug["oidc_configured"] = auth.IsConfigured()
		debug["environment"] = map[string]interface{}{
			"oidc_client_id_set":     len(os.Getenv("OIDC_CLIENT_ID")) > 0,
			"oidc_client_secret_set": len(os.Getenv("OIDC_CLIENT_SECRET")) > 0,
			"oidc_issuer_url":        os.Getenv("OIDC_ISSUER_URL"),
			"oidc_redirect_url":      os.Getenv("OIDC_REDIRECT_URL"),
			"oidc_skip_tls_verify":   os.Getenv("OIDC_SKIP_TLS_VERIFY"),
		}

		// Database connectivity check
		debug["database_connected"] = false
		if db != nil {
			if err := db.Ping(); err == nil {
				debug["database_connected"] = true

				// Check migrations table and current version
				var migrationExists bool
				err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations')").Scan(&migrationExists)
				if err == nil {
					debug["migration_table_exists"] = migrationExists

					if migrationExists {
						// Get current migration version
						var version int
						var dirty bool
						err = db.QueryRow("SELECT version, dirty FROM schema_migrations LIMIT 1").Scan(&version, &dirty)
						if err == nil {
							debug["current_migration_version"] = version
							debug["migration_dirty"] = dirty
						}
					}
				}

				// Check if users table exists and get count
				var userTableExists bool
				err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&userTableExists)
				if err == nil {
					debug["users_table_exists"] = userTableExists

					if userTableExists {
						var userCount int
						err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
						if err == nil {
							debug["users_count"] = userCount
						} else {
							debug["users_table_error"] = err.Error()
						}
					} else {
						debug["users_table_missing"] = true
						debug["migration_needed"] = "Migration 20250929 not applied - users table missing"
						debug["solution"] = "Restart application to apply pending migrations"
					}
				}
			} else {
				debug["database_error"] = err.Error()
			}
		}

		// OIDC service initialization check
		if auth.IsConfigured() {
			oidcService, err := auth.NewOIDCService(db)
			if err != nil {
				debug["oidc_service_error"] = err.Error()
				logger.Error("Debug: Failed to initialize OIDC service: " + err.Error())
			} else if oidcService != nil {
				debug["oidc_service_initialized"] = true

				// Test auth URL generation
				state, err := auth.GenerateState()
				if err != nil {
					debug["state_generation_error"] = err.Error()
				} else {
					authURL := oidcService.GetAuthURL(state)
					debug["auth_url_generated"] = len(authURL) > 0
					debug["sample_auth_url"] = authURL[:min(100, len(authURL))] + "..."
				}
			} else {
				debug["oidc_service_initialized"] = false
			}
		}

		// Network connectivity test to OIDC provider
		if issuerURL := os.Getenv("OIDC_ISSUER_URL"); issuerURL != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Test with normal TLS verification
			client := &http.Client{Timeout: 5 * time.Second}
			req, err := http.NewRequestWithContext(ctx, "GET", issuerURL+"/.well-known/openid-configuration", nil)
			if err != nil {
				debug["network_test_error"] = "Failed to create request: " + err.Error()
			} else {
				resp, err := client.Do(req)
				if err != nil {
					debug["network_test_error"] = "Failed to connect to OIDC provider: " + err.Error()

					// If TLS fails, test if OIDC_SKIP_TLS_VERIFY would help
					skipTLS := strings.ToLower(os.Getenv("OIDC_SKIP_TLS_VERIFY")) == "true"
					debug["skip_tls_verify_enabled"] = skipTLS

					if skipTLS {
						// Test with TLS verification disabled (as used in production)
						tlsClient := &http.Client{
							Timeout: 5 * time.Second,
							Transport: &http.Transport{
								TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
							},
						}
						tlsReq, _ := http.NewRequestWithContext(ctx, "GET", issuerURL+"/.well-known/openid-configuration", nil)
						tlsResp, tlsErr := tlsClient.Do(tlsReq)
						if tlsErr != nil {
							debug["network_test_skip_tls_error"] = "Even with TLS skip failed: " + tlsErr.Error()
						} else {
							tlsResp.Body.Close()
							debug["network_test_skip_tls_success"] = true
							debug["skip_tls_status"] = tlsResp.StatusCode
							debug["diagnosis"] = "CA certificates missing in container - but OIDC should work with SKIP_TLS_VERIFY=true"
						}
					} else {
						debug["suggestion"] = "Set OIDC_SKIP_TLS_VERIFY=true or add CA certificates to container"
					}
				} else {
					resp.Body.Close()
					debug["network_test_success"] = true
					debug["oidc_provider_status"] = resp.StatusCode
					debug["ca_certificates_ok"] = true
				}
			}
		}

		c.JSON(http.StatusOK, debug)
	}
}

// PostDebugMigrations forces the application of user table migrations
// This endpoint should be removed after resolving migration issues
//
//	@Summary		Force Apply User Migrations
//	@Description	Forces the application of user table migrations (20250929)
//	@Tags			debug
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/debug/migrations [post]
func PostDebugMigrations(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
		}

		if db == nil {
			result["error"] = "Database connection not available"
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		// Check if users table already exists
		var userTableExists bool
		err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&userTableExists)
		if err != nil {
			result["error"] = "Failed to check users table: " + err.Error()
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		if userTableExists {
			result["message"] = "Users table already exists - no migration needed"
			result["users_table_exists"] = true
			c.JSON(http.StatusOK, result)
			return
		}

		// Apply user table migration manually
		userTableSQL := `
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    sub VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    picture VARCHAR(512),
    is_active BOOLEAN DEFAULT true,
    is_admin BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(64) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_sub ON users(sub);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_active ON user_sessions(is_active);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Insert default admin user
INSERT INTO users (sub, email, name, is_active, is_admin) 
VALUES ('admin', 'admin@localhost', 'Administrator', true, true)
ON CONFLICT (sub) DO NOTHING;
`

		// Execute the migration SQL
		_, err = db.Exec(userTableSQL)
		if err != nil {
			result["error"] = "Failed to apply user table migration: " + err.Error()
			logger.Error("Manual migration failed: " + err.Error())
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		// Verify the tables were created
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'users')").Scan(&userTableExists)
		if err != nil {
			result["error"] = "Failed to verify users table creation: " + err.Error()
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		var userSessionsExists bool
		err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_sessions')").Scan(&userSessionsExists)
		if err != nil {
			result["error"] = "Failed to verify user_sessions table creation: " + err.Error()
			c.JSON(http.StatusInternalServerError, result)
			return
		}

		result["success"] = true
		result["users_table_created"] = userTableExists
		result["user_sessions_table_created"] = userSessionsExists
		result["message"] = "User tables migration applied successfully"

		logger.Info("Manual user table migration applied successfully")
		c.JSON(http.StatusOK, result)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
