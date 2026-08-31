package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
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
// 2. Establishes connection to the database and configures connection pool
//
// 3. Sets up database migrations:
//   - Creates a postgres driver instance
//   - Initializes migration source from embedded filesystem
//   - Applies pending migrations
//
// The function will panic if it fails to establish the database connection.
// It logs information about successful connection and migration application,
// as well as any errors that occur during the migration process.
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

	db, errSql := sql.Open("postgres", psqlSetup)
	if errSql != nil {
		slog.Error("There is an error while connecting to the database: " + errSql.Error())
		panic(errSql)
	} else {
		Db = db
		slog.Info("Database: connection established.")
	}

	// Configure connection pool to prevent unbounded connection growth
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)
	slog.Info("Database: connection pool configured (max_open=25, max_idle=10).")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Failed to create database driver: " + err.Error())
		return
	}

	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		slog.Error("Failed to create migration source: " + err.Error())
		return
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		slog.Error("Failed to create migration instance: " + err.Error())
		return
	}

	// Check if database is in a dirty state
	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		slog.Error("Failed to get migration version: " + err.Error())
	}

	// If database is dirty, try to force to the current version and retry
	if dirty {
		slog.Warn(fmt.Sprintf("Database is in dirty state at version %d. Attempting to fix...", version))
		if err := m.Force(int(version)); err != nil {
			slog.Error("Failed to force migration version: " + err.Error())
			slog.Error("Manual intervention required. Run: migrate force <version>")
			return
		}
		slog.Info(fmt.Sprintf("Forced database to clean state at version %d", version))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("Failed to apply migrations: " + err.Error())
		slog.Error("Migration may be incomplete. Check database state and consider manual migration.")
	} else if err == migrate.ErrNoChange {
		slog.Info("Migrations: no new migrations to apply.")
	} else {
		slog.Info("Migrations: successfully applied.")
	}
}
