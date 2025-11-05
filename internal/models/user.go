package models

import "time"

type UserStatus string
const (
    UserActive    UserStatus = "active"
    UserSuspended UserStatus = "suspended"
)

type User struct {
    ID              int64      `gorm:"primaryKey"`
    OrgID           int64      `gorm:"index"`
    Email           string     `gorm:"uniqueIndex;size:255;not null"`
    Name            string     `gorm:"size:200"`
    AuthProvider    string     `gorm:"size:20;default:local"`
    PasswordHash    string     `gorm:"size:255"`
    ConnectUser     string    `gorm:"size:255" json:"connect_user"`
    Status          UserStatus `gorm:"size:16;default:active"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
    Roles           []Role     `gorm:"many2many:user_roles;"`
}
