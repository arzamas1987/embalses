package config

import (
	"os"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	DatabaseURL string
	APIAddr     string
	MCPAddr     string
	WebPort     string
	GeminiKey   string
	GeminiModel string
	AEMETKey    string
	AppEnv      string
}

// Load reads configuration from environment variables.
func Load() Config {
	return Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"),
		APIAddr:     getEnv("API_ADDR", ":8080"),
		MCPAddr:     getEnv("MCP_ADDR", ":8081"),
		WebPort:     getEnv("WEB_PORT", "5173"),
		GeminiKey:   getEnv("GEMINI_API_KEY", ""),
		GeminiModel: getEnv("GEMINI_MODEL", ""),
		AEMETKey:    getEnv("AEMET_API_KEY", ""),
		AppEnv:      getEnv("APP_ENV", "development"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
