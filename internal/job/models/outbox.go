package models

import "time"

type Outbox struct {
	ID          string    `gorm:"primaryKey;type:text"`
	RunID       string    `gorm:"not null;type:text"`
	AssetID     string    `gorm:"not null;type:text"`
	JobType     string    `gorm:"not null;type:text"`
	Status      string    `gorm:"not null;type:text;default:pending"`
	Attempts    int       `gorm:"not null;default:0"`
	MaxAttempts int       `gorm:"not null;default:3"`
	ScheduledAt time.Time `gorm:"not null;default:now()"`
	UpdatedAt   time.Time `gorm:"not null;default:now()"`
	FinishedAt  *time.Time `gorm:"type:timestamptz"`

	Run   Run   `gorm:"foreignKey:RunID"`
	Asset Asset `gorm:"foreignKey:AssetID"`
}
