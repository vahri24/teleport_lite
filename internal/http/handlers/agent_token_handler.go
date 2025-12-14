package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"teleport_lite/internal/auth"
	"teleport_lite/internal/models"
)

// CreateRegistrationToken creates a new one-time or time-limited registration token.
// Protected endpoint — should be called by admins.
func CreateRegistrationToken(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ResourceID *int64 `json:"resource_id"`
			TTLMinutes int    `json:"ttl_minutes"` // optional, 0 = no expiry
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// generate random token
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}
		tok := base64.RawURLEncoding.EncodeToString(b)

		var expires *time.Time
		if req.TTLMinutes > 0 {
			t := time.Now().Add(time.Duration(req.TTLMinutes) * time.Minute)
			expires = &t
		}

		rt := models.RegistrationToken{
			Token:      tok,
			ResourceID: req.ResourceID,
			ExpiresAt:  expires,
			Used:       false,
		}

		if err := db.Create(&rt).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Record audit log for token creation
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

		meta := map[string]interface{}{"ttl_minutes": req.TTLMinutes}
		metaJSON, _ := json.Marshal(meta)

		audit := models.AuditLog{
			OrgID:         orgID,
			UserID:        initiatorID,
			Action:        "registration_token.create",
			ResourceType:  "registration_token",
			ResourceID:    rt.ID,
			Metadata:      datatypes.JSON(metaJSON),
			IP:            c.ClientIP(),
			UserAgent:     c.GetHeader("User-Agent"),
			InitiatorName: initiatorName,
			CreatedAt:     time.Now(),
		}
		_ = db.Create(&audit).Error

		c.JSON(http.StatusCreated, gin.H{
			"token":      tok,
			"expires_at": expires,
			"id":         rt.ID,
		})
	}
}
