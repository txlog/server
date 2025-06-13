package controllers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetPackagesIndex(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "packages.html", gin.H{
			"Context": c,
			"title":   "Packages",
		})
	}
}
