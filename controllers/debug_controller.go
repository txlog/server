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

				// Check if users table exists and get count
				var userCount int
				err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
				if err == nil {
					debug["users_count"] = userCount
				} else {
					debug["users_table_error"] = err.Error()
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
