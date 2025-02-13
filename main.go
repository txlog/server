package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron/v2"
	_ "github.com/joho/godotenv/autoload"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/txlog/server/database"
	_ "github.com/txlog/server/docs"
	"github.com/txlog/server/execution"
	"github.com/txlog/server/machineID"
	"github.com/txlog/server/transaction"
	"github.com/txlog/server/transactionItem"
	"github.com/txlog/server/util"
	"golang.org/x/exp/rand"
)

// @title			Txlog Server
// @version		1.0
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
		// txlog version
		v1.GET("/version", getVersion)

		// txlog build
		v1.GET("/transactions/ids", transaction.GetTransactionIDs(database.Db))
		v1.POST("/transactions", transaction.PostTransaction(database.Db))
		v1.POST("/executions", execution.PostExecution(database.Db))

		// txlog machine_id \
		//   --hostname=G15.example.com
		v1.GET("/machines/id", machineID.GetMachineID(database.Db))

		// txlog executions \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		//   --success=true \
		v1.GET("/executions", execution.GetExecution(database.Db))

		// txlog transactions \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		v1.GET("/transactions", transaction.GetTransactions(database.Db))

		// txlog items \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		//   --transaction_id=4
		v1.GET("/items/ids", transactionItem.GetItemIDs(database.Db))
		v1.GET("/items", transactionItem.GetItems(database.Db))
	}

	s, _ := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()

	rand.Seed(uint64(time.Now().UnixNano()))
	crontab := fmt.Sprintf("%d * * * * *", rand.Intn(59)+1)
	_, _ = s.NewJob(
		gocron.CronJob(
			// Run every minute at a random second
			crontab,
			true,
		),
		gocron.NewTask(
			func() {
				retentionDays := os.Getenv("EXECUTION_RETENTION_DAYS")
				if retentionDays == "" {
					retentionDays = "7" // default to 7 days if not set
				}
				_, _ = database.Db.Exec("DELETE FROM executions WHERE executed_at < NOW() - INTERVAL '" + retentionDays + " day'")

				fmt.Println("Housekeeping: executions older than " + retentionDays + " days are deleted.")
			},
		),
	)

	s.Start()
	fmt.Println("Scheduler started: " + crontab)

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
		"version": "1.0",
	})
}
