package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/you/teleport-lite/internal/models"
)

func ListUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var users []models.User
        if err := db.Preload("Roles").Find(&users).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, users)
    }
}

func CreateUser(db *gorm.DB) gin.HandlerFunc {
    type req struct {
        OrgID   uint64 `json:"org_id" binding:"required"`
        Email   string `json:"email" binding:"required,email"`
        Name    string `json:"name"`
        Password string `json:"password"`
    }
    return func(c *gin.Context) {
        var r req
        if err := c.ShouldBindJSON(&r); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }
        // hash password if provided...
        u := models.User{OrgID: r.OrgID, Email: r.Email, Name: r.Name}
        if err := db.Create(&u).Error; err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }
        c.JSON(201, u)
    }
}
