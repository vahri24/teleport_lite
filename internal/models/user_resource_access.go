package models

import "time"

// UserResourceAccess maps a user to a resource with an optional SSH username
type UserResourceAccess struct {
	ID          uint64 `gorm:"primaryKey"`
	OrgID       uint64 `gorm:"index;not null"`
	UserID      int64  `gorm:"index;not null"`
	ResourceID  uint64 `gorm:"index;not null"`
	ConnectUser string `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	User     *User     `gorm:"foreignKey:UserID"`
	Resource *Resource `gorm:"foreignKey:ResourceID"`
}
