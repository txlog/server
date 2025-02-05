package main

import (
	"github.com/gin-gonic/gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/tavsec/gin-healthcheck/config"
	"github.com/txlog/server/database"
	"github.com/txlog/server/transaction"
	"github.com/txlog/server/util"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	database.ConnectDatabase()

	healthcheck.New(r, config.DefaultConfig(), util.Checks())

	r.GET("/v1/transaction", func(c *gin.Context) {
		transaction.GetTransaction(c, database.Db)
	})

	r.POST("/v1/transaction", func(c *gin.Context) {
		transaction.PostTransaction(c, database.Db)
	})

	r.Run()
}
