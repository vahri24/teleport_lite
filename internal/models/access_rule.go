package models

import "time"

type AccessRule struct {
	ID             uint64 `gorm:"primaryKey"`
	OrgID          uint64 `gorm:"index;not null"`
	ResourceID     uint64 `gorm:"index;not null"`
	RoleID         uint64 `gorm:"index;not null"`
	PermissionID   uint64 `gorm:"index;not null"`
	ConstraintExpr string `gorm:"type:text"` // optional CEL or condition logic
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relationships
	Org        *Organization `gorm:"foreignKey:OrgID"`
	Resource   *Resource     `gorm:"foreignKey:ResourceID"`
	Role       *Role         `gorm:"foreignKey:RoleID"`
	Permission *Permission   `gorm:"foreignKey:PermissionID"`
}
