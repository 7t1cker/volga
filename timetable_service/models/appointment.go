package models

import "time"

type Appointment struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    TimetableID uint      `json:"timetableId"`
    UserID      uint      `json:"userId"`
    Time        time.Time `json:"time"`
    CreatedAt   time.Time `json:"-"`
    UpdatedAt   time.Time `json:"-"`
    DeletedAt   *time.Time `gorm:"index" json:"-"`
}
