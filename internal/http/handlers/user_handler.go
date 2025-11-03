package handlers

import (
    "net/http"
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
