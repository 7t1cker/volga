package models

import (
	"time"
)

	type Timetable struct {
		ID         uint          `gorm:"primaryKey" json:"id"`
		HospitalID uint          `json:"hospitalId"`
		DoctorID   uint          `json:"doctorId"`
		From       time.Time     `json:"from"`
		To         time.Time     `json:"to"`
		Room       string        `json:"room"`
		Appointments []Appointment `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
		CreatedAt  time.Time     `json:"-"`
		UpdatedAt  time.Time     `json:"-"`
		DeletedAt  *time.Time    `gorm:"index" json:"-"`
	}
