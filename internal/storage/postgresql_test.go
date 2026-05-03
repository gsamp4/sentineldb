package storage

import (
	"os"
	"strings"
	"testing"
)

func TestApplySchemaToDatabaseURL(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		schema      string
		want        string
	}{
		{
			name:        "keeps url unchanged when schema is empty",
			databaseURL: "postgres://localhost/sentineldb?sslmode=disable",
			schema:      "",
			want:        "postgres://localhost/sentineldb?sslmode=disable",
		},
		{
			name:        "adds search path to postgres url",
			databaseURL: "postgres://localhost/sentineldb?sslmode=disable",
			schema:      "tenant_a",
			want:        "search_path=tenant_a",
		},
		{
			name:        "keeps existing search path",
			databaseURL: "postgres://localhost/sentineldb?search_path=custom&sslmode=disable",
			schema:      "tenant_a",
			want:        "postgres://localhost/sentineldb?search_path=custom&sslmode=disable",
		},
		{
			name:        "adds search path to keyword dsn",
			databaseURL: "host=localhost dbname=sentineldb sslmode=disable",
			schema:      "tenant_a",
			want:        "host=localhost dbname=sentineldb sslmode=disable search_path=tenant_a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := applySchemaToDatabaseURL(tt.databaseURL, tt.schema)
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if strings.Contains(tt.want, "search_path=") && strings.HasPrefix(tt.databaseURL, "postgres://") && tt.schema != "" && !strings.Contains(tt.databaseURL, "search_path=") {
				if !strings.Contains(got, tt.want) {
					t.Fatalf("expected %q to contain %q", got, tt.want)
				}
				return
			}

			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

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
