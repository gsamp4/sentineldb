package models

import (
	"encoding/json"
	"time"
)

type Finding struct {
	ID         string          `gorm:"primaryKey;type:text"`
	AssetID    string          `gorm:"not null;type:text"`
	RunID      string          `gorm:"not null;type:text"`
	Source     string          `gorm:"not null;type:text"`
	Severity   string          `gorm:"not null;type:text"`
	Title      string          `gorm:"not null;type:text"`
	Detail     json.RawMessage `gorm:"type:jsonb"`
	SeenAt     time.Time       `gorm:"not null;default:now()"`
	ResolvedAt *time.Time      `gorm:"type:timestamptz"`

	Asset Asset `gorm:"foreignKey:AssetID"`
	Run   Run   `gorm:"foreignKey:RunID"`
}
