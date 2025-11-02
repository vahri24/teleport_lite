package models

import "time"

type UserStatus string
const (
    UserActive    UserStatus = "active"
    UserSuspended UserStatus = "suspended"
)

type User struct {
    ID           uint64     `gorm:"primaryKey"`
    OrgID        uint64     `gorm:"index"`
    Email        string     `gorm:"uniqueIndex;size:255;not null"`
    Name         string     `gorm:"size:200"`
    AuthProvider string     `gorm:"size:20;default:local"`
    PasswordHash string     `gorm:"size:255"`
    Status       UserStatus `gorm:"size:16;default:active"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
    Roles        []Role     `gorm:"many2many:user_roles;"`
}
