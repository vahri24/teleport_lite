package models

// UserRole represents the join between users and roles within an organization.
// The underlying `user_roles` table uses a composite primary key
// (user_id, role_id, org_id) and does not have a single `id` column.
type UserRole struct {
	UserID int64 `gorm:"primaryKey"`
	RoleID int64 `gorm:"primaryKey"`
	OrgID  int64 `gorm:"primaryKey"`
}
