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
func GetLogin(oidcService *auth.OIDCService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is already logged in
		if sessionID, err := c.Cookie("session_id"); err == nil && sessionID != "" {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			return
		}

		c.HTML(http.StatusOK, "login.html", gin.H{
			"title": "Login - Txlog Server",
		})
	}
}

// PostLogin initiates OIDC authentication flow
//
//	@Summary		Initiate OIDC authentication
//	@Description	Redirects to OIDC provider for authentication
//	@Tags			auth
//	@Produce		json
//	@Success		302	{string}	string	"Redirect to OIDC provider"
//	@Failure		500	{object}	map[string]string
//	@Router			/auth/login [post]
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
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// GetCallback handles OIDC callback
//
//	@Summary		Handle OIDC callback
//	@Description	Process OIDC callback and create user session
//	@Tags			auth
//	@Param			code	query	string	true	"Authorization code"
//	@Param			state	query	string	true	"State parameter"
//	@Produce		json
//	@Success		302	{string}	string	"Redirect to dashboard"
//	@Failure		400	{object}	map[string]string
//	@Failure		401	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/auth/callback [get]
func GetCallback(oidcService *auth.OIDCService) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" {
			logger.Error("Authorization code is missing")
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=auth_failed")
			return
		}

		// Verify state parameter
		storedState, err := c.Cookie("oidc_state")
		if err != nil || storedState != state {
			logger.Error("State parameter mismatch")
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=invalid_state")
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
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=token_exchange_failed")
			return
		}

		// Extract ID token
		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			logger.Error("ID token is missing")
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=id_token_missing")
			return
		}

		// Verify ID token
		idToken, err := oidcService.VerifyIDToken(ctx, rawIDToken)
		if err != nil {
			logger.Error("Failed to verify ID token: " + err.Error())
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=id_token_invalid")
			return
		}

		// Create or update user
		user, err := oidcService.CreateOrUpdateUser(ctx, idToken)
		if err != nil {
			logger.Error("Failed to create/update user: " + err.Error())
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=user_creation_failed")
			return
		}

		if !user.IsActive {
			logger.Warn("User " + user.Email + " is not active")
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=account_disabled")
			return
		}

		// Create user session
		sessionID, err := oidcService.CreateUserSession(user.ID)
		if err != nil {
			logger.Error("Failed to create user session: " + err.Error())
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=session_creation_failed")
			return
		}

		// Set session cookie
		c.SetCookie("session_id", sessionID, 7*24*3600, "/", "", false, true) // 7 days

		logger.Info("User " + user.Email + " logged in successfully")
		c.Redirect(http.StatusTemporaryRedirect, "/")
	}
}

// PostLogout logs out the user
//
//	@Summary		Logout user
//	@Description	Invalidates user session and redirects to home page
//	@Tags			auth
//	@Produce		json
//	@Success		303	{string}	string	"Redirect to home page"
//	@Router			/auth/logout [post]
func PostLogout(oidcService *auth.OIDCService) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err == nil && sessionID != "" {
			if err := oidcService.InvalidateUserSession(sessionID); err != nil {
				logger.Error("Failed to invalidate user session: " + err.Error())
			}
		}

		// Clear session cookie
		c.SetCookie("session_id", "", -1, "/", "", false, true)

		c.Redirect(http.StatusSeeOther, "/")
	}
}
