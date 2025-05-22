package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/txlog/server/controllers"
	"github.com/txlog/server/database"
	_ "github.com/txlog/server/docs"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/scheduler"
	"github.com/txlog/server/util"
	"github.com/txlog/server/version"
)

//go:embed images
var staticFiles embed.FS

//go:embed templates/*
var templateFS embed.FS

// @title			Txlog Server
// @version		v1
// @description	The centralized system that stores transaction data
// @termsOfService	https://github.com/txlog
// @contact.name	Txlog repository issues
// @contact.url	https://github.com/txlog/server/issues
// @license.name	MIT License
// @license.url	https://github.com/txlog/.github/blob/main/profile/LICENSE.md
// @host			localhost:8080
// @schemes		http https
func main() {
	logger.InitLogger()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.ConnectDatabase()
	scheduler.StartScheduler()

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(EnvironmentVariablesMiddleware())

	funcMap := template.FuncMap{
		"text2html":        util.Text2HTML,
		"formatPercentage": util.FormatPercentage,
		"formatInteger":    util.FormatInteger,
		"iterate":          util.Iterate,
		"add":              util.Add,
		"min":              util.Min,
		"version":          util.Version,
		"dnfUser":          util.DnfUser,
		"brand":            util.Brand,
		"hasAction":        util.HasAction,
	}

	if os.Getenv("GIN_MODE") == "" {
		tmpl := template.Must(template.New("any").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))
		r.SetHTMLTemplate(tmpl)

		fsys, _ := fs.Sub(staticFiles, "images")
		r.StaticFS("/images", http.FS(fsys))
	} else {
		r.SetFuncMap(funcMap)
		r.LoadHTMLGlob("templates/*.html")
		r.Static("/images", "./images")
	}

	healthcheck.New(r, util.CheckConfig(), util.Check())

	r.NoRoute(controllers.Get404)

	r.GET("/", controllers.GetRootIndex(database.Db))
	r.GET("/assets", controllers.GetAssetsIndex(database.Db))
	r.GET("/executions/:execution_id", controllers.GetExecutionID(database.Db))
	r.GET("/insights", controllers.GetInsightsIndex)
	r.GET("/license", controllers.GetLicensesIndex)
	r.GET("/assets/:machine_id", controllers.GetMachineID(database.Db))
	r.GET("/sponsor", controllers.GetSponsorIndex)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.PersistAuthorization(true),
		ginSwagger.DocExpansion("none"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	v1 := r.Group("/v1")
	{
		// txlog version
		v1.GET("/version", controllers.GetVersions(version.SemVer))

		// txlog build
		v1.GET("/transactions/ids", controllers.GetTransactionIDs(database.Db))
		v1.POST("/transactions", controllers.PostTransactions(database.Db))
		v1.POST("/executions", controllers.PostExecutions(database.Db))

		// txlog machine_id \
		//   --hostname=G15.example.com
		v1.GET("/machines/ids", controllers.GetMachineIDs(database.Db))
		v1.GET("/machines", controllers.GetMachines(database.Db))

		// txlog executions \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		//   --success=true \
		v1.GET("/executions", controllers.GetExecutions(database.Db))

		// txlog transactions \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		v1.GET("/transactions", controllers.GetTransactions(database.Db))

		// txlog items \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		//   --transaction_id=4
		v1.GET("/items/ids", controllers.GetItemIDs(database.Db))
		v1.GET("/items", controllers.GetItems(database.Db))
	}

	v2 := r.Group("/v2")
	{
		// txlog version
		v2.GET("/version", controllers.GetVersions(version.SemVer))

		// txlog build
		v2.GET("/transactions/ids", controllers.GetTransactionIDs(database.Db))
		v2.POST("/transactions", controllers.PostTransactions(database.Db))
		v2.POST("/executions", controllers.PostExecutions(database.Db))

		// txlog machine_id \
		//   --hostname=G15.example.com
		v2.GET("/assets/ids", controllers.GetMachineIDs(database.Db))
		v2.GET("/assets", controllers.GetMachines(database.Db))
		v2.GET("/assets/requiring-restart", controllers.GetMachines(database.Db))

		// txlog executions \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		//   --success=true \
		v2.GET("/executions", controllers.GetExecutions(database.Db))

		// txlog transactions \
		//   --machine_id=e250c98c14e947ba96359223785375bb
		v2.GET("/transactions", controllers.GetTransactions(database.Db))

		// txlog items \
		//   --machine_id=e250c98c14e947ba96359223785375bb \
		//   --transaction_id=4
		v2.GET("/items/ids", controllers.GetItemIDs(database.Db))
		v2.GET("/items", controllers.GetItems(database.Db))
	}

	r.Run()
}

func EnvironmentVariablesMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		envVars := map[string]string{
			"instance":                 os.Getenv("INSTANCE"),
			"pgsqlHost":                os.Getenv("PGSQL_HOST"),
			"pgsqlPort":                os.Getenv("PGSQL_PORT"),
			"pgsqlUser":                os.Getenv("PGSQL_USER"),
			"pgsqlDb":                  os.Getenv("PGSQL_DB"),
			"pgsqlPassword":            util.MaskString(os.Getenv("PGSQL_PASSWORD")),
			"pgsqlSslmode":             os.Getenv("PGSQL_SSLMODE"),
			"cronRetentionDays":        os.Getenv("CRON_RETENTION_DAYS"),
			"cronRetentionExpression":  os.Getenv("CRON_RETENTION_EXPRESSION"),
			"cronStatisticsExpression": os.Getenv("CRON_STATS_EXPRESSION"),
			"ignoreEmptyExecution":     os.Getenv("IGNORE_EMPTY_EXECUTION"),
			"latestVersion":            os.Getenv("LATEST_VERSION"),
		}

		c.Set("env", envVars)

		c.Next()
	}
}
