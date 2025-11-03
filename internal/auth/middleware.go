package auth

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims structure.
type Claims struct {
	UserID uint64 `json:"uid"`
	OrgID  uint64 `json:"oid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWT returns a Gin middleware that validates JWT tokens from
// either the Authorization header or a "token" cookie.
func JWT(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")

		// ✅ Fallback: read from cookie if no Authorization header
		if tokenStr == "" {
			if cookie, err := c.Cookie("token"); err == nil {
				tokenStr = "Bearer " + cookie
			}
		}

		// If still missing, reject
		if tokenStr == "" {
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		// Extract and set claims in context
		if claims, ok := token.Claims.(*Claims); ok {
			c.Set("claims", claims)
			c.Next()
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
		c.Abort()
	}
}
