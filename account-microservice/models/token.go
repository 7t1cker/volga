package models

import "time"

type Token struct {
    ID        uint      `gorm:"primaryKey" json:"-"`
    Token     string    `gorm:"unique;not null" json:"token"`
    AccountID uint      `json:"account_id"`
    ExpiresAt time.Time `json:"expires_at"`
}
