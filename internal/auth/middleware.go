package auth

import (
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID uint64 `json:"uid"`
    OrgID  uint64 `json:"oid"`
    jwt.RegisteredClaims
}

func JWT(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        h := c.GetHeader("Authorization")
        if !strings.HasPrefix(h, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"missing bearer token"})
            return
        }
        tokenStr := strings.TrimPrefix(h, "Bearer ")
        token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error":"invalid token"})
            return
        }
        c.Set("claims", token.Claims.(*Claims))
        c.Next()
    }
}
