package container

import (
	"gorm.io/gorm"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/handler"
	"github.com/example/go-gin-jwt-auth/internal/repository"
	"github.com/example/go-gin-jwt-auth/internal/service"
)

// Container holds all application dependencies (handlers, services, repositories)
type Container struct {
	Config      *config.Config
	DB          *gorm.DB
	AuthHandler *handler.AuthHandler
	TodoHandler *handler.TodoHandler
}

// NewContainer wires together all layers (Repository -> Service -> Handler)
func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	// 1. Repositories
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	todoRepo := repository.NewTodoRepository(db)

	// 2. Services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)
	todoService := service.NewTodoService(todoRepo)

	// 3. Handlers
	authHandler := handler.NewAuthHandler(authService, cfg)
	todoHandler := handler.NewTodoHandler(todoService)

	return &Container{
		Config:      cfg,
		DB:          db,
		AuthHandler: authHandler,
		TodoHandler: todoHandler,
	}
}
