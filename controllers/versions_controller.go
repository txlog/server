package controllers

import (
	"github.com/gin-gonic/gin"
)

// GetVersions returns a gin.HandlerFunc that responds with the application version.
// It takes a version string parameter and returns a JSON response containing the version.
//
// Parameters:
//   - version: string representing the current application version
//
// Returns:
//   - gin.HandlerFunc that handles HTTP requests by returning version information
func GetVersions(version string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"version": version,
		})
	}
}
