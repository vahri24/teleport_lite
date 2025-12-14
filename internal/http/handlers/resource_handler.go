package handlers

import (
	"encoding/json"
	"net/http"
	//"os"
	//"os/user"
	//"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"teleport_lite/internal/auth"
	"teleport_lite/internal/models"
)

func ListResources(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var resource []models.Resource
		if err := db.Find(&resource).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"resources": resource})
	}
}

// UpdateResourceUsers allows an admin to assign SSH users to a resource.
// Expects JSON: { "ssh_users": ["user1","user2"] }
func UpdateResourceUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload struct {
			SSHUsers []string `json:"ssh_users"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var resource models.Resource
		if err := db.First(&resource, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
			return
		}

		// Merge or set metadata.ssh_users
		var meta map[string]interface{}
		if len(resource.Metadata) > 0 {
			// datatypes.JSON is a []byte alias; use encoding/json to unmarshal
			_ = json.Unmarshal(resource.Metadata, &meta)
		}
		if meta == nil {
			meta = map[string]interface{}{}
		}
		meta["ssh_users"] = payload.SSHUsers
		metaJSON, _ := json.Marshal(meta)

		if err := db.Model(&resource).Update("metadata", datatypes.JSON(metaJSON)).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Audit log: who assigned users
		var initiatorName string
		var initiatorID int64
		var orgID int64
		if claimsI, ok := c.Get("claims"); ok {
			if cl, ok := claimsI.(*auth.Claims); ok {
				initiatorID = int64(cl.UserID)
				orgID = int64(cl.OrgID)
				var u models.User
				if err := db.First(&u, cl.UserID).Error; err == nil {
					initiatorName = u.Name
				}
			}
		}
		metaLog := map[string]interface{}{"ssh_users": payload.SSHUsers}
		metaLogJSON, _ := json.Marshal(metaLog)
		audit := models.AuditLog{
			OrgID:         orgID,
			UserID:        initiatorID,
			Action:        "resource.assign_ssh_users",
			ResourceType:  "resource",
			ResourceID:    resource.ID,
			Metadata:      datatypes.JSON(metaLogJSON),
			IP:            c.ClientIP(),
			UserAgent:     c.GetHeader("User-Agent"),
			InitiatorName: initiatorName,
			CreatedAt:     time.Now(),
		}
		_ = db.Create(&audit).Error

		c.JSON(http.StatusOK, gin.H{"message": "ssh users updated"})
	}
}
