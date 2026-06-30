package config

import (
	"os"
	"strings"
	"testing"
)

const validSecret = "test-secret-at-least-32-bytes-long-1234567890"

func TestLoad(t *testing.T) {
	t.Run("valid secret applies defaults", func(t *testing.T) {
		t.Setenv("JWT_SECRET", validSecret)
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != 8080 {
			t.Errorf("Port = %d, want default 8080", cfg.Port)
		}
		if cfg.JWTExpiry.Hours() != 1 {
			t.Errorf("JWTExpiry = %v, want default 1h", cfg.JWTExpiry)
		}
		if len(cfg.CORSAllowedOrigins) != 1 || cfg.CORSAllowedOrigins[0] != "http://localhost:3000" {
			t.Errorf("CORSAllowedOrigins = %v, want [http://localhost:3000]", cfg.CORSAllowedOrigins)
		}
	})

	t.Run("short secret is rejected", func(t *testing.T) {
		t.Setenv("JWT_SECRET", "tooshort")
		if _, err := Load(); err == nil {
			t.Fatal("expected error for a sub-32-byte secret, got nil")
		}
	})

	t.Run("missing secret is rejected", func(t *testing.T) {
		os.Unsetenv("JWT_SECRET")
		if _, err := Load(); err == nil {
			t.Fatal("expected error for a missing secret, got nil")
		}
	})
}

func TestLogValueRedactsSecret(t *testing.T) {
	secret := "super-secret-value-please-keep-this-hidden"
	c := Config{JWTSecret: secret}
	rendered := c.LogValue().String()

	if strings.Contains(rendered, secret) {
		t.Errorf("LogValue leaked the JWT secret: %q", rendered)
	}
	if !strings.Contains(rendered, "[REDACTED]") {
		t.Errorf("LogValue missing redaction marker: %q", rendered)
	}
}
