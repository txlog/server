package v1

import (
	"github.com/gin-gonic/gin"
)

// GetVersions returns a JSON response containing the server version.
//
//	@Summary		Get server version
//	@Description	Get server version
//	@Tags			version
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	interface{}
//	@Router			/v1/version [get]
func GetVersions(version string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"version": version,
		})
	}
}
