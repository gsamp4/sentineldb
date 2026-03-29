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
		dbURL = "postgres://localhost/xpolynews?sslmode=disable"
	}
	_, err := NewConnection(dbURL)
	if err != nil {
		t.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
}