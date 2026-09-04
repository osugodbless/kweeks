package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDotEnvLoadsFileAndRealEnvWins(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("KWEEKS_HTTP_ADDR=:9999\nSMTP_FROM=kweeks@test.ng\nKWEEKS_ENV=fromfile\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.Unsetenv("KWEEKS_HTTP_ADDR")
		os.Unsetenv("SMTP_FROM")
		os.Unsetenv("KWEEKS_ENV")
	})

	os.Unsetenv("KWEEKS_HTTP_ADDR")
	os.Unsetenv("SMTP_FROM")
	os.Unsetenv("KWEEKS_ENV")
	loadDotEnv(dir)

	if got := os.Getenv("KWEEKS_HTTP_ADDR"); got != ":9999" {
		t.Fatalf("file value not loaded: %q", got)
	}
	if got := os.Getenv("SMTP_FROM"); got != "kweeks@test.ng" {
		t.Fatalf("file value not loaded: %q", got)
	}

	// Real environment must win over the file (godotenv never overrides).
	os.Setenv("KWEEKS_ENV", "production")
	loadDotEnv(dir)
	if got := os.Getenv("KWEEKS_ENV"); got != "production" {
		t.Fatalf("real env should win, got %q", got)
	}

	// Missing file is a silent no-op.
	loadDotEnv(filepath.Join(t.TempDir(), "nope"))
}
