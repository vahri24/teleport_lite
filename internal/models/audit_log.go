package models

import (
	"gorm.io/datatypes"
	"time"
)

type AuditLog struct {
	ID            int64          `gorm:"primaryKey"`
	OrgID         int64          `gorm:"index;not null"`
	UserID        int64          `gorm:"index"`             // nullable (system actions possible)
	Action        string         `gorm:"size:200;not null"` // e.g. "users.create", "servers.ssh"
	ResourceType  string         `gorm:"size:100"`          // e.g. "server", "user"
	ResourceID    int64          `gorm:"index"`             // optional link to resource
	Metadata      datatypes.JSON `gorm:"type:json"`         // details of what changed
	IP            string         `gorm:"size:64"`
	InitiatorName string         `gorm:"size:255" json:"initiator_name"`
	UserAgent     string         `gorm:"size:255"`
	CreatedAt     time.Time

	// Relationships
	Org  *Organization `gorm:"foreignKey:OrgID"`
	User *User         `gorm:"foreignKey:UserID"`
}
