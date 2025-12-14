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

	// ✅ Secure static route for agent download
	shared := r.Group("/internal/shared")
	shared.Use(auth.JWT(db, jwtSecret))
	{
		shared.Static("/", "./internal/shared")
	}
	// favicon fix
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	r.GET("/login", renderLogin())
	// Logout route clears cookie server-side and redirects to login
	r.GET("/logout", handlers.LogoutHandler())
	// Profile (protected)
	r.GET("/profile", auth.JWT(db, jwtSecret), handlers.ProfileHandler(db))

	// ✅ Dashboard (PUBLIC PAGE for now so redirect works)
	r.GET("/dashboard", auth.JWT(db, jwtSecret), func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.Redirect(http.StatusSeeOther, "/login")
			return
		}

		cl := claims.(*auth.Claims)
		c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
			"title":  "Dashboard",
			"name":   cl.Email, // or cl.UserID, depending on your Claims struct
			"org_id": cl.OrgID,
		})
	})

	// ✅ Users
	r.GET("/users", func(c *gin.Context) {
		c.HTML(http.StatusOK, "users.tmpl", gin.H{
			"title": "Users",
		})
	})

	// ✅ Resource
	r.GET("/resources", func(c *gin.Context) {
		c.HTML(http.StatusOK, "resources.tmpl", gin.H{
			"title": "Resources",
		})
	})

	r.GET("/audit", func(c *gin.Context) {
		c.HTML(http.StatusOK, "audit.tmpl", gin.H{
			"title": "Audit",
		})
	})

	r.GET("/roles", func(c *gin.Context) {
		c.HTML(http.StatusOK, "roles.tmpl", gin.H{
			"title": "Roles",
		})
	})

	// Public routes
	r.POST("/api/v1/auth/login", handlers.LoginHandler(db, jwtSecret))
	r.POST("/agents/register", handlers.RegisterAgent(db))
	r.POST("/agents/heartbeat", handlers.AgentHeartbeat(db))

	// ✅ Protected API routes (still secure)
	chk := rbac.Checker{DB: db}
	authMW := auth.JWT(db, jwtSecret)

	api := r.Group("/api/v1", authMW)
	{
		// Current user info & permissions
		api.GET("/me", handlers.MeHandler(db))
		// Users
		api.GET("/users", require(chk, "users:read"), handlers.ListUsers(db))
		api.POST("/users", require(chk, "users:write"), handlers.CreateUser(db))
		api.POST("/users/:id/deactivate", require(chk, "users:assign-role"), handlers.DeactivateUser(db))
		api.POST("/users/:id/activate", require(chk, "users:assign-role"), handlers.ActivateUser(db))
		api.POST("/users/:id/password", require(chk, "users:assign-role"), handlers.ChangePassword(db))
		//api.POST("/users/:id/roles", require(chk, "users:assign-role"), assignRole(db))
		api.GET("/users/connect-list", require(chk, "users:read"), handlers.ListConnectUsers(db))

		// Roles
		api.GET("/roles", require(chk, "roles:read"), handlers.ListRoles(db))
		api.POST("/roles", require(chk, "roles:write"), handlers.CreateRole(db))
		api.POST("/roles/:id/permissions", require(chk, "roles:write"), assignPerms(db))

		// Assign_Roles
		assign := api.Group("/assign")
		assign.GET("/users", require(chk, "roles:read"), handlers.ListUserRoles(db))

		// Assign roles to a user
		api.POST("/users/:id/roles", require(chk, "users:assign-role"), handlers.AssignRoles(db))

		// Resources
		api.GET("/resources", require(chk, "resources:read"), handlers.ListResources(db))
		// Registration tokens for agents/resources (requires generate-token permission)
		api.POST("/agents/tokens", require(chk, "resources:generate-token"), handlers.CreateRegistrationToken(db))
		// Assign SSH users to a resource (admin only)
		api.POST("/resources/:id/users", require(chk, "users:assign-role"), handlers.UpdateResourceUsers(db))
		// Assign resource access to a user (admin only)
		api.POST("/users/:id/access", require(chk, "users:assign-role"), handlers.UpdateUserAccess(db))
		//api.GET("/resources/local", require(chk, "resources:read"), handlers.GetLocalResource)
		api.POST("/resources", require(chk, "resources:write"), createResource(db))

		//SSH
		api.GET("/ws/ssh", handlers.SSHWS(db))

		// Audit Trail
		api.GET("/audit", require(chk, "audit:read"), handlers.ListAudit(db))

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
