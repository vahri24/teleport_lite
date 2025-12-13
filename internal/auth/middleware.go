package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"teleport_lite/internal/models"
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID uint64 `json:"uid"`
	OrgID  uint64 `json:"oid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWT returns a Gin middleware that validates JWT tokens from
// either the Authorization header or a "token" cookie and verifies
// that the user is still active in the database.
func JWT(db *gorm.DB, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")

		// ✅ Fallback: read from cookie if no Authorization header
		if tokenStr == "" {
			if cookie, err := c.Cookie("token"); err == nil {
				tokenStr = "Bearer " + cookie
			}
		}

		// If still missing, decide response format based on the request.
		if tokenStr == "" {
			accept := c.GetHeader("Accept")
			if strings.Contains(accept, "text/html") && c.Request.Method == "GET" {
				// For browser navigation requests, redirect to login page.
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			c.Abort()
			return
		}

		// Trim "Bearer " prefix
		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		// Parse the JWT
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			accept := c.GetHeader("Accept")
			if strings.Contains(accept, "text/html") && c.Request.Method == "GET" {
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			accept := c.GetHeader("Accept")
			if strings.Contains(accept, "text/html") && c.Request.Method == "GET" {
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			c.Abort()
			return
		}

		// Verify user still exists and is active
		var user models.User
		if err := db.First(&user, claims.UserID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}
		if user.Status != models.UserActive {
			accept := c.GetHeader("Accept")
			if strings.Contains(accept, "text/html") && c.Request.Method == "GET" {
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
			c.JSON(http.StatusForbidden, gin.H{"error": "account suspended"})
			c.Abort()
			return
		}

		// Set claims in context and proceed
		c.Set("claims", claims)
		c.Next()
	}
}
