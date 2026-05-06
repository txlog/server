package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	healthcheck "github.com/tavsec/gin-healthcheck"
	"github.com/txlog/server/auth"
	"github.com/txlog/server/controllers"
	v1API "github.com/txlog/server/controllers/api/v1"
	"github.com/txlog/server/database"
	_ "github.com/txlog/server/docs"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/middleware"
	"github.com/txlog/server/models"
	"github.com/txlog/server/scheduler"
	"github.com/txlog/server/util"
	"github.com/txlog/server/version"
)

//go:embed images
var staticFiles embed.FS

//go:embed static
var staticCSSFiles embed.FS

//go:embed templates/*
var templateFS embed.FS

// @title						Txlog Server
// @version					v1
// @description				The centralized system that stores transaction data
// @termsOfService				https://github.com/txlog
// @contact.name				Txlog repository issues
// @contact.url				https://github.com/txlog/server/issues
// @license.name				MIT License
// @license.url				https://github.com/txlog/.github/blob/main/profile/LICENSE.md
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						X-API-Key
// @description				API key authentication for /v1 endpoints. Generate your API key in the admin panel at /admin
func main() {
	logger.InitLogger()

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.ConnectDatabase()

	// Sync topology pattern regular expressions
	tm := models.NewTopologyManager(database.Db)
	if count, err := tm.SyncCompiledPatterns(); err != nil {
		logger.Error("Failed to sync compiled topology patterns at startup: " + err.Error())
	} else if count > 0 {
		logger.Info(fmt.Sprintf("Topology: %d patterns synchronized with the current template engine.", count))
	} else {
		logger.Info("Topology: all patterns are up to date.")
	}

	scheduler.StartScheduler(database.Db)

	// Inject the background task trigger into controllers safely without direct package cycle
	controllers.SetSchedulerOSVTrigger(func() { scheduler.UpdateVulnerabilitiesJob(database.Db) })

	// Initialize OIDC service (optional)
	var oidcService *auth.OIDCService
	oidcService, err := auth.NewOIDCService(database.Db)
	if err != nil {
		logger.Error("Failed to initialize OIDC service: " + err.Error())
		os.Exit(1)
	}

	// Initialize LDAP service (optional)
	var ldapService *auth.LDAPService
	ldapService, err = auth.NewLDAPService(database.Db)
	if err != nil {
		logger.Error("Failed to initialize LDAP service: " + err.Error())
		os.Exit(1)
	}

	// Log authentication status
	if oidcService != nil {
		logger.Info("OIDC authentication enabled")
	}
	if ldapService != nil {
		logger.Info("LDAP authentication enabled")
	}
	if oidcService == nil && ldapService == nil {
		logger.Info("No authentication configured - API endpoints accessible without API key")
	} else {
		logger.Info("API key authentication required for /v1 endpoints")
	}

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(func(c *gin.Context) {
		c.SetSameSite(http.SameSiteLaxMode)
		c.Next()
	})
	r.Use(EnvironmentVariablesMiddleware())
	r.Use(middleware.AuthMiddleware(database.Db))

	funcMap := template.FuncMap{
		"add":              util.Add,
		"brand":            util.Brand,
		"derefBool":        util.DerefBool,
		"dnfUser":          util.DnfUser,
		"formatInteger":    util.FormatInteger,
		"formatPercentage": util.FormatPercentage,
		"formatDateTime":   util.FormatDateTime,
		"formatDate":       util.FormatDate,
		"hasAction":        util.HasAction,
		"hasPrefix":        util.HasPrefix,
		"initial":          util.Initial,
		"iterate":          util.Iterate,
		"maskString":       util.MaskString,
		"min":              util.Min,
		"text2html":        util.Text2HTML,
		"timeStatusClass":  util.TimeStatusClass,
		"trimPrefix":       util.TrimPrefix,
		"version":          util.Version,
		"versionsEqual":    util.VersionsEqual,
	}

	if os.Getenv("GIN_MODE") == "" {
		tmpl := template.Must(template.New("any").Funcs(funcMap).ParseFS(templateFS, "templates/*.html"))
		r.SetHTMLTemplate(tmpl)

		fsys, _ := fs.Sub(staticFiles, "images")
		r.StaticFS("/images", http.FS(fsys))

		cssFsys, _ := fs.Sub(staticCSSFiles, "static/css")
		r.StaticFS("/css", http.FS(cssFsys))
	} else {
		r.SetFuncMap(funcMap)
		r.LoadHTMLGlob("templates/*.html")
		r.Static("/images", "./images")
		r.Static("/css", "./static/css")
	}

	healthcheck.New(r, util.CheckConfig(), util.Check(database.Db))

	r.NoRoute(controllers.Get404)

	// Authentication routes (if OIDC or LDAP is configured)
	if oidcService != nil || ldapService != nil {
		r.GET("/login", controllers.GetLogin(oidcService, ldapService))
		r.POST("/auth/logout", controllers.PostLogout(oidcService, ldapService))
	}

	// OIDC-specific routes
	if oidcService != nil {
		r.POST("/auth/login", controllers.PostLogin(oidcService))
		r.GET("/auth/callback", controllers.GetCallback(oidcService))
	}

	// LDAP-specific routes
	if ldapService != nil {
		r.POST("/auth/ldap/login", controllers.PostLDAPLogin(ldapService))
	}

	// Main application routes
	r.GET("/", controllers.GetRootIndex(database.Db))
	r.GET("/assets", controllers.GetAssetsIndex(database.Db))
	r.GET("/packages", controllers.GetPackagesIndex(database.Db))
	r.GET("/topology", controllers.GetTopologyIndex(database.Db))

	// Admin routes (requires admin middleware)
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AdminMiddleware())
	{
		adminGroup.GET("", controllers.GetAdminIndex(database.Db))
		adminGroup.POST("/migrations/run", controllers.PostAdminRunMigrations(database.Db))
		adminGroup.POST("/migrations/run_osv_update", controllers.PostAdminRunOSVUpdate(database.Db))
		adminGroup.POST("/migrations/reset_osv", controllers.PostAdminResetOSV(database.Db))
		adminGroup.POST("/cleanup/inactive-assets", controllers.PostAdminCleanupInactiveAssets(database.Db))
		adminGroup.DELETE("/assets/:machine_id", controllers.DeleteMachineID(database.Db))

		// Topology configuration routes
		adminGroup.GET("/topology/preview", controllers.GetAdminTopologyPreview(database.Db))
		adminGroup.GET("/topology/preview-env", controllers.GetAdminTopologyPreviewEnv(database.Db))
		adminGroup.GET("/topology/preview-svc", controllers.GetAdminTopologyPreviewSvc(database.Db))
		adminGroup.POST("/topology/patterns", controllers.PostAdminTopologyCreatePattern(database.Db))
		adminGroup.POST("/topology/patterns/update", controllers.PostAdminTopologyUpdatePattern(database.Db))
		adminGroup.POST("/topology/patterns/delete", controllers.PostAdminTopologyDeletePattern(database.Db))
		adminGroup.POST("/topology/environments", controllers.PostAdminTopologyCreateEnvironment(database.Db))
		adminGroup.POST("/topology/environments/update", controllers.PostAdminTopologyUpdateEnvironment(database.Db))
		adminGroup.POST("/topology/environments/delete", controllers.PostAdminTopologyDeleteEnvironment(database.Db))
		adminGroup.POST("/topology/services", controllers.PostAdminTopologyCreateService(database.Db))
		adminGroup.POST("/topology/services/update", controllers.PostAdminTopologyUpdateService(database.Db))
		adminGroup.POST("/topology/services/delete", controllers.PostAdminTopologyDeleteService(database.Db))
	}

	// Admin routes that require OIDC or LDAP (user and API key management)
	if oidcService != nil || ldapService != nil {
		adminAuthGroup := r.Group("/admin")
		adminAuthGroup.Use(middleware.AdminMiddleware())
		{
			adminAuthGroup.POST("/update", controllers.PostAdminUpdateUser(database.Db))
			adminAuthGroup.POST("/delete", controllers.PostAdminDeleteUser(database.Db))
			adminAuthGroup.POST("/apikeys/create", controllers.PostAdminCreateAPIKey(database.Db))
			adminAuthGroup.POST("/apikeys/revoke", controllers.PostAdminRevokeAPIKey(database.Db))
			adminAuthGroup.POST("/apikeys/delete", controllers.DeleteAdminAPIKey(database.Db))
		}
	}
	r.GET("/assets/:machine_id", controllers.GetMachineID(database.Db))
	r.GET("/executions/:execution_id", controllers.GetExecutionID(database.Db))
	r.GET("/insights", controllers.GetInsightsIndex)
	r.GET("/license", controllers.GetLicensesIndex)
	r.GET("/analytics/progression", controllers.GetPackagesByWeekIndex(database.Db))
	r.GET("/api/packages-by-month", controllers.GetPackagesByMonth(database.Db))
	r.GET("/packages/:name", controllers.GetPackageByName(database.Db))

	// Analytics pages
	r.GET("/analytics/anomalies", controllers.GetAnalyticsAnomalies(database.Db))
	r.GET("/analytics/security", controllers.GetAnalyticsSecurity(database.Db))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(
		swaggerfiles.Handler,
		ginSwagger.PersistAuthorization(true),
		ginSwagger.DocExpansion("none"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	))

	v1Group := r.Group("/v1")

	// Only require API key when authentication is enabled (OIDC or LDAP)
	if oidcService != nil || ldapService != nil {
		v1Group.Use(middleware.APIKeyMiddleware(database.Db))
	}
	{
		// txlog version
		v1Group.GET("/version", v1API.GetVersions(version.SemVer))

		// txlog build
		v1Group.GET("/transactions/ids", v1API.GetTransactionIDs(database.Db))
		v1Group.POST("/transactions", v1API.PostTransactions(database.Db))
		v1Group.POST("/executions", v1API.PostExecutions(database.Db))

		// Assets requiring restart
		v1Group.GET("/assets/requiring-restart", v1API.GetAssetsRequiringRestart(database.Db))

		// Package listing
		v1Group.GET("/packages/:name/:version/:release/assets", v1API.GetAssetsUsingPackageVersion(database.Db))

		// Reports endpoints
		v1Group.GET("/reports/monthly", v1API.GetMonthlyReport(database.Db))
		v1Group.GET("/reports/anomalies", v1API.GetAnomalies(database.Db))
		v1Group.GET("/reports/fixed-vulnerabilities", v1API.GetFixedVulnerabilities(database.Db))

		// Endpoints for agent pre-v1.6.0
		v1Group.GET("/machines/ids", v1API.GetMachineIDs(database.Db))
		v1Group.GET("/machines", v1API.GetMachines(database.Db))
		v1Group.GET("/executions", v1API.GetExecutions(database.Db))
		v1Group.GET("/transactions", v1API.GetTransactions(database.Db))
		v1Group.GET("/items/ids", v1API.GetItemIDs(database.Db))
		v1Group.GET("/items", v1API.GetItems(database.Db))
		v1Group.GET("/vulnerabilities", v1API.GetTransactionVulnerabilities(database.Db))
	}

	r.Run()
}

