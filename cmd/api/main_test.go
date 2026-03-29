package main

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
    t.Setenv("SERVER_PORT", "8080")
    t.Setenv("DATABASE_URL", "postgres://localhost/test")
    t.Setenv("JWT_SECRET_KEY", "secret")

    cfg, err := loadConfig()
    if err != nil {
        t.Fatalf("failed to load config: %v", err)
    }

    if cfg.ServerPort != "8080" {
        t.Errorf("expected 8080, got %s", cfg.ServerPort)
    }
    if cfg.DatabaseURL != "postgres://localhost/test" {
        t.Errorf("expected postgres URL, got %s", cfg.DatabaseURL)
    }
}
