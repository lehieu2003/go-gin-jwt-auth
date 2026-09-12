package main

import (
	"log"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/container"
	"github.com/example/go-gin-jwt-auth/internal/database"
	"github.com/example/go-gin-jwt-auth/internal/router"
	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. Connect to Database (PostgreSQL)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 4. Initialize DI Container (wiring Repos -> Services -> Handlers)
	c := container.NewContainer(cfg, db)

	// 5. Initialize Router
	r := router.SetupRouter(c)

	// 6. Start HTTP Server
	log.Printf("Starting server on port %s (env: %s)...", cfg.Port, cfg.Environment)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
