package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetInsightsIndex(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "insights.html", gin.H{
		"Context": ctx,
		"title":   "Insights on transaction data",
	})
}
