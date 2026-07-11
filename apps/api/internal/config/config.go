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

	// JWT
	JWTSecret           string
	AccessTokenMinutes  int // duration of access token in minutes
	RefreshTokenDays    int // duration of refresh token in days

	// File upload
	UploadDir     string // local directory to store uploaded files
	PublicBaseURL string // base URL used to build public URLs for uploaded files

	// Payment gateway
	// PaymentGateway selects the active QRIS provider: "mock" (default) or "midtrans".
	PaymentGateway string
	// MidtransServerKey is the Midtrans Server Key (only required when PaymentGateway = "midtrans").
	MidtransServerKey string
	// MidtransIsProduction switches between the Midtrans sandbox (false) and production (true) endpoints.
	MidtransIsProduction bool
}

// Load reads configuration from environment variables, falling back to safe defaults.
func Load() *Config {
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	accessTokenMinutes, _ := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_MINUTES", "15"))
	refreshTokenDays, _ := strconv.Atoi(getEnv("JWT_REFRESH_TOKEN_DAYS", "30"))
	if accessTokenMinutes <= 0 {
		accessTokenMinutes = 15
	}
	if refreshTokenDays <= 0 {
		refreshTokenDays = 30
	}
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

		JWTSecret:          getEnv("JWT_SECRET", "change-me-in-production"),
		AccessTokenMinutes: accessTokenMinutes,
		RefreshTokenDays:   refreshTokenDays,

		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:8080"),

		PaymentGateway:       getEnv("PAYMENT_GATEWAY", "mock"),
		MidtransServerKey:    getEnv("MIDTRANS_SERVER_KEY", ""),
		MidtransIsProduction: getEnv("MIDTRANS_IS_PRODUCTION", "false") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
