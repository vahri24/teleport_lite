package seed

import (
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"teleport_lite/internal/models"
)

func FirstSetup(db *gorm.DB) error {
	// -------------------------
	// 1) Ensure default org
	// -------------------------
	org := models.Organization{Name: "Default Organization", Slug: "default"}
	if err := db.Where("slug = ?", org.Slug).FirstOrCreate(&org).Error; err != nil {
		return err
	}

	// -------------------------
	// 2) Ensure roles
	// -------------------------
	adminRole := models.Role{OrgID: org.ID, Name: "Administrator", Slug: "admin"}
	devopsRole := models.Role{OrgID: org.ID, Name: "DevOps", Slug: "devops"}
	readonlyRole := models.Role{OrgID: org.ID, Name: "ReadOnly", Slug: "readonly"}

	if err := db.Where("org_id=? AND slug=?", org.ID, adminRole.Slug).FirstOrCreate(&adminRole).Error; err != nil {
		return err
	}
	if err := db.Where("org_id=? AND slug=?", org.ID, devopsRole.Slug).FirstOrCreate(&devopsRole).Error; err != nil {
		return err
	}
	if err := db.Where("org_id=? AND slug=?", org.ID, readonlyRole.Slug).FirstOrCreate(&readonlyRole).Error; err != nil {
		return err
	}

	// -------------------------
	// 3) Ensure permissions
	// -------------------------
	perms := []models.Permission{
		{Key: "users:read", Description: "View users", Resource: "users", Action: "read"},
		{Key: "users:write", Description: "Manage users", Resource: "users", Action: "write"},
		{Key: "users:assign-role", Description: "Assign roles to users", Resource: "users", Action: "assign-role"},
		{Key: "roles:read", Description: "View roles", Resource: "roles", Action: "read"},
		{Key: "roles:write", Description: "Manage roles", Resource: "roles", Action: "write"},
		{Key: "resources:read", Description: "View resources", Resource: "resources", Action: "read"},
		{Key: "resources:generate-token", Description: "Generate registration tokens", Resource: "resources", Action: "generate-token"},
		{Key: "resources:write", Description: "Manage resources", Resource: "resources", Action: "write"},
		{Key: "audit:read", Description: "View audit logs", Resource: "audit", Action: "read"},
	}

	permIDs := map[string]uint64{}

	for _, p := range perms {
		tmp := p
		if err := db.Where("`key` = ?", tmp.Key).FirstOrCreate(&tmp).Error; err != nil {
			return err
		}
		permIDs[tmp.Key] = tmp.ID
	}

	// -------------------------
	// 4) role_permissions mapping
	// -------------------------
	// Helper: ensure mapping exists. Use a direct INSERT IGNORE into the
	// `role_permissions` join table to avoid GORM's "model value required"
	// error when operating on a table without a corresponding model.
	ensureRolePerm := func(roleID int64, permID uint64) error {
		// Use INSERT IGNORE so duplicate pairs do not produce an error.
		res := db.Exec("INSERT IGNORE INTO role_permissions (role_id, permission_id) VALUES (?, ?)", roleID, permID)
		return res.Error
	}

	// Admin gets ALL permissions
	for _, pid := range permIDs {
		if err := ensureRolePerm(adminRole.ID, pid); err != nil {
			return err
		}
	}

	// DevOps: manage resources + read audit + read roles/users
	devopsKeys := []string{"resources:read", "resources:generate-token", "resources:write", "audit:read", "roles:read", "users:read"}
	for _, k := range devopsKeys {
		if err := ensureRolePerm(devopsRole.ID, permIDs[k]); err != nil {
			return err
		}
	}

	// ReadOnly: read-only permissions
	readonlyKeys := []string{"users:read", "roles:read", "resources:read", "audit:read"}
	for _, k := range readonlyKeys {
		if err := ensureRolePerm(readonlyRole.ID, permIDs[k]); err != nil {
			return err
		}
	}

	// -------------------------
	// 5) Ensure admin user
	// -------------------------
	const adminEmail = "admin@example.com"
	const adminPass = "admin123" // change after first login

	passHash, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)

	adminUser := models.User{
		OrgID:        org.ID,
		Email:        adminEmail,
		Name:         "Admin User",
		Status:       "active",
		AuthProvider: "local",
		PasswordHash: string(passHash),
	}

	if err := db.Where("org_id=? AND email=?", org.ID, adminEmail).FirstOrCreate(&adminUser).Error; err != nil {
		return err
	}

	// -------------------------
	// 6) user_roles mapping (admin user -> admin role)
	// Use direct INSERT IGNORE to avoid GORM ordering/ID assumptions when the
	// join table uses a composite primary key (user_id, role_id, org_id).
	// -------------------------
	if res := db.Exec("INSERT IGNORE INTO user_roles (user_id, role_id, org_id) VALUES (?, ?, ?)", adminUser.ID, adminRole.ID, org.ID); res.Error != nil {
		return res.Error
	}

	log.Printf("✅ Seed OK | admin=%s pass=%s | org=%s | roles=[admin,devops,readonly] | perms=%d",
		adminEmail, adminPass, org.Slug, len(perms),
	)
	return nil
}
