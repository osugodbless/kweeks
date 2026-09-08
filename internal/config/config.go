// Package config loads runtime configuration from environment variables
// following 12-factor conventions: config via env, never code.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config is the full runtime configuration for the kweeks server.
type Config struct {
	Env        string
	HTTPAddr   string
	ShutdownTO time.Duration
	// WebRoot, when set, serves the built SPA (web/dist) from the same origin
	// as /api so no separate reverse proxy is required. Empty disables it.
	WebRoot string

	// PublicURL is the externally-reachable base URL (origin only) used to
	// build the /claim link in redemption emails.
	PublicURL string

	// Database
	DatabaseURL string

	// BMONI Embedded (embedded-dev for sandbox). The owner key is the hex
	// secp256k1 private key that signs every provisioned wallet's
	// owner-proof + proposal digests. All host identity input (name, phone,
	// BVN, address) comes from the user via signup + the wallet-setup wizard —
	// never from a shared persona in env.
	BmoniBaseURL  string
	BmoniAPIKey   string
	BmoniOwnerKey string

	// Email (redemption recovery artifact; never the critical path)
	SmtpHost string
	SmtpPort int
	SmtpUser string
	SmtpPass string
	FromAddr string
}

// Load reads configuration from the environment.
func Load() (*Config, error) {
	loadDotEnv(".")
	c := &Config{
		Env: getEnv("KWEEKS_ENV", "development"), HTTPAddr: getEnv("KWEEKS_HTTP_ADDR", ":8080"),
		ShutdownTO: 10 * time.Second,
		WebRoot:    getEnv("KWEEKS_WEB_ROOT", ""),
		PublicURL:  getEnv("KWEEKS_PUBLIC_URL", ""),

		DatabaseURL: getEnv("DATABASE_URL", ""),

		BmoniBaseURL:  getEnv("BMONI_BASE_URL", "https://embedded-dev.bmoni.com"),
		BmoniAPIKey:   getEnv("BMONI_API_KEY", ""),
		BmoniOwnerKey: getEnv("BMONI_OWNER_KEY", ""),

		SmtpHost: getEnv("SMTP_HOST", ""),
		SmtpPort: getEnvInt("SMTP_PORT", 587),
		SmtpUser: getEnv("SMTP_USER", ""),
		SmtpPass: getEnv("SMTP_PASS", ""),
		FromAddr: getEnv("SMTP_FROM", "kweeks@example.com"),
	}

	if err := c.validate(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// loadDotEnv reads the project .env file into the process environment using
// github.com/joho/godotenv. godotenv never overrides variables that already
// exist in the real environment, so real env always wins over the file.
// A missing file is ignored silently.
func loadDotEnv(dir string) {
	_ = godotenv.Load(filepath.Join(dir, ".env"))
}
