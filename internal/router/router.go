package router

import (
	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/container"
	"github.com/example/go-gin-jwt-auth/internal/handler"
	"github.com/example/go-gin-jwt-auth/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures all API routes using the DI container
func SetupRouter(c *container.Container) *gin.Engine {
	r := gin.Default()

	// Global middlewares
	r.Use(middleware.CORSMiddleware())

	// Health check route
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status":  "ok",
			"service": "go-gin-auth-service",
		})
	})

	// API v1 routes
	api := r.Group("/api/v1")
	{
		setupAuthRoutes(api, c.AuthHandler, c.Config)
		setupTodoRoutes(api, c.TodoHandler, c.Config)
	}

	return r
}

func setupAuthRoutes(rg *gin.RouterGroup, h *handler.AuthHandler, cfg *config.Config) {
	auth := rg.Group("/auth")
	{
		// Public endpoints
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)

		// Protected endpoints
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(cfg))
		{
			protected.GET("/me", h.GetMe)
		}
	}
}

func setupTodoRoutes(rg *gin.RouterGroup, h *handler.TodoHandler, cfg *config.Config) {
	todos := rg.Group("/todos")
	todos.Use(middleware.AuthMiddleware(cfg))
	{
		todos.POST("", h.Create)
		todos.GET("", h.List)
		todos.GET("/:id", h.Get)
		todos.PATCH("/:id", h.Update)
		todos.DELETE("/:id", h.Delete)
	}
}