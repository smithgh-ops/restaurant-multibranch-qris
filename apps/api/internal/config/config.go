package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application loaded from environment variables.
type Config struct {
	// Server
	AppName string
	AppEnv  string
	Port    string

	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// CORS
	CORSOrigins string

	// JWT (placeholder; real secret must be set via env in production)
	JWTSecret string
}

// Load reads configuration from environment variables, falling back to safe defaults.
func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	return &Config{
		AppName: getEnv("APP_NAME", "RestoQRIS"),
		AppEnv:  getEnv("APP_ENV", "development"),
		Port:    getEnv("PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "resto"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "resto_db"),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		CORSOrigins: getEnv("CORS_ORIGINS", "http://localhost:5173"),

		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
