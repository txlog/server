package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// GetSponsorIndex handles HTTP GET requests for the sponsor page.
// It renders the sponsor.html template with the given context and title.
//
// Parameters:
//   - ctx: A pointer to the Gin context containing HTTP request and response information
//
// Returns:
//   - Renders the sponsor.html template with HTTP 200 OK status
func GetSponsorIndex(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "sponsor.html", gin.H{
		"Context": ctx,
		"title":   "Sponsor",
	})
}
