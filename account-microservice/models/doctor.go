package models

type Doctor struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	LastName       string `json:"lastName"`
	FirstName      string `json:"firstName"`
	Specialization string `json:"specialization"`
}
