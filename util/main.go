package util

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/tavsec/gin-healthcheck/checks"
	"github.com/tavsec/gin-healthcheck/config"
)

var Db *sql.DB

func CheckConfig() config.Config {
	return config.Config{
		HealthPath:  "/v1/health",
		Method:      "GET",
		StatusOK:    200,
		StatusNotOK: 503,
		FailureNotification: struct {
			Threshold uint32
			Chan      chan error
		}{
			Threshold: 1,
		},
	}
}

func Check() []checks.Check {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error is occurred  on .env file please check")
	}

	dbHostCheck := checks.NewEnvCheck("PGSQL_HOST")
	dbPortCheck := checks.NewEnvCheck("PGSQL_PORT")
	dbUserCheck := checks.NewEnvCheck("PGSQL_USER")
	dbNameCheck := checks.NewEnvCheck("PGSQL_DB")
	dbPasswordCheck := checks.NewEnvCheck("PGSQL_PASSWORD")
	dbSslModeCheck := checks.NewEnvCheck("PGSQL_SSLMODE")

	psqlSetup := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		os.Getenv("PGSQL_HOST"),
		os.Getenv("PGSQL_PORT"),
		os.Getenv("PGSQL_USER"),
		os.Getenv("PGSQL_DB"),
		os.Getenv("PGSQL_PASSWORD"),
		os.Getenv("PGSQL_SSLMODE"),
	)

	db, _ := sql.Open("postgres", psqlSetup)

	return []checks.Check{
		checks.SqlCheck{Sql: db},
		dbHostCheck,
		dbPortCheck,
		dbUserCheck,
		dbNameCheck,
		dbPasswordCheck,
		dbSslModeCheck,
	}
}
