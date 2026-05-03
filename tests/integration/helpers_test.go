package integration_test

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"sentineldb/internal/job/models"
	"sentineldb/internal/storage"
	"sentineldb/pkg/logger"

	"gorm.io/gorm"
)

func newIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()

	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("skipping integration test, set INTEGRATION_TEST=1 to run")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://localhost/sentineldb?sslmode=disable"
	}

	schemaName := fmt.Sprintf("integration_%d", time.Now().UnixNano())
	adminDB, err := storage.NewConnection(dbURL)
	if err != nil {
		t.Fatalf("failed to connect to integration database: %v", err)
	}

	if err := adminDB.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName)).Error; err != nil {
		t.Fatalf("failed to create integration schema %s: %v", schemaName, err)
	}

	adminSQLDB, err := adminDB.DB()
	if err != nil {
		t.Fatalf("failed to get admin sql.DB: %v", err)
	}

	if err := adminSQLDB.Close(); err != nil {
		t.Fatalf("failed to close admin database connection: %v", err)
	}

	scopedDBURL, err := databaseURLWithSchema(dbURL, schemaName)
	if err != nil {
		t.Fatalf("failed to scope integration database url to schema %s: %v", schemaName, err)
	}

	db, err := storage.NewConnection(scopedDBURL)
	if err != nil {
		t.Fatalf("failed to connect to integration schema %s: %v", schemaName, err)
	}

	if err := db.AutoMigrate(
		&models.Asset{},
		&models.Run{},
		&models.Outbox{},
		&models.AssetSnapshot{},
		&models.Finding{},
	); err != nil {
		t.Fatalf("failed to migrate integration tables: %v", err)
	}

	cleanupIntegrationTables(t, db)

	t.Cleanup(func() {
		cleanupIntegrationTables(t, db)

		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("failed to get sql.DB during cleanup: %v", err)
		}

		if err := sqlDB.Close(); err != nil {
			t.Fatalf("failed to close integration database: %v", err)
		}

		adminDB, err := storage.NewConnection(dbURL)
		if err != nil {
			t.Fatalf("failed to reconnect for schema cleanup: %v", err)
		}

		if err := adminDB.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schemaName)).Error; err != nil {
			t.Fatalf("failed to drop integration schema %s: %v", schemaName, err)
		}

		adminSQLDB, err := adminDB.DB()
		if err != nil {
			t.Fatalf("failed to get cleanup admin sql.DB: %v", err)
		}

		if err := adminSQLDB.Close(); err != nil {
			t.Fatalf("failed to close cleanup admin connection: %v", err)
		}
	})

	return db
}

func cleanupIntegrationTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec("TRUNCATE TABLE findings, asset_snapshots, outboxes, runs, assets RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("failed to cleanup integration tables: %v", err)
	}
}

func newTestLogger() *logger.Logger {
	return logger.New(logger.Options{
		Level:  logger.LevelDebug,
		Prefix: "TEST: ",
	})
}

func mustExec(t *testing.T, db *gorm.DB, query string, args ...interface{}) {
	t.Helper()

	if err := db.Exec(query, args...).Error; err != nil {
		t.Fatalf("failed to execute query %q: %v", query, err)
	}
}

func databaseURLWithSchema(databaseURL string, schema string) (string, error) {
	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		parsedURL, err := url.Parse(databaseURL)
		if err != nil {
			return "", err
		}

		query := parsedURL.Query()
		query.Set("search_path", schema)
		parsedURL.RawQuery = query.Encode()

		return parsedURL.String(), nil
	}

	return fmt.Sprintf("%s search_path=%s", databaseURL, schema), nil
}

func seedAsset(t *testing.T, db *gorm.DB, id, assetType, value string) models.Asset {
	t.Helper()

	asset := models.Asset{
		ID:        id,
		Type:      assetType,
		Value:     value,
		Active:    true,
		CreatedAt: time.Now(),
	}

	if err := db.Create(&asset).Error; err != nil {
		t.Fatalf("failed to seed asset: %v", err)
	}

	return asset
}

func seedRun(t *testing.T, db *gorm.DB, id string) models.Run {
	t.Helper()

	run := models.Run{
		ID:        id,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := db.Create(&run).Error; err != nil {
		t.Fatalf("failed to seed run: %v", err)
	}

	return run
}

func seedOutbox(t *testing.T, db *gorm.DB, job models.Outbox) models.Outbox {
	t.Helper()

	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("failed to seed outbox job: %v", err)
	}

	return job
}
