package handlers

import (
	"net/http"
	"strings"
	"teleport_lite/internal/auth"
	"teleport_lite/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ListUsers returns all users from DB
func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		// preload Roles so frontend can display assigned roles
		if err := db.Preload("Roles").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}

// DeactivateUser sets a user's status to suspended. Requires appropriate permission at the route level.
func DeactivateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		user.Status = models.UserSuspended
		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deactivated"})
	}
}

// ActivateUser sets a user's status to active.
func ActivateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		user.Status = models.UserActive
		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user activated"})
	}
}

// ChangePassword allows an admin to set a new password for a user.
func ChangePassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload struct {
			Password string `json:"password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		user.PasswordHash = string(hash)
		if err := db.Save(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "password updated"})
	}
}

// AssignRoles replaces the roles assigned to a user with the provided list.
func AssignRoles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var payload struct {
			RoleIDs []int64 `json:"role_ids"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.First(&user, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// Load role models
		var roles []models.Role
		if len(payload.RoleIDs) > 0 {
			if err := db.Where("id IN ?", payload.RoleIDs).Find(&roles).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		// Replace associations manually because the join table `user_roles`
		// includes an `org_id` column which GORM's default many2many
		// helpers won't populate. Use a transaction: delete existing
		// mappings for this user+org and insert the new ones.
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": tx.Error.Error()})
			return
		}

		if err := tx.Exec("DELETE FROM user_roles WHERE user_id = ? AND org_id = ?", user.ID, user.OrgID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, r := range roles {
			if err := tx.Exec("INSERT INTO user_roles (user_id, role_id, org_id) VALUES (?, ?, ?)", user.ID, r.ID, user.OrgID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "roles updated"})
	}
}

// CreateUser inserts a new user
func CreateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type inputDTO struct {
			OrgID    int64  `json:"org_id,string" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Name     string `json:"name" binding:"required"`
			Password string `json:"password" binding:"required"`
			Status   string `json:"status"` // optional; e.g. "active"
		}
		var in inputDTO
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Basic normalization
		in.Email = strings.TrimSpace(strings.ToLower(in.Email))
		in.Name = strings.TrimSpace(in.Name)

		// App-level password rules (adjust to taste)
		if len(in.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 characters"})
			return
		}

		// Prevent duplicate email per org (unique key recommended at DB level too)
		var existing int64
		if err := db.Model(&models.User{}).
			Where("org_id = ? AND email = ?", in.OrgID, in.Email).
			Count(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if existing > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists in this organization"})
			return
		}

		// Hash password
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		// Build user model (assumes models.User has PasswordHash string field)
		user := models.User{
			OrgID:        in.OrgID,
			Email:        in.Email,
			Name:         in.Name,
			Status:       models.UserStatus(in.Status),
			PasswordHash: string(hash),
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Safe response (don’t return password/hash)
		type userResp struct {
			ID     int64             `json:"id"`
			OrgID  int64             `json:"org_id"`
			Email  string            `json:"email"`
			Name   string            `json:"name"`
			Status models.UserStatus `json:"status"`
		}
		c.JSON(http.StatusCreated, gin.H{"user": userResp{
			ID: user.ID, OrgID: user.OrgID, Email: user.Email, Name: user.Name, Status: user.Status,
		}})
	}
}

// ListConnectUsers returns current user's allowed SSH identities
func ListConnectUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		cl := claims.(*auth.Claims)

		var user models.User
		if err := db.First(&user, cl.UserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// Split the connect_user field into list
		var userList []string
		if user.ConnectUser != "" {
			userList = strings.Split(user.ConnectUser, ",")
			for i := range userList {
				userList[i] = strings.TrimSpace(userList[i])
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"email":        user.Email,
			"connect_user": userList,
		})
	}
}

// MeHandler returns current user info and permission keys
func MeHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsI, ok := c.Get("claims")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		cl := claimsI.(*auth.Claims)

		var user models.User
		if err := db.First(&user, cl.UserID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		var perms []string
		// JOIN user_roles -> roles -> role_permissions -> permissions
		if err := db.Table("user_roles ur").
			Joins("JOIN roles r ON r.id = ur.role_id AND r.org_id = ?", cl.OrgID).
			Joins("JOIN role_permissions rp ON rp.role_id = r.id").
			Joins("JOIN permissions p ON p.id = rp.permission_id").
			Where("ur.user_id = ? AND ur.org_id = ?", cl.UserID, cl.OrgID).
			Pluck("p.`key`", &perms).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"user":        user,
			"permissions": perms,
		})
	}
}
