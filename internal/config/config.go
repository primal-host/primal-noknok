package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const Version = "0.5.0"

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
	ListenAddr string

	OAuthPrivateKey string // multibase-encoded ES256 private key
	SessionTTL      string // duration string, e.g. "24h"
	OwnerDID       string
	OwnerUsername  string
	CookieDomain   string   // primary cookie domain (first entry)
	CookieDomains  []string // all cookie domains (parsed from COOKIE_DOMAINS)
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
		OwnerDID:      os.Getenv("OWNER_DID"),
		OwnerUsername: envOrDefault("OWNER_USERNAME", ""),
		CookieDomain: envOrDefault("COOKIE_DOMAIN", ".localhost"),
		PublicURL:     envOrDefault("PUBLIC_URL", "http://noknok.localhost"),
	}

	// Parse COOKIE_DOMAINS (comma-separated). Falls back to single CookieDomain.
	if domains := os.Getenv("COOKIE_DOMAINS"); domains != "" {
		for _, d := range strings.Split(domains, ",") {
			d = strings.TrimSpace(d)
			if d != "" {
				c.CookieDomains = append(c.CookieDomains, d)
			}
		}
		if len(c.CookieDomains) > 0 {
			c.CookieDomain = c.CookieDomains[0]
		}
	} else {
		c.CookieDomains = []string{c.CookieDomain}
	}

	pw, err := envOrFile("DB_PASSWORD")
	if err != nil {
		return nil, fmt.Errorf("DB_PASSWORD: %w", err)
	}
	c.DBPassword = pw

	oauthKey, err := envOrFile("OAUTH_KEY")
	if err != nil {
		return nil, fmt.Errorf("OAUTH_KEY: %w", err)
	}
	c.OAuthPrivateKey = oauthKey

	if c.OwnerDID == "" {
		return nil, fmt.Errorf("OWNER_DID is required")
	}

	if c.OAuthPrivateKey == "" {
		return nil, fmt.Errorf("OAUTH_KEY is required")
	}

	return c, nil
}

// DSN returns a PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		url.QueryEscape(c.DBUser), url.QueryEscape(c.DBPassword),
		c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

// DomainForHost returns the cookie domain that matches the given host.
// For example, host "ker.ai" matches domain ".ker.ai", host "gitea.primal.host"
// matches ".primal.host". Returns the primary domain if no match is found.
func (c *Config) DomainForHost(host string) string {
	// Strip port if present.
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	for _, d := range c.CookieDomains {
		base := strings.TrimPrefix(d, ".")
		if host == base || strings.HasSuffix(host, d) {
			return d
		}
	}
	return c.CookieDomain
}

// IsExternalHost returns true if the host belongs to a different cookie domain
// than the primary PublicURL.
func (c *Config) IsExternalHost(host string) bool {
	return c.DomainForHost(host) != c.CookieDomain
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
