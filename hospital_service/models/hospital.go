package models

import (
	"time"

	"gorm.io/gorm"
)

type Hospital struct {
    ID           uint       `gorm:"primaryKey" json:"id"`
    Name         string     `gorm:"not null" json:"name"`
    Address      string     `json:"address"`
    ContactPhone string     `json:"contactPhone"`
    Rooms        []Room     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"rooms"`
    CreatedAt    time.Time  `json:"-"`
    UpdatedAt    time.Time  `json:"-"`
    DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}