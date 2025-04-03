package config

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Siddharth9890/osquery-mvp/internal/database"
	_ "github.com/go-sql-driver/mysql"
)

func NewDatabaseConnection(dsn string) (*sql.DB, error) {
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(time.Hour)

	return conn, nil
}

func NewDatabaseService(conn *sql.DB) *database.Service {
	return database.NewService(conn)
}
