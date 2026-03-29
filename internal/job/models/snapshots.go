package models

import (
	"encoding/json"
	"time"
)

type AssetSnapshot struct {
	ID         string          `gorm:"primaryKey;type:text"`
	AssetID    string          `gorm:"not null;type:text"`
	RunID      string          `gorm:"not null;type:text"`
	Source     string          `gorm:"not null;type:text"`
	Data       json.RawMessage `gorm:"not null;type:jsonb"`
	SnapshotAt time.Time       `gorm:"not null;default:now()"`

	Asset Asset `gorm:"foreignKey:AssetID"`
	Run   Run   `gorm:"foreignKey:RunID"`
}
