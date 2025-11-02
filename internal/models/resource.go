package models

import (
    "time"
    "gorm.io/datatypes"
)

type Resource struct {
    ID          uint64         `gorm:"primaryKey"`
    OrgID       uint64         `gorm:"index;not null"`                 // organization scope
    Name        string         `gorm:"size:200;not null"`              // human-friendly name
    Type        string         `gorm:"size:100;not null"`              // e.g. "server", "cluster"
    ExternalRef string         `gorm:"size:255"`                       // optional: hostname, UUID, etc.
    Metadata    datatypes.JSON `gorm:"type:json"`                      // arbitrary key/value pairs
    CreatedAt   time.Time
    UpdatedAt   time.Time

    // Relationships
    Org         *Organization  `gorm:"foreignKey:OrgID"`
    AccessRules []AccessRule   `gorm:"foreignKey:ResourceID"`
}
