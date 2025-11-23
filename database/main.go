package database

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/XSAM/otelsql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	logger "github.com/txlog/server/logger"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

var Db *sql.DB

//go:embed migrations/*
var migrationsFS embed.FS

// ConnectDatabase establishes a connection to the PostgreSQL database using environment
// variables for configuration. It performs the following steps:
//
// 1. Creates a database connection string using environment variables:
//   - PGSQL_HOST: Database host
//   - PGSQL_PORT: Database port
//   - PGSQL_USER: Database user
//   - PGSQL_DB: Database name
//   - PGSQL_PASSWORD: Database password
//   - PGSQL_SSLMODE: SSL mode for connection
//
// 2. Establishes connection to the database with OpenTelemetry instrumentation
//
// 3. Sets up database migrations:
//   - Creates a postgres driver instance
//   - Initializes migration source from embedded filesystem
//   - Applies pending migrations
//
// The function will panic if it fails to establish the database connection.
// It logs information about successful connection and migration application,
// as well as any errors that occur during the migration process.
//
// If OpenTelemetry is configured, all SQL queries will be automatically traced
// with detailed information including query text, execution time, and errors.
func ConnectDatabase() {
	psqlSetup := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("PGSQL_HOST"),
		os.Getenv("PGSQL_PORT"),
		os.Getenv("PGSQL_USER"),
		os.Getenv("PGSQL_DB"),
		os.Getenv("PGSQL_PASSWORD"),
		os.Getenv("PGSQL_SSLMODE"),
	)

	// Open database connection with OpenTelemetry instrumentation
	// This will automatically trace all SQL queries, transactions, and connection operations
	db, errSql := otelsql.Open("postgres", psqlSetup,
		otelsql.WithAttributes(
			semconv.DBSystemPostgreSQL,
			semconv.DBName(os.Getenv("PGSQL_DB")),
		),
		// SQL Commenter adds trace context as SQL comments for database-side correlation
		// Set to false to disable (query text is captured in spans by default)
		otelsql.WithSQLCommenter(false),
		// Trace all database calls including Ping, Exec, Query, etc.
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			Ping:                 true,
			RowsNext:             false, // Skip tracing individual row fetches to reduce noise
			DisableErrSkip:       false,
			OmitConnResetSession: true, // Skip tracing connection reset for less noise
		}),
	)
	if errSql != nil {
		logger.Error("There is an error while connecting to the database: " + errSql.Error())
		panic(errSql)
	}

	// Register database stats for monitoring
	if err := otelsql.RegisterDBStatsMetrics(db, otelsql.WithAttributes(
		semconv.DBSystemPostgreSQL,
		semconv.DBName(os.Getenv("PGSQL_DB")),
	)); err != nil {
		logger.Warn("Failed to register database metrics: " + err.Error())
		// Continue anyway - metrics are optional
	}

	Db = db
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		logger.Info("Database: connection established with OpenTelemetry instrumentation.")
	} else {
		logger.Info("Database: connection established.")
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		logger.Error("Failed to create database driver: " + err.Error())
		return
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		logger.Error("Failed to create migration source: " + err.Error())
		return
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		logger.Error("Failed to create migration instance: " + err.Error())
		return
	}

	// Check if database is in a dirty state
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		logger.Error("Failed to get migration version: " + err.Error())
	}

	// If database is dirty, try to force to the current version and retry
	if dirty {
		logger.Warn(fmt.Sprintf("Database is in dirty state at version %d. Attempting to fix...", version))
		if err := m.Force(int(version)); err != nil {
			logger.Error("Failed to force migration version: " + err.Error())
			logger.Error("Manual intervention required. Run: migrate force <version>")
			return
		}
		logger.Info(fmt.Sprintf("Forced database to clean state at version %d", version))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("Failed to apply migrations: " + err.Error())
		logger.Error("Migration may be incomplete. Check database state and consider manual migration.")
	} else if err == migrate.ErrNoChange {
		logger.Info("Migrations: no new migrations to apply.")
	} else {
		logger.Info("Migrations: successfully applied.")
	}
}
