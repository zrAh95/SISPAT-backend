package config

import (
	"os"
)

type Config struct {
	AppEnv          string
	AppPort         string
	DBHost          string
	DBPort          string
	DBName          string
	DBUser          string
	DBPass          string
	SessionCookie   string
	SessionTTLHours int
	SessionSecure   bool
	UploadDir       string
	MaxUploadBytes  int64
}

func Load() Config {
	return Config{
		AppEnv:          getenv("APP_ENV", "local"),
		AppPort:         getenv("APP_PORT", "8080"),
		DBHost:          getenv("DB_HOST", "127.0.0.1"),
		DBPort:          getenv("DB_PORT", "3306"),
		DBName:          getenv("DB_NAME", "sispat"),
		DBUser:          getenv("DB_USER", "root"),
		DBPass:          os.Getenv("DB_PASSWORD"),
		SessionCookie:   getenv("SESSION_COOKIE", "sispat_session"),
		SessionTTLHours: 24,
		SessionSecure:   os.Getenv("APP_ENV") == "production",
		UploadDir:       getenv("UPLOAD_DIR", "./storage"),
		MaxUploadBytes:  5 * 1024 * 1024,
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
