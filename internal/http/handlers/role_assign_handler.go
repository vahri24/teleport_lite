package handlers

import (
	"net/http"
	"teleport_lite/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ================================
// 1. List users with assigned roles
// ================================
func ListUserRoles(gdb *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User

		// Load roles through user_roles table
		if err := gdb.Preload("Roles").Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"users": users})
	}
}
