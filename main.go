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
	"github.com/txlog/server/scheduler"
	"github.com/txlog/server/util"
)

// version of the application
var version = "1.1.1"

//go:embed assets
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
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.ConnectDatabase()
	scheduler.StartScheduler()

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(EnvironmentVariablesMiddleware())

	r.SetFuncMap(template.FuncMap{
		"iterate": func(start, count int) []int {
			var items []int
			for i := start; i <= count; i++ {
				items = append(items, i)
			}
			return items
		},
		"add": func(a, b int) int {
			return a + b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			} else {
				return b
			}
		},
	})

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

	r.NoRoute(controllers.Get404)

	r.GET("/", controllers.GetRootIndex(database.Db))
	r.GET("/executions", controllers.GetExecutionsIndex(database.Db))
	r.GET("/settings", controllers.GetSettingsIndex)
	r.GET("/license", controllers.GetLicensesIndex)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.PersistAuthorization(true),
		ginSwagger.DocExpansion("none"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	v1 := r.Group("/v1")
	{
		// txlog version
		v1.GET("/version", controllers.GetVersions(version))

		// txlog build
		v1.GET("/transactions/ids", controllers.GetTransactionIDs(database.Db))
		v1.POST("/transactions", controllers.PostTransactions(database.Db))
		v1.POST("/executions", controllers.PostExecutions(database.Db))

		// txlog machine_id \
		//   --hostname=G15.example.com
		v1.GET("/machines/ids", controllers.GetMachineIDs(database.Db))

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

	r.Run()
}

func EnvironmentVariablesMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		envVars := map[string]string{
			"INSTANCE": os.Getenv("INSTANCE"),
		}

		c.Set("env", envVars)

		c.Next()
	}
}
