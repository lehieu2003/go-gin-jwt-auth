package database

import (
	"fmt"
	"log"

	"github.com/example/go-gin-jwt-auth/internal/config"
	"github.com/example/go-gin-jwt-auth/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	logLevel := logger.Info
	if cfg.Environment == "production" {
		logLevel = logger.Error
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	// Run auto migrations
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	log.Println("Running database migrations...")
	err := db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.Todo{},
	)
	if err != nil {
		return err
	}
	log.Println("Database migration completed successfully")
	return nil
}
