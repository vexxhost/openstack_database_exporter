package db

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// Connect establishes a database connection from an oslo.db-style MySQL URL.
//
// Supported input formats:
//   - mysql://user:password@host:port/database
//   - mysql+pymysql://user:password@host:port/database
//   - mysql+mysqldb://user:password@host:port/database
//   - mysql+mysqlconnector://user:password@host:port/database
func Connect(connectionString string) (*sql.DB, error) {
	dsn, err := ParseOsloDBConnectionString(connectionString)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
