package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	Environment          string
	DBHost               string
	DBPort               string
	DBUser               string
	DBPassword           string
	DBName               string
	DBSSLMode            string
	AccessTokenSecret    string
	RefreshTokenSecret   string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	CookieDomain         string
	CookieSecure         bool
	CookieSameSite       string
}

func LoadConfig() *Config {
	// Load .env file if it exists (ignore error if not present in production)
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	accessDuration, err := time.ParseDuration(getEnv("ACCESS_TOKEN_DURATION", "15m"))
	if err != nil {
		accessDuration = 15 * time.Minute
	}

	refreshDuration, err := time.ParseDuration(getEnv("REFRESH_TOKEN_DURATION", "168h")) // 7 days default
	if err != nil {
		refreshDuration = 7 * 24 * time.Hour
	}

	cookieSecure, _ := strconv.ParseBool(getEnv("COOKIE_SECURE", "false"))

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		Environment:          getEnv("ENV", "development"),
		DBHost:               getEnv("DB_HOST", "localhost"),
		DBPort:               getEnv("DB_PORT", "5432"),
		DBUser:               getEnv("DB_USER", "postgres"),
		DBPassword:           getEnv("DB_PASSWORD", "postgres"),
		DBName:               getEnv("DB_NAME", "auth_db"),
		DBSSLMode:            getEnv("DB_SSLMODE", "disable"),
		AccessTokenSecret:    getEnv("ACCESS_TOKEN_SECRET", "super-secret-access-token-key-change-in-production"),
		RefreshTokenSecret:   getEnv("REFRESH_TOKEN_SECRET", "super-secret-refresh-token-key-change-in-production"),
		AccessTokenDuration:  accessDuration,
		RefreshTokenDuration: refreshDuration,
		CookieDomain:         getEnv("COOKIE_DOMAIN", ""),
		CookieSecure:         cookieSecure,
		CookieSameSite:       getEnv("COOKIE_SAMESITE", "Lax"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
