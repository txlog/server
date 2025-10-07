package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/auth"
	logger "github.com/txlog/server/logger"
)

// GetLogin displays the login page
func GetLogin(oidcService *auth.OIDCService, ldapService *auth.LDAPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is already logged in
		if sessionID, err := c.Cookie("session_id"); err == nil && sessionID != "" {
			c.Redirect(http.StatusSeeOther, "/")
			return
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title":        "Login - Txlog Server",
			"ldap_enabled": ldapService != nil,
			"oidc_enabled": oidcService != nil,
		})
	}
}

// PostLogin initiates OIDC authentication flow
func PostLogin(oidcService *auth.OIDCService) gin.HandlerFunc {
	return func(c *gin.Context) {
		state, err := auth.GenerateState()
		if err != nil {
			logger.Error("Failed to generate state: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}

		// Store state in session/cookie for verification
		c.SetCookie("oidc_state", state, 300, "/", "", false, true) // 5 minutes

		authURL := oidcService.GetAuthURL(state)
		c.Redirect(http.StatusSeeOther, authURL)
	}
}

// GetCallback handles OIDC callback
func GetCallback(oidcService *auth.OIDCService) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" {
			logger.Error("Authorization code is missing")
			c.Redirect(http.StatusSeeOther, "/login?error=auth_failed")
			return
		}

		// Verify state parameter
		storedState, err := c.Cookie("oidc_state")
		if err != nil || storedState != state {
			logger.Error("State parameter mismatch")
			c.Redirect(http.StatusSeeOther, "/login?error=invalid_state")
			return
		}

		// Clear state cookie
		c.SetCookie("oidc_state", "", -1, "/", "", false, true)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Exchange code for tokens
		token, err := oidcService.ExchangeCodeForTokens(ctx, code)
		if err != nil {
			logger.Error("Failed to exchange code for tokens: " + err.Error())
			c.Redirect(http.StatusSeeOther, "/login?error=token_exchange_failed")
			return
		}

		// Extract ID token
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			logger.Error("ID token is missing")
			c.Redirect(http.StatusSeeOther, "/login?error=id_token_missing")
			return
		}

		// Verify ID token
		idToken, err := oidcService.VerifyIDToken(ctx, rawIDToken)
		if err != nil {
			logger.Error("Failed to verify ID token: " + err.Error())
			c.Redirect(http.StatusSeeOther, "/login?error=id_token_invalid")
			return
		}

		// Create or update user
		user, err := oidcService.CreateOrUpdateUser(ctx, idToken)
		if err != nil {
			logger.Error("Failed to create/update user: " + err.Error())
			c.Redirect(http.StatusSeeOther, "/login?error=user_creation_failed")
			return
		}

		if !user.IsActive {
			logger.Warn("User " + user.Email + " is not active")
			c.Redirect(http.StatusSeeOther, "/login?error=account_disabled")
			return
		}

		// Create user session
		sessionID, err := oidcService.CreateUserSession(user.ID)
		if err != nil {
			logger.Error("Failed to create user session: " + err.Error())
			c.Redirect(http.StatusSeeOther, "/login?error=session_creation_failed")
			return
		}

		// Set session cookie
		c.SetCookie("session_id", sessionID, 7*24*3600, "/", "", false, true) // 7 days

		logger.Info("User " + user.Email + " logged in successfully")
		c.Redirect(http.StatusSeeOther, "/")
	}
}

// PostLDAPLogin handles LDAP login with username and password
func PostLDAPLogin(ldapService *auth.LDAPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if username == "" || password == "" {
			logger.Error("Username or password is empty")
			c.Redirect(http.StatusSeeOther, "/login?error=invalid_credentials")
			return
		}

		// Authenticate with LDAP
		user, err := ldapService.Authenticate(username, password)
		if err != nil {
			logger.Error("LDAP authentication failed: " + err.Error())

			// Categorize the error for better user feedback
			errorCode := auth.CategorizeAuthError(err)
			c.Redirect(http.StatusSeeOther, "/login?error="+errorCode)
			return
		}

		if !user.IsActive {
			logger.Warn("User " + user.Email + " is not active")
			c.Redirect(http.StatusSeeOther, "/login?error=account_disabled")
			return
		}

		// Create user session
		sessionID, err := ldapService.CreateUserSession(user.ID)
		if err != nil {
			logger.Error("Failed to create user session: " + err.Error())
			c.Redirect(http.StatusSeeOther, "/login?error=session_creation_failed")
			return
		}

		// Set session cookie
		c.SetCookie("session_id", sessionID, 7*24*3600, "/", "", false, true) // 7 days

		logger.Info("User " + user.Email + " logged in successfully via LDAP")
		c.Redirect(http.StatusSeeOther, "/")
	}
}

// PostLogout logs out the user
func PostLogout(oidcService *auth.OIDCService, ldapService *auth.LDAPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err == nil && sessionID != "" {
			// Try to invalidate session using whichever service is available
			if oidcService != nil {
				if err := oidcService.InvalidateUserSession(sessionID); err != nil {
					logger.Error("Failed to invalidate user session: " + err.Error())
				}
			} else if ldapService != nil {
				if err := ldapService.InvalidateUserSession(sessionID); err != nil {
					logger.Error("Failed to invalidate user session: " + err.Error())
				}
			}
		}

		// Clear session cookie
		c.SetCookie("session_id", "", -1, "/", "", false, true)

		c.Redirect(http.StatusSeeOther, "/")
	}
}
