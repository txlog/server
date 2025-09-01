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
	v1API "github.com/txlog/server/controllers/api/v1"
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
		"add":              util.Add,
		"brand":            util.Brand,
		"derefBool":        util.DerefBool,
		"dnfUser":          util.DnfUser,
		"formatInteger":    util.FormatInteger,
		"formatPercentage": util.FormatPercentage,
		"hasAction":        util.HasAction,
		"iterate":          util.Iterate,
		"min":              util.Min,
		"text2html":        util.Text2HTML,
		"version":          util.Version,
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
	r.GET("/packages", controllers.GetPackagesIndex(database.Db))
	r.GET("/assets/:machine_id", controllers.GetMachineID(database.Db))
	r.DELETE("/assets/:machine_id", controllers.DeleteMachineID(database.Db))
	r.GET("/assets/:machine_id", controllers.GetMachineID(database.Db))
	r.DELETE("/assets/:machine_id", controllers.DeleteMachineID(database.Db))
	r.GET("/executions/:execution_id", controllers.GetExecutionID(database.Db))
	r.GET("/insights", controllers.GetInsightsIndex)
	r.GET("/license", controllers.GetLicensesIndex)
	r.GET("/assets/:machine_id", controllers.GetMachineID(database.Db))
	r.DELETE("/assets/:machine_id", controllers.DeleteMachineID(database.Db))
	r.GET("/package-progression", controllers.GetPackagesByWeekIndex(database.Db))
	r.GET("/packages", controllers.GetPackagesIndex(database.Db))
	r.GET("/packages/:name", controllers.GetPackageByName(database.Db))
	r.GET("/sponsor", controllers.GetSponsorIndex)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.PersistAuthorization(true),
		ginSwagger.DocExpansion("none"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	r.GET("/package-progression", controllers.GetPackagesByWeekIndex(database.Db))
	r.GET("/packages/:name", controllers.GetPackageByName(database.Db))

	v1Group := r.Group("/v1")
	{
		// txlog version
		v1Group.GET("/version", v1API.GetVersions(version.SemVer))

		// txlog build
		v1Group.GET("/transactions/ids", v1API.GetTransactionIDs(database.Db))
		v1Group.POST("/transactions", v1API.PostTransactions(database.Db))
		v1Group.POST("/executions", v1API.PostExecutions(database.Db))

		// Assets requiring restart
		v1Group.GET("/assets/requiring-restart", v1API.GetAssetsRequiringRestart(database.Db))

		// Endpoints for agent pre-v1.6.0
		v1Group.GET("/machines/ids", v1API.GetMachineIDs(database.Db))
		v1Group.GET("/machines", v1API.GetMachines(database.Db))
		v1Group.GET("/executions", v1API.GetExecutions(database.Db))
		v1Group.GET("/transactions", v1API.GetTransactions(database.Db))
		v1Group.GET("/items/ids", v1API.GetItemIDs(database.Db))
		v1Group.GET("/items", v1API.GetItems(database.Db))
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
