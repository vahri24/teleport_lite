package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"teleport_lite/internal/auth"
	"teleport_lite/internal/http/handlers"
	//"teleport_lite/internal/models"
	"teleport_lite/internal/rbac"
)

func NewRouter(db *gorm.DB, jwtSecret string) *gin.Engine {
	r := gin.Default()
	r.LoadHTMLGlob("internal/ui/views/*.tmpl")
	r.Static("/static", "internal/ui/static")

	// favicon fix
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

    r.GET("/login", renderLogin())

	// ✅ Dashboard (PUBLIC PAGE for now so redirect works)
	r.GET("/dashboard", auth.JWT(jwtSecret), func(c *gin.Context) {
    claims, exists := c.Get("claims")
    if !exists {
        c.Redirect(http.StatusSeeOther, "/login")
        return
    }

    cl := claims.(*auth.Claims)
    c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
        "title":  "Dashboard",
        "name":   cl.Email,  // or cl.UserID, depending on your Claims struct
        "org_id": cl.OrgID,
    	})
	})

	// ✅ Resource (PUBLIC PAGE for now so redirect works)
	r.GET("/resources", func(c *gin.Context) {
	c.HTML(http.StatusOK, "resources.tmpl", gin.H{
		"title": "Resources",
	})
})


	// Public routes
	r.POST("/api/v1/auth/login", handlers.LoginHandler(db, jwtSecret))
	r.POST("/agents/register", handlers.RegisterAgent(db))
	r.POST("/agents/heartbeat", handlers.AgentHeartbeat(db))
	

	// ✅ Protected API routes (still secure)
	chk := rbac.Checker{DB: db}
	authMW := auth.JWT(jwtSecret)

	api := r.Group("/api/v1", authMW)
	{
		// Users
		api.GET("/users", require(chk, "users:read"), handlers.ListUsers(db))
		api.POST("/users", require(chk, "users:write"), handlers.CreateUser(db))
		api.POST("/users/:id/roles", require(chk, "users:assign-role"), assignRole(db))
		api.GET("/users/connect-list", require(chk, "users:read"), handlers.ListConnectUsers(db))


		// Roles
		api.GET("/roles", require(chk, "roles:read"), handlers.ListRoles(db))
		api.POST("/roles", require(chk, "roles:write"), handlers.CreateRole(db))
		api.POST("/roles/:id/permissions", require(chk, "roles:write"), assignPerms(db))

		// Resources
		api.GET("/resources", require(chk, "resources:read"), handlers.ListResources(db))
		//api.GET("/resources/local", require(chk, "resources:read"), handlers.GetLocalResource)
		api.POST("/resources", require(chk, "resources:write"), createResource(db))

		//SSH
		api.GET("/ws/ssh", handlers.SSHWS(db))

		// Audit Trail
		api.GET("/audit", require(chk, "audit:read"), handlers.ListAudit(db))

		// ✅ Agent Endpoints
		
	}

	// ✅ Remove protected root route to prevent template conflicts
	// r.GET("/", authMW, renderDashboard(db)) ❌

	return r
}

func renderLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{"title": "Login"})
	}
}

// Stub endpoints
//func listUsers(db *gorm.DB) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		var users []models.User
//		db.Find(&users)
//		c.JSON(http.StatusOK, gin.H{"users": users})
//	}
//}

//func createUser(db any) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		c.JSON(http.StatusCreated, gin.H{"message": "user created (stub)"})
//	}
//}

func assignRole(db any) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "role assigned (stub)"})
	}
}

func assignPerms(db any) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "permissions assigned (stub)"})
	}
}

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
