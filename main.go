package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"

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
)

// version of the application
var version = "1.6.3"

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
		"formatPercentage": func(porcentagem float64) string {
			s := strconv.FormatFloat(porcentagem, 'f', 2, 64)
			s = strings.ReplaceAll(s, ".", ",")

			parts := strings.Split(s, ",")
			integerPart := parts[0]
			decimalPart := parts[1]

			isNegative := strings.HasPrefix(integerPart, "-")
			if isNegative {
				integerPart = integerPart[1:]
			}

			n := len(integerPart)
			if n <= 3 {
				if isNegative {
					return "-" + integerPart + "," + decimalPart
				}
				return integerPart + "," + decimalPart
			}

			var result string
			for i := 0; i < n; i++ {
				if (n-i)%3 == 0 && i != 0 {
					result += "."
				}
				result += string(integerPart[i])
			}
			if isNegative {
				return "-" + result + "," + decimalPart
			}
			return result + "," + decimalPart
		},
		"formatInteger": func(num int) string {
			s := strconv.Itoa(num)
			isNegative := strings.HasPrefix(s, "-")
			if isNegative {
				s = s[1:]
			}

			n := len(s)
			if n <= 3 {
				if isNegative {
					return "-" + s
				}
				return s
			}

			var result string
			for i := 0; i < n; i++ {
				if (n-i)%3 == 0 && i != 0 {
					result += "."
				}
				result += string(s[i])
			}
			if isNegative {
				return "-" + result
			}
			return result
		},
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
		"version": func() string { return version },
	}

	if os.Getenv("GIN_MODE") == "" {
		tmpl := template.Must(template.New("any").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))
		r.SetHTMLTemplate(tmpl)

		fsys, _ := fs.Sub(staticFiles, "assets")
		r.StaticFS("/assets", http.FS(fsys))
	} else {
		r.SetFuncMap(funcMap)
		r.LoadHTMLGlob("templates/*.html")
		r.Static("/assets", "./assets")
	}

	healthcheck.New(r, util.CheckConfig(), util.Check())

	r.NoRoute(controllers.Get404)

	r.GET("/", controllers.GetRootIndex(database.Db))
	r.GET("/assets", controllers.GetAssetsIndex(database.Db))
	r.GET("/executions/:execution_id", controllers.GetExecutionID(database.Db))
	r.GET("/insights", controllers.GetInsightsIndex)
	r.GET("/license", controllers.GetLicensesIndex)
	r.GET("/machines/:machine_id", controllers.GetMachineID(database.Db))
	r.GET("/settings", controllers.GetSettingsIndex)
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
		v1.GET("/version", controllers.GetVersions(version))

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
