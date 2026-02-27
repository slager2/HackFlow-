package config

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all the configuration variables for the application
type Config struct {
	Env          string
	Port         string
	DBHost       string
	DBUser       string
	DBPass       string
	DBName       string
	DBPort       string
	TavilyAPIKey string
	GeminiAPIKey string
}

// Load reads the application configuration from environment variables
// and the .env file if it exists.
func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		slog.Debug("No .env file found, using system environment variables")
	}

	cfg := &Config{
		Env:  getEnvOrDefault("ENV", "development"),
		Port: getEnvOrDefault("PORT", "8080"),
		// For Docker Desktop, host.docker.internal connects to the host machine's localhost
		DBHost:       getEnvOrDefault("DB_HOST", "host.docker.internal"),
		DBUser:       getEnvOrDefault("DB_USER", "hackflow_user"),
		DBPass:       getEnvOrDefault("DB_PASSWORD", "supersecretpassword"),
		DBName:       getEnvOrDefault("DB_NAME", "hackflow"),
		DBPort:       getEnvOrDefault("DB_PORT", "5432"),
		TavilyAPIKey: os.Getenv("TAVILY_API_KEY"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}

	return cfg
}

func getEnvOrDefault(key, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		return fallback
	}
	return value
}
