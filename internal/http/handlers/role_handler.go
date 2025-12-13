package handlers

import (
	"net/http"
	"teleport_lite/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ✅ Exported (capitalized) functions
func ListRoles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var roles []models.Role
		if err := db.Find(&roles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Build response with user counts per role
		out := make([]map[string]interface{}, 0, len(roles))
		for _, r := range roles {
			var cnt int64
			if err := db.Model(&models.UserRole{}).Where("role_id = ?", r.ID).Count(&cnt).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			item := map[string]interface{}{
				"id":          r.ID,
				"org_id":      r.OrgID,
				"name":        r.Name,
				"slug":        r.Slug,
				"description": r.Description,
				"is_system":   r.IsSystem,
				"created_at":  r.CreatedAt,
				"users_count": cnt,
			}
			out = append(out, item)
		}

		c.JSON(http.StatusOK, gin.H{"roles": out})
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
