package main

import (
	"github.com/gin-gonic/gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/txlog/server/database"
	"github.com/txlog/server/transaction"
	"github.com/txlog/server/util"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	database.ConnectDatabase()

	healthcheck.New(r, util.CheckConfig(), util.Check())

	r.GET("/v1/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "0.1",
		})
	})

	r.GET("/v1/transaction", func(c *gin.Context) {
		transaction.GetTransaction(c, database.Db)
	})

	r.POST("/v1/transaction", func(c *gin.Context) {
		transaction.PostTransaction(c, database.Db)
	})

	r.Run()
}
