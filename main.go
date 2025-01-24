package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/txlog/server/database"
	"github.com/txlog/server/transaction"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	database.ConnectDatabase()

	r.GET("/v1/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/v1/transaction", func(c *gin.Context) {
		transaction.PostTransaction(c, database.Db)
	})

	r.Run()
}
