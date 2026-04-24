package config

import "os"

// Config holds all application configuration.
type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
	Environment string
	MediaDir    string // Path to medical image files (V1 med-data-bak)
}

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", "radimaster-dev-secret-change-in-production"),
		Environment: getEnv("ENVIRONMENT", "development"),
		MediaDir:    getEnv("MEDIA_DIR", "./med-data"),
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
