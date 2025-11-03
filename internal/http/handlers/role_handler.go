package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "teleport_lite/internal/models"
)

// ✅ Exported (capitalized) functions
func ListRoles(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var roles []models.Role
        if err := db.Find(&roles).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, gin.H{"roles": roles})
    }
}

func CreateRole(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var input struct {
            OrgID int64  `json:"org_id" binding:"required"`
            Name  string `json:"name" binding:"required"`
            Slug  string `json:"slug" binding:"required"`
        }

        if err := c.ShouldBindJSON(&input); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        role := models.Role{
            OrgID: input.OrgID,
            Name:  input.Name,
            Slug:  input.Slug,
        }

        if err := db.Create(&role).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusCreated, gin.H{"role": role})
    }
}
