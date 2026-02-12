package config

import (
	"fmt"
	"os"
	"strings"
)

const Version = "0.1.0"

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	ListenAddr string

	SessionSecret string
	SessionTTL    string // duration string, e.g. "24h"
	OwnerDID      string
	CookieDomain  string
	PublicURL      string
}

// Load reads configuration from environment variables.
// Supports _FILE suffix for Docker secrets (e.g. DB_PASSWORD_FILE).
func Load() (*Config, error) {
	c := &Config{
		DBHost:       envOrDefault("DB_HOST", "localhost"),
		DBPort:       envOrDefault("DB_PORT", "5432"),
		DBName:       envOrDefault("DB_NAME", "noknok"),
		DBUser:       envOrDefault("DB_USER", "dba_noknok"),
		DBSSLMode:    envOrDefault("DB_SSLMODE", "disable"),
		ListenAddr:   envOrDefault("LISTEN_ADDR", ":4321"),
		SessionTTL:   envOrDefault("SESSION_TTL", "24h"),
		OwnerDID:     os.Getenv("OWNER_DID"),
		CookieDomain: envOrDefault("COOKIE_DOMAIN", ".localhost"),
		PublicURL:     envOrDefault("PUBLIC_URL", "http://noknok.localhost"),
	}

	pw, err := envOrFile("DB_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("DB_PASSWORD: %w", err)
	}
	c.DBPassword = pw

	secret, err := envOrFile("SESSION_SECRET")
	if err != nil {
		return nil, fmt.Errorf("SESSION_SECRET: %w", err)
	}
	c.SessionSecret = secret

	if c.OwnerDID == "" {
		return nil, fmt.Errorf("OWNER_DID is required")
	}

	if c.SessionSecret == "" {
		return nil, fmt.Errorf("SESSION_SECRET is required")
	}

	return c, nil
}

// DSN returns a PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envOrFile reads a value from env var KEY, or from a file at KEY_FILE.
func envOrFile(key string) (string, error) {
	if v := os.Getenv(key); v != "" {
		return v, nil
	}
	fileKey := key + "_FILE"
	if path := os.Getenv(fileKey); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", fileKey, err)
		}
		return strings.TrimSpace(string(data)), nil
	}
	return "", nil
}
