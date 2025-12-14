package models

import "time"

// RegistrationToken represents a one-time or time-limited token used to
// authorize agent/resource registration with the controller.
type RegistrationToken struct {
	ID         int64      `gorm:"primaryKey"`
	Token      string     `gorm:"size:128;index;not null"`
	ResourceID *int64     `gorm:"index;null"`
	Used       bool       `gorm:"default:false"`
	ExpiresAt  *time.Time `gorm:"index;null"`
	CreatedAt  time.Time
}
