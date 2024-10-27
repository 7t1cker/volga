package models

type Role struct {
	ID       uint       `gorm:"primaryKey" json:"-"`
	Name     string     `gorm:"unique;not null" json:"name"`
	Accounts []*Account `gorm:"many2many:account_roles;" json:"-"`
}
