package models

import "time"

type Role struct {
	ID          int64  `gorm:"primaryKey"`
	OrgID       int64  `gorm:"index"`
	Name        string `gorm:"size:200;not null"`
	Slug        string `gorm:"size:200;not null"`
	Description string
	IsSystem    bool `gorm:"default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}
