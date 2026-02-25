package handler

import (
	"github.com/galihaleanda/todo-app/internal/middleware"
	pkgjwt "github.com/galihaleanda/todo-app/pkg/jwt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Router wires all handlers to gin routes.
type Router struct {
	auth      *AuthHandler
	task      *TaskHandler
	project   *ProjectHandler
	analytics *AnalyticsHandler
	jwt       *pkgjwt.Manager
	log       *logrus.Logger
}

// NewRouter creates a Router with all dependencies.
func NewRouter(
	auth *AuthHandler,
	task *TaskHandler,
	project *ProjectHandler,
	analytics *AnalyticsHandler,
	jwt *pkgjwt.Manager,
	log *logrus.Logger,
) *Router {
	return &Router{auth: auth, task: task, project: project, analytics: analytics, jwt: jwt, log: log}
}

// Setup registers all routes and returns the gin engine.
func (r *Router) Setup() *gin.Engine {
	engine := gin.New()

	// Global middleware
	engine.Use(middleware.Recovery(r.log))
	engine.Use(middleware.RequestLogger(r.log))
	engine.Use(middleware.CORS())

	v1 := engine.Group("/api/v1")

	// Health check — no auth required
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes — public
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", r.auth.Register)
		authGroup.POST("/login", r.auth.Login)
		authGroup.POST("/refresh", r.auth.RefreshToken)
	}

	// Protected routes
	protected := v1.Group("")
	protected.Use(middleware.Auth(r.jwt))
	{
		protected.POST("/auth/logout", r.auth.Logout)

		// Tasks
		tasks := protected.Group("/tasks")
		{
			tasks.POST("", r.task.Create)
			tasks.GET("", r.task.List)
			tasks.GET("/:id", r.task.GetByID)
			tasks.PATCH("/:id", r.task.Update)
			tasks.DELETE("/:id", r.task.Delete)
		}

		// Projects
		projects := protected.Group("/projects")
		{
			projects.POST("", r.project.Create)
			projects.GET("", r.project.List)
			projects.GET("/:id", r.project.GetByID)
			projects.PATCH("/:id", r.project.Update)
			projects.DELETE("/:id", r.project.Delete)
		}

		// Analytics
		analytics := protected.Group("/analytics")
		{
			analytics.GET("/dashboard", r.analytics.Dashboard)
			analytics.GET("/daily", r.analytics.DailyStats)
		}
	}

	return engine
}
