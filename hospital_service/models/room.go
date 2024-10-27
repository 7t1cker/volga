package models

type Room struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Name       string `json:"name"`
	HospitalID uint   `json:"hospitalId"`
}
