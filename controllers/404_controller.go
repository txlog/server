package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Get404(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, "/v1") {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "API Resource not found",
		})
	} else {
		ctx.HTML(http.StatusNotFound, "404.html", gin.H{
			"Context": ctx,
			"title":   "Not Found",
		})
	}
}
