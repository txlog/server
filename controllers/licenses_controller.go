package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func GetLicensesIndex(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "license.html", gin.H{
		"Context": ctx,
		"title":   "License",
	})
}
