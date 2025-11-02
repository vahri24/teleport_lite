package models

import "time"

type Organization struct {
    ID        uint64    `gorm:"primaryKey"`
    Name      string    `gorm:"size:200;not null"`
    Slug      string    `gorm:"size:200;uniqueIndex;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time

    Users    []User
    Roles    []Role
    Resources []Resource
}
