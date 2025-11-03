package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"teleport_lite/internal/models"
)

// LoginHandler authenticates the user and returns JWT
func LoginHandler(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user models.User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"uid":   user.ID,
			"oid":   user.OrgID,
			"email": user.Email,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
			return
		}

		// ✅ Set JWT as cookie (browser will send it automatically)
		c.SetCookie(
			"token",         // name
			tokenString,     // value
			3600*24,         // expires in 1 day
			"/",             // path
			"",              // domain (same origin)
			false,           // secure (false for localhost; true for HTTPS)
			true,            // HttpOnly
		)

		// ✅ Also return token in JSON (for Postman or JS use)
		c.JSON(http.StatusOK, gin.H{
			"token": tokenString,
			"user": gin.H{
				"email":  user.Email,
				"name":   user.Name,
				"org_id": user.OrgID,
			},
		})
	}
}
