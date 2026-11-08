package config

import (
	"testing"
)

func TestLoadConfigDefaults(t *testing.T) {
	for _, k := range []string{"HOST", "PORT", "Port", "ENVIRONMENT", "DB_DRIVER", "DB_PATH", "DATABASE_URL", "dbUrl"} {
		t.Setenv(k, "")
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Fatalf("expected Host 127.0.0.1, got %q", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected Port 8080, got %d", cfg.Port)
	}
	if cfg.DBDriver != "sqlite" {
		t.Fatalf("expected DBDriver sqlite, got %q", cfg.DBDriver)
	}
	if cfg.DBPath != "./data/pkm.db" {
		t.Fatalf("expected DBPath ./data/pkm.db, got %q", cfg.DBPath)
	}
}

func TestLoadConfigPortAndDBFallback(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("Port", "9090")
	t.Setenv("DATABASE_URL", "user:pass@tcp(127.0.0.1:3306)/test")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg.Port != 9090 {
		t.Fatalf("expected legacy Port=9090 to work, got %d", cfg.Port)
	}
	if cfg.DatabaseUrl != "user:pass@tcp(127.0.0.1:3306)/test" {
		t.Fatalf("expected DATABASE_URL override, got %q", cfg.DatabaseUrl)
	}

	t.Setenv("PORT", "8088")
	cfg2, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg2.Port != 8088 {
		t.Fatalf("expected PORT=8088 to take precedence over Port, got %d", cfg2.Port)
	}
}
