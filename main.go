package main

import (
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/txlog/server/database"
	_ "github.com/txlog/server/docs"
	"github.com/txlog/server/execution"
	"github.com/txlog/server/transaction"
	"github.com/txlog/server/util"
)

// @title			Txlog Server
// @version		0.2
// @description	The centralized system that stores transaction data
// @termsOfService	https://github.com/txlog
// @contact.name	Txlog repository issues
// @contact.url	https://github.com/txlog/server/issues
// @license.name	MIT License
// @license.url	https://github.com/txlog/.github/blob/main/profile/LICENSE.md
// @host			localhost:8080
// @schemes		http https
func main() {
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	database.ConnectDatabase()

	healthcheck.New(r, util.CheckConfig(), util.Check())

	r.GET("/", func(c *gin.Context) { c.Redirect(302, "/swagger/index.html") })
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	v1 := r.Group("/v1")
	{
		v1.GET("/version", getVersion)
		v1.GET("/transaction", transaction.GetTransaction(database.Db))
		v1.POST("/transaction", transaction.PostTransaction(database.Db))

		v1.POST("/execution", execution.PostExecution(database.Db))
	}

	r.Run()
}

// getVersion Get the server version
//
//	@Summary		Server version
//	@Description	Get the server version
//	@Tags			version
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	interface{}
//	@Router			/v1/version [get]
func getVersion(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"version": "0.2",
	})
}
