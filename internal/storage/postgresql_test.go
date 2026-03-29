package storage

import (
	"os"
	"testing"
)

func TestPostgreSQLConnection(t *testing.T) {
    if os.Getenv("INTEGRATION_TEST") == "" {
        t.Skip("skipping integration test, set INTEGRATION_TEST=1 to run")
    }

    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://localhost/sentineldb?sslmode=disable"
    }

    db, err := NewConnection(dbURL)
    if err != nil {
        t.Fatalf("expected connection to succeed, got: %v", err)
    }

    sqlDB, err := db.DB()
    if err != nil {
        t.Fatalf("expected to get sql.DB, got: %v", err)
    }
    defer sqlDB.Close()
}