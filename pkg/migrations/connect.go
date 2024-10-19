package migrations

import (
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	DbEnvVar = "DATABASE_URL"
)

func ConnectDB() (*sqlx.DB, error) {
	// Get the connection string.
	connStr := getConnectionStr()
	// Open the database connection.
	db, err := sqlx.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	return db, nil
}

func getConnectionStr() string {
	// Get the connection string from the environment.
	connStr := os.Getenv(DbEnvVar)
	// Append "?timeout=90s&multiStatements=true&parseTime=true" to the connection string. But remove any current query string.
	if strings.Contains(connStr, "?") {
		connStr = strings.Split(connStr, "?")[0]
	}
	return fmt.Sprintf("%s?timeout=90s&multiStatements=true&parseTime=true", connStr)
}
