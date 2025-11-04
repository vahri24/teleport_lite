package models

import (
    "time"
    "gorm.io/datatypes"
)

type Resource struct {
    ID          int64          `gorm:"primaryKey"`
    OrgID       int64          `gorm:"index;not null"`
    Name        string         `gorm:"size:200;not null"`
    Type        string         `gorm:"size:100;not null"`
    ExternalRef string         `gorm:"size:255"`
    Host        string         `gorm:"size:100;not null" json:"Host"`
    Metadata    datatypes.JSON `gorm:"type:json"`
    Status      string         `gorm:"size:50" json:"Status"`
    CreatedAt   time.Time
    UpdatedAt   time.Time

    Org         *Organization `gorm:"foreignKey:OrgID"`
    AccessRules []AccessRule  `gorm:"foreignKey:ResourceID"`
}
