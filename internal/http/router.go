package httpserver

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/you/teleport-lite/internal/auth"
    "github.com/you/teleport-lite/internal/rbac"
)

func NewRouter(db *gorm.DB, jwtSecret string) *gin.Engine {
    r := gin.Default()
    r.Static("/static", "./static")

    // Public
    r.POST("/api/v1/auth/login", loginHandler(db, jwtSecret))
    r.GET("/login", renderLogin())
    
    // Protected
    chk := rbac.Checker{DB: db}
    authMW := auth.JWT(jwtSecret)

    api := r.Group("/api/v1", authMW)
    {
        api.GET("/users", require(chk, "users:read"), listUsers(db))
        api.POST("/users", require(chk, "users:write"), createUser(db))
        api.GET("/roles", require(chk, "roles:read"), listRoles(db))
        api.POST("/roles", require(chk, "roles:write"), createRole(db))
        api.POST("/roles/:id/permissions", require(chk, "roles:write"), assignPerms(db))
        api.POST("/users/:id/roles", require(chk, "users:assign-role"), assignRole(db))

        api.GET("/resources", require(chk, "resources:read"), listResources(db))
        api.POST("/resources", require(chk, "resources:write"), createResource(db))

        api.GET("/audit", require(chk, "audit:read"), listAudit(db))
    }

    // Minimal HTML pages (Go templates) for admin
    r.GET("/", authMW, renderDashboard(db))
    r.GET("/users", authMW, renderUsers(db))
    r.GET("/roles", authMW, renderRoles(db))

    return r
}

func require(chk rbac.Checker, permKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, _ := c.Get("claims")
        cl := claims.(*auth.Claims)
        ok, err := chk.Can(c, cl.UserID, cl.OrgID, permKey)
        if err != nil || !ok {
            c.AbortWithStatusJSON(403, gin.H{"error":"forbidden", "missing": permKey})
            return
        }
        c.Next()
    }
}
