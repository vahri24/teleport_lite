package httpserver

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "github.com/golang-jwt/jwt/v5"
    "teleport_lite/internal/auth"
    "teleport_lite/internal/rbac"
    "teleport_lite/internal/http/handlers"
    "teleport_lite/internal/models"
)

func NewRouter(db *gorm.DB, jwtSecret string) *gin.Engine {
    r := gin.Default()
    r.LoadHTMLGlob("internal/ui/views/*.tmpl") 
    r.Static("/static", "internal/ui/static")

    // Public routes
    r.POST("/api/v1/auth/login", handlers.LoginHandler(db, jwtSecret))
    r.GET("/login", renderLogin())

    r.GET("/dashboard", func(c *gin.Context) {
    name := c.Query("name")
    org := c.Query("org_id")

    c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
        "title":  "Dashboard",
        "name":   name,
        "org_id": org,
        })
    })

    // Protected routes
    chk := rbac.Checker{DB: db}
    authMW := auth.JWT(jwtSecret)

    api := r.Group("/api/v1", authMW)
    {
        
        // Users
        api.GET("/users", require(chk, "users:read"), handlers.ListUsers(db))
        api.POST("/users", require(chk, "users:write"), handlers.CreateUser(db))
        api.POST("/users/:id/roles", require(chk, "users:assign-role"), assignRole(db))

        // Roles
        api.GET("/roles", require(chk, "roles:read"), handlers.ListRoles(db))
        api.POST("/roles", require(chk, "roles:write"), handlers.CreateRole(db))
        api.POST("/roles/:id/permissions", require(chk, "roles:write"), assignPerms(db))

        // Resources
        api.GET("/resources", require(chk, "resources:read"), listResources(db))
        api.POST("/resources", require(chk, "resources:write"), createResource(db))

        // Audit
        api.GET("/audit", require(chk, "audit:read"), listAudit(db))
    }

    // UI pages (Go templates)
    r.GET("/", authMW, renderDashboard(db))
    r.GET("/users", authMW, renderUsers(db))
    r.GET("/roles", authMW, renderRoles(db))

    return r
}

// --- Temporary stub handlers so the app compiles --- //

// Auth
func loginHandler(db any, jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Normally you'd verify username/password here
        // For testing, we just issue a token for a fake user
        token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
            "uid": 1,   // test user ID
            "oid": 1,   // test org ID
            "exp": time.Now().Add(time.Hour * 24).Unix(),
        })

        tokenString, err := token.SignedString([]byte(jwtSecret))
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{
            "token": tokenString,
            "note":  "use this token in Authorization header: Bearer <token>",
        })
    }
}

func renderLogin() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.HTML(http.StatusOK, "login.tmpl", gin.H{"title": "Login"})
    }
}

// Users
func listUsers(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var users []models.User
        db.Find(&users)
        c.JSON(http.StatusOK, gin.H{"users": users})
    }
}

func createUser(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusCreated, gin.H{"message": "user created (stub)"})
    }
}

func assignRole(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "role assigned (stub)"})
    }
}

// Roles
func ListRoles(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"roles": []string{}})
    }
}

func CreateRole(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusCreated, gin.H{"message": "role created (stub)"})
    }
}

func assignPerms(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "permissions assigned (stub)"})
    }
}

// Resources
func listResources(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"resources": []string{}})
    }
}

func createResource(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusCreated, gin.H{"message": "resource created (stub)"})
    }
}

// Audit
func listAudit(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"audit": []string{}})
    }
}

// UI Templates
func renderDashboard(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{"title": "Dashboard"})
    }
}

func renderUsers(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.HTML(http.StatusOK, "users.tmpl", gin.H{"title": "Users"})
    }
}

func renderRoles(db any) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.HTML(http.StatusOK, "roles.tmpl", gin.H{"title": "Roles"})
    }
}

// Middleware for permission checks
func require(chk rbac.Checker, permKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims, _ := c.Get("claims")
        cl := claims.(*auth.Claims)
        ok, err := chk.Can(c, cl.UserID, cl.OrgID, permKey)
        if err != nil || !ok {
            c.AbortWithStatusJSON(403, gin.H{"error": "forbidden", "missing": permKey})
            return
        }
        c.Next()
    }
}
