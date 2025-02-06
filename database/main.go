package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var Db *sql.DB

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
		fmt.Println("There is an error while connecting to the database ", errSql)
		panic(errSql)
	} else {
		Db = db
		fmt.Println("Successfully connected to database!")
	}
}
