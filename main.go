package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/txlog/server/database"
	"github.com/txlog/server/transaction"
	"github.com/txlog/server/util"
)

func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	database.ConnectDatabase()

	healthcheck.New(r, util.CheckConfig(), util.Check())

	r.GET("/v1/version", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "0.2",
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
