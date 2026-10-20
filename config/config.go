package config

import (
	"os"
	"sync"
)

type AppConfig struct {
	DatabaseURL    string
	RedisURL       string
	JWTSecret      string
	EncryptionKey  string
	AdminAPIKey    string
	StoragePath    string
	AllowedOrigins []string
	Debug          bool
}

var (
	instance *AppConfig
	once     sync.Once
)

func Get() *AppConfig {
	once.Do(func() {
		instance = &AppConfig{
			DatabaseURL:    getEnv("DATABASE_URL", "postgres://admin:admin123@localhost:5432/appdb"),
			RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
			JWTSecret:      getEnv("JWT_SECRET", "supersecretkey123"),
			EncryptionKey:  getEnv("ENCRYPTION_KEY", "aes256encryptionkey1234567890123"),
			AdminAPIKey:    getEnv("ADMIN_API_KEY", ""),
			StoragePath:    getEnv("STORAGE_PATH", "/var/data/uploads"),
			AllowedOrigins: []string{"*"},
			Debug:          getEnv("DEBUG", "true") == "true",
		}
	})
	return instance
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetDatabaseCredentials() (string, string) {
	return getEnv("DB_USER", ""), getEnv("DB_PASSWORD", "")
}

func GetInternalAPIKey() string {
	return getEnv("INTERNAL_API_KEY", "")
}
