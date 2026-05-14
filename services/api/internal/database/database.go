package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// BuildDSN constructs a PostgreSQL connection string from the given parameters.
// Empty parameters are replaced with sensible defaults.
func BuildDSN(host, port, dbname, user, password, sslmode string) string {
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if dbname == "" {
		dbname = "napkin_notes"
	}
	if user == "" {
		user = "postgres"
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
		host, port, dbname, user, password, sslmode)
}

// Connect opens a connection to the database using the provided DSN
// and verifies it with a ping.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
