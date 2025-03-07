package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Get500(ctx *gin.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, "/v1") {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
		})
	} else {
		ctx.HTML(http.StatusInternalServerError, "500.html", gin.H{
			"Context": ctx,
			"title":   "Internal Server Error",
		})
	}
}
