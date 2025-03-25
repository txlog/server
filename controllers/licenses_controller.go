package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// GetLicensesIndex handles HTTP GET requests for the license page.
// It renders the license.html template with the given context and title.
//
// Parameters:
//   - ctx: A pointer to the Gin context containing HTTP request and response information
//
// Returns:
//   - Renders the license.html template with HTTP 200 OK status
func GetLicensesIndex(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "license.html", gin.H{
		"Context": ctx,
		"title":   "License",
	})
}
