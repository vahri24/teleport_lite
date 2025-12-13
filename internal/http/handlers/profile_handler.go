package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"teleport_lite/internal/auth"
	"teleport_lite/internal/models"
)

// ProfileHandler renders the profile page for the currently authenticated user.
func ProfileHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsI, ok := c.Get("claims")
		if !ok {
			// JWT middleware should have redirected, but ensure fallback.
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		cl := claimsI.(*auth.Claims)

		var user models.User
		if err := db.Where("id = ? AND org_id = ?", cl.UserID, cl.OrgID).First(&user).Error; err != nil {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		c.HTML(http.StatusOK, "profile.tmpl", gin.H{
			"title": "Profile",
			"User":  user,
		})
	}
}
