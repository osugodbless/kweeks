// Package config loads runtime configuration from environment variables
// following 12-factor conventions: config via env, never code.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config is the full runtime configuration for the kweeks server.
type Config struct {
	Env        string
	HTTPAddr   string
	ShutdownTO time.Duration

	// Database
	DatabaseURL string

	// BMONI Embedded (embedded-dev for sandbox)
	BmoniBaseURL          string
	BmoniAPIKey           string
	BmoniOwnerKey         string // hex secp256k1 private key for the instructor wallet
	BmoniInstructorUserID string
	BmoniWalletID         string // optional explicit recipient wallet for sandbox sends

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

		DatabaseURL: getEnv("DATABASE_URL", ""),

		BmoniBaseURL:          getEnv("BMONI_BASE_URL", "https://embedded-dev.bmoni.com"),
		BmoniAPIKey:           getEnv("BMONI_API_KEY", ""),
		BmoniOwnerKey:         getEnv("BMONI_OWNER_KEY", ""),
		BmoniInstructorUserID: getEnv("BMONI_INSTRUCTOR_USER_ID", ""),
		BmoniWalletID:         getEnv("BMONI_WALLET_ID", ""),

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

// loadDotEnv reads a simple KEY=VALUE .env file from dir and sets any key not
// already present in the environment (real environment wins over the file).
// Missing files are ignored silently.
func loadDotEnv(dir string) {
	path := filepath.Join(dir, ".env")
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists && key != "" {
			_ = os.Setenv(key, val)
		}
	}
}
