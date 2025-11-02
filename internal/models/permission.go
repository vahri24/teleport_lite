package models

import "time"

type Permission struct {
    ID          uint64    `gorm:"primaryKey"`
    Key         string    `gorm:"uniqueIndex;size:200;not null"`
    Description string    `gorm:"size:255"`
    Resource    string    `gorm:"size:100"`
    Action      string    `gorm:"size:100"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
