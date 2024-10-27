package models

import (
	"time"
)

type History struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    Date       time.Time `json:"date"`
    PacientID  uint      `json:"pacientId"`
    HospitalID uint      `json:"hospitalId"`
    DoctorID   uint      `json:"doctorId"`
    Room       string    `json:"room"`
    Data       string    `json:"data"`
    CreatedAt  time.Time `json:"-"`
    UpdatedAt  time.Time `json:"-"`
    DeletedAt  *time.Time `gorm:"index" json:"-"`
}
