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

	// Database
	DatabaseURL string

	// BMONI Embedded (embedded-dev for sandbox)
	BmoniBaseURL          string
	BmoniAPIKey           string
	BmoniOwnerKey         string // hex secp256k1 private key for the instructor wallet
	BmoniInstructorUserID string
	BmoniWalletID         string // optional explicit recipient wallet for sandbox sends

	// BMONI onboarding persona (NGN rail). The sandbox resolves a fixed set of
	// personas; provisioning uses these identity values verbatim. The three KYC
	// document uploads stay an operator step: point BmoniDocIdentification,
	// BmoniDocProofOfAddress, BmoniDocBiometric at JPEG/PNG files to have the
	// server submit them, or leave empty to stop before uploads.
	BmoniPersonaFirstName string
	BmoniPersonaLastName  string
	BmoniPersonaEmail     string
	BmoniPersonaPhone     string // E.164, e.g. +2348000000001
	BmoniPersonaBVN       string
	BmoniPersonaDOB       string // YYYY-MM-DD
	BmoniPersonaAddress   string
	BmoniPersonaCity      string
	BmoniPersonaState     string

	BmoniDocIdentification string
	BmoniDocProofOfAddress string
	BmoniDocBiometric      string
	BmoniProvisionOnSignup bool // when true, signup also provisions a real BMONI user + CNGN wallet

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

		BmoniPersonaFirstName: getEnv("BMONI_PERSONA_FIRST_NAME", ""),
		BmoniPersonaLastName:  getEnv("BMONI_PERSONA_LAST_NAME", ""),
		BmoniPersonaEmail:     getEnv("BMONI_PERSONA_EMAIL", ""),
		BmoniPersonaPhone:     getEnv("BMONI_PERSONA_PHONE", ""),
		BmoniPersonaBVN:       getEnv("BMONI_PERSONA_BVN", ""),
		BmoniPersonaDOB:       getEnv("BMONI_PERSONA_DOB", ""),
		BmoniPersonaAddress:   getEnv("BMONI_PERSONA_ADDRESS", ""),
		BmoniPersonaCity:      getEnv("BMONI_PERSONA_CITY", ""),
		BmoniPersonaState:     getEnv("BMONI_PERSONA_STATE", ""),

		BmoniDocIdentification: getEnv("BMONI_DOC_IDENTIFICATION", ""),
		BmoniDocProofOfAddress: getEnv("BMONI_DOC_PROOF_OF_ADDRESS", ""),
		BmoniDocBiometric:      getEnv("BMONI_DOC_BIOMETRIC", ""),
		BmoniProvisionOnSignup: getEnvBool("BMONI_PROVISION_ON_SIGNUP", false),

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

func getEnvBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
