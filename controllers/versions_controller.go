package controllers

import (
	"github.com/gin-gonic/gin"
)

func GetVersions(version string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"version": version,
		})
	}
}
