package models

import "time"

type Organization struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"size:200;not null"`
	Slug      string `gorm:"size:200;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Users     []User     `gorm:"foreignKey:OrgID"`
	Roles     []Role     `gorm:"foreignKey:OrgID"`
	Resources []Resource `gorm:"foreignKey:OrgID"`
}
