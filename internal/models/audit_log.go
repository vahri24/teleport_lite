package models

import (
    "time"
    "gorm.io/datatypes"
)

type AuditLog struct {
    ID            uint64         `gorm:"primaryKey"`
    OrgID         uint64         `gorm:"index;not null"`
    UserID        *uint64        `gorm:"index"`             // nullable (system actions possible)
    Action        string         `gorm:"size:200;not null"` // e.g. "users.create", "servers.ssh"
    ResourceType  string         `gorm:"size:100"`          // e.g. "server", "user"
    ResourceID    *uint64        `gorm:"index"`             // optional link to resource
    Metadata      datatypes.JSON `gorm:"type:json"`         // details of what changed
    IP            string         `gorm:"size:64"`
    UserAgent     string         `gorm:"size:255"`
    CreatedAt     time.Time

    // Relationships
    Org  *Organization `gorm:"foreignKey:OrgID"`
    User *User         `gorm:"foreignKey:UserID"`
}
