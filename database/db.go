package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/nub-clubs-connect/nub_admin_api/config"
)

var DB *sql.DB

func Init() error {
	var err error
	DB, err = sql.Open("postgres", config.AppConfig.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	fmt.Println("✓ Database connection established successfully")
	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
