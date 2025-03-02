package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"regexp"
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

//go:embed assets
var staticFiles embed.FS

//go:embed templates/*
var templateFS embed.FS

// @title			Txlog Server
// @version		1.1.1
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

	database.ConnectDatabase()
	startScheduler()

	r := gin.Default()
	r.SetTrustedProxies(nil)

	if os.Getenv("GIN_MODE") == "" {
		tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))
		r.SetHTMLTemplate(tmpl)

		fsys, _ := fs.Sub(staticFiles, "assets")
		r.StaticFS("/assets", http.FS(fsys))
	} else {
		r.LoadHTMLGlob("templates/*.html")
		r.Static("/assets", "./assets")
	}

	healthcheck.New(r, util.CheckConfig(), util.Check())

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Context": c,
			"title":   "Assets",
		})
	})

	r.GET("/settings", func(c *gin.Context) {
		c.HTML(http.StatusOK, "settings.html", gin.H{
			"Context":                c,
			"title":                  "Server Settings",
			"pgsqlHost":              os.Getenv("PGSQL_HOST"),
			"pgsqlPort":              os.Getenv("PGSQL_PORT"),
			"pgsqlUser":              os.Getenv("PGSQL_USER"),
			"pgsqlDb":                os.Getenv("PGSQL_DB"),
			"pgsqlPassword":          util.MaskString(os.Getenv("PGSQL_PASSWORD")),
			"pgsqlSslmode":           os.Getenv("PGSQL_SSLMODE"),
			"executionRetentionDays": os.Getenv("EXECUTION_RETENTION_DAYS"),
		})
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.DocExpansion("none"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

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
		v1.GET("/machines/ids", machineID.GetMachineID(database.Db))

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
		"version": "1.1.1",
	})
}

// startScheduler initializes and runs a scheduled task for database housekeeping.
// It creates a new scheduler that runs every 2 hours at a random minute/second
// to delete old execution records from the database. The retention period is
// controlled by the EXECUTION_RETENTION_DAYS environment variable (defaults to 7 days).
// The random scheduling helps distribute the load when running multiple instances.
func startScheduler() {
	s, _ := gocron.NewScheduler()
	defer func() { _ = s.Shutdown() }()

	rand.Seed(uint64(time.Now().UnixNano()))
	// Run every two hours at a random minute/second
	crontab := fmt.Sprintf("%d %d */2 * * *", rand.Intn(59), rand.Intn(59))
	_, _ = s.NewJob(
		gocron.CronJob(
			crontab,
			true,
		),
		gocron.NewTask(
			func() {
				retentionDays := os.Getenv("EXECUTION_RETENTION_DAYS")
				if retentionDays == "" {
					retentionDays = "7" // default to 7 days if not set
				}
				if regexp.MustCompile(`^[0-9]+$`).MatchString(retentionDays) {
					_, _ = database.Db.Exec("DELETE FROM executions WHERE executed_at < NOW() - INTERVAL '" + retentionDays + " day'")
				}

				fmt.Println("Housekeeping: executions older than " + retentionDays + " days are deleted.")
			},
		),
	)

	s.Start()
	fmt.Println("Scheduler started: " + crontab)
}
