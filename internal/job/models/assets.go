package models

import "time"

type Asset struct {
	ID        string    `gorm:"primaryKey;type:text"`
	Type      string    `gorm:"not null;type:text"`
	Value     string    `gorm:"not null;type:text"`
	Label     *string   `gorm:"type:text"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}
