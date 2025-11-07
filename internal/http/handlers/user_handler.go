package handlers

import (
    "net/http"
    "strings"
    "teleport_lite/internal/auth" 
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "teleport_lite/internal/models"
    "golang.org/x/crypto/bcrypt"
)

// ListUsers returns all users from DB
func ListUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var users []models.User
        if err := db.Find(&users).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"users": users})
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
			ID     int64              `json:"id"`
			OrgID  int64              `json:"org_id"`
			Email  string             `json:"email"`
			Name   string             `json:"name"`
			Status models.UserStatus  `json:"status"`
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
