package rbac

import (
    "context"
    "errors"
    "gorm.io/gorm"
    "strings"
)

type Checker struct { DB *gorm.DB }

func (c Checker) Can(ctx context.Context, userID, orgID uint64, permKey string) (bool, error) {
    // JOIN user_roles -> roles -> role_permissions -> permissions by `key`
    var count int64
    err := c.DB.
        Table("user_roles ur").
        Joins("JOIN roles r ON r.id = ur.role_id AND r.org_id = ?", orgID).
        Joins("JOIN role_permissions rp ON rp.role_id = r.id").
        Joins("JOIN permissions p ON p.id = rp.permission_id").
        Where("ur.user_id = ? AND ur.org_id = ? AND p.`key` = ?", userID, orgID, permKey).
        Count(&count).Error
    return count > 0, err
}

// Helper to compose like "users:read" from resource+action
func Key(resource, action string) string { return strings.ToLower(resource + ":" + action) }
