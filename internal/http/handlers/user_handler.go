package handlers

import (
    "net/http"
    "strings"
    "teleport_lite/internal/auth" 
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "teleport_lite/internal/models"
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
        var input struct {
            OrgID  int64  `json:"org_id" binding:"required"`
            Email  string `json:"email" binding:"required,email"`
            Name   string `json:"name"`
            Status string `json:"status"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        user := models.User{
            OrgID:  input.OrgID,
            Email:  input.Email,
            Name:   input.Name,
            Status: models.UserStatus(input.Status),
        }

        if err := db.Create(&user).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"user": user})
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
