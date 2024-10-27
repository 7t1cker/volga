package models

import (
	"time"

	"gorm.io/gorm"
)

type Account struct {
    ID        uint       `gorm:"primaryKey" json:"id"`
    LastName  string     `json:"lastName"`
    FirstName string     `json:"firstName"`
    Username  string     `gorm:"unique;not null" json:"username"`
    Password  string     `json:"-"`
    Roles     []*Role    `gorm:"many2many:account_roles;constraint:OnDelete:CASCADE;" json:"roles"`
    Specializations []*Specialization `gorm:"many2many:doctor_specializations;" json:"specializations,omitempty"`
    CreatedAt time.Time  `json:"-"`
    UpdatedAt time.Time  `json:"-"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
