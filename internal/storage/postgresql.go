package storage

import (
	"fmt"
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
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

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
