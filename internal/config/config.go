package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config groups every runtime configuration needed by the API server.
type Config struct {
	Env               string
	Port              int
	DatabaseURL       string
	RedisURL          string
	SessionCookieName string
	SessionTTL        time.Duration
	CorsOrigins       []string
}

const (
	defaultPort         = 3000
	defaultSessionTTL   = 30 * 24 * time.Hour
	defaultCookieName   = "sessionID"
	defaultEnvironment  = "DEV"
	corsOriginsFallback = "[\"*\"]"
)

// Load builds a Config based on the environment variables present.
func Load() (Config, error) {
	cfg := Config{
		Env:               strings.ToUpper(getEnv("ENV", defaultEnvironment)),
		SessionCookieName: getEnv("SESSION_COOKIE_NAME", defaultCookieName),
		SessionTTL:        defaultSessionTTL,
	}

	if ttlStr := os.Getenv("SESSION_TTL_HOURS"); ttlStr != "" {
		if ttl, err := strconv.Atoi(ttlStr); err == nil && ttl > 0 {
			cfg.SessionTTL = time.Duration(ttl) * time.Hour
		}
	}

	cfg.Port = parsePort(getEnv("PORT", strconv.Itoa(defaultPort)))
	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	cfg.RedisURL = os.Getenv("REDIS_URL")

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	if cfg.RedisURL == "" {
		return Config{}, fmt.Errorf("REDIS_URL is required")
	}

	cfg.CorsOrigins = parseOrigins(getEnv("CORS_ORIGINS", corsOriginsFallback))

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func parsePort(value string) int {
	port, err := strconv.Atoi(value)
	if err != nil || port == 0 {
		return defaultPort
	}

	return port
}

func parseOrigins(raw string) []string {
	var origins []string
	if err := json.Unmarshal([]byte(raw), &origins); err == nil && len(origins) > 0 {
		return origins
	}

	// Allow comma separated fallback to make local development easier.
	cleaned := strings.TrimSpace(raw)
	if cleaned == "" {
		return []string{"*"}
	}

	if strings.Contains(cleaned, ",") {
		parts := strings.Split(cleaned, ",")
		for i := range parts {
			parts[i] = strings.TrimSpace(parts[i])
		}
		return parts
	}

	return []string{cleaned}
}