func EnvironmentVariablesMiddleware() gin.HandlerFunc {
	// Snapshot env vars once at middleware creation, not per-request
	// This serves as a template for per-request maps
	staticEnvVars := map[string]string{
		"instance":                 os.Getenv("INSTANCE"),
		"logLevel":                 os.Getenv("LOG_LEVEL"),
		"ginMode":                  os.Getenv("GIN_MODE"),
		"port":                     os.Getenv("PORT"),
		"pgsqlHost":                os.Getenv("PGSQL_HOST"),
		"pgsqlPort":                os.Getenv("PGSQL_PORT"),
		"pgsqlUser":                os.Getenv("PGSQL_USER"),
		"pgsqlDb":                  os.Getenv("PGSQL_DB"),
		"pgsqlPassword":            util.MaskString(os.Getenv("PGSQL_PASSWORD")),
		"pgsqlSslmode":             os.Getenv("PGSQL_SSLMODE"),
		"cronRetentionDays":        os.Getenv("CRON_RETENTION_DAYS"),
		"cronRetentionExpression":  os.Getenv("CRON_RETENTION_EXPRESSION"),
		"cronStatisticsExpression": os.Getenv("CRON_STATS_EXPRESSION"),
		"cronOsvExpression":        os.Getenv("CRON_OSV_EXPRESSION"),
		"oidcIssuerUrl":            os.Getenv("OIDC_ISSUER_URL"),
		"oidcClientId":             os.Getenv("OIDC_CLIENT_ID"),
		"oidcClientSecret":         util.MaskString(os.Getenv("OIDC_CLIENT_SECRET")),
		"oidcRedirectUrl":          os.Getenv("OIDC_REDIRECT_URL"),
		"ldapHost":                 os.Getenv("LDAP_HOST"),
		"ldapPort":                 os.Getenv("LDAP_PORT"),
		"ldapUseTls":               os.Getenv("LDAP_USE_TLS"),
		"ldapBindDn":               os.Getenv("LDAP_BIND_DN"),
		"ldapBindPassword":         util.MaskString(os.Getenv("LDAP_BIND_PASSWORD")),
		"ldapBaseDn":               os.Getenv("LDAP_BASE_DN"),
		"ldapUserFilter":           os.Getenv("LDAP_USER_FILTER"),
		"ldapAdminGroup":           os.Getenv("LDAP_ADMIN_GROUP"),
		"ldapViewerGroup":          os.Getenv("LDAP_VIEWER_GROUP"),
		"ldapGroupFilter":          os.Getenv("LDAP_GROUP_FILTER"),
	}

	return func(c *gin.Context) {
		// Create a copy of the map for this request to avoid concurrent map writes
		envVars := make(map[string]string, len(staticEnvVars)+1)
		for k, v := range staticEnvVars {
			envVars[k] = v
		}

		// latestVersion is dynamic (updated by scheduler), read per-request
		envVars["latestVersion"] = os.Getenv("LATEST_VERSION")
		c.Set("env", envVars)

		// Add user to template context if available
		if userInterface, exists := c.Get("user"); exists {
			c.Set("user", userInterface)
		}

		// Add OIDC status to template context
		c.Set("oidc_enabled", auth.IsConfigured())
		c.Set("ldap_enabled", auth.IsLDAPConfigured())

		c.Next()
	}
}
