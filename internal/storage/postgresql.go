package storage

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	defaultMaxConns    = 5
	defaultMinConns    = 1
	defaultMaxConnLife = 30 * time.Minute
	defaultMaxConnIdle = 5 * time.Minute
)

func NewConnection(databaseURL string) (*gorm.DB, error) {
	configuredDatabaseURL, err := applySchemaToDatabaseURL(databaseURL, os.Getenv("DB_SCHEMA"))
	if err != nil {
		return nil, err
	}

	gormLogger := logger.New(
		log.New(os.Stdout, "", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(configuredDatabaseURL), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(int(defaultMaxConns))
	sqlDB.SetMaxIdleConns(int(defaultMinConns))
	sqlDB.SetConnMaxLifetime(defaultMaxConnLife)
	sqlDB.SetConnMaxIdleTime(defaultMaxConnIdle)

	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func applySchemaToDatabaseURL(databaseURL string, schema string) (string, error) {
	if schema == "" || strings.Contains(strings.ToLower(databaseURL), "search_path=") {
		return databaseURL, nil
	}

	if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
		parsedURL, err := url.Parse(databaseURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse database url: %w", err)
		}

		query := parsedURL.Query()
		query.Set("search_path", schema)
		parsedURL.RawQuery = query.Encode()

		return parsedURL.String(), nil
	}

	return fmt.Sprintf("%s search_path=%s", databaseURL, schema), nil
}
