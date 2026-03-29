package models

import "time"

type Run struct {
	ID         string     `gorm:"primaryKey;type:text"`
	CreatedAt  time.Time  `gorm:"not null;default:now()"`
	Status     string     `gorm:"not null;type:text;default:pending"`
	Error      *string    `gorm:"type:text"`
	StartedAt  *time.Time `gorm:"type:timestamptz"`
	FinishedAt *time.Time `gorm:"type:timestamptz"`
}
