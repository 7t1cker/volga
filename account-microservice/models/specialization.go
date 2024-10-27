package models

type Specialization struct {
	ID      uint       `gorm:"primaryKey" json:"id"`
	Name    string     `gorm:"unique;not null" json:"name"`
	Doctors []*Account `gorm:"many2many:doctor_specializations;" json:"-"`
}
