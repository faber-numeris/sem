package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig_Defaults(t *testing.T) {
	for _, k := range []string{"CONFIG_FILE", "PORT", "ALLOWED_ORIGINS", "INSTANCE_URL", "DB_HOST", "PG_HOST", "DB_PORT", "PG_PORT", "DB_USER", "PG_USER", "DB_PASSWORD", "PG_PASSWORD", "DB_NAME", "PG_DB", "DB_SSLMODE"} {
		t.Setenv(k, "")
	}
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port() != 8080 {
		t.Fatalf("expected 8080, got %d", cfg.Port())
	}
	if len(cfg.AllowedOrigins()) != 2 {
		t.Fatalf("expected 2 origins, got %v", cfg.AllowedOrigins())
	}
	if cfg.InstanceURL() != "http://localhost:8080" {
		t.Fatalf("expected default instance url, got %s", cfg.InstanceURL())
	}
	if cfg.DBHost() != "localhost" || cfg.DBPort() != 5432 || cfg.DBUser() != "postgres" || cfg.DBPassword() != "postgres" || cfg.DBName() != "postgres" || cfg.DBSSLMode() != "disable" {
		t.Fatalf("unexpected db defaults: %+v", cfg)
	}
}

func TestNewConfig_EnvOverrides(t *testing.T) {
	t.Setenv("CONFIG_FILE", "")
	t.Setenv("PORT", "9090")
	t.Setenv("ALLOWED_ORIGINS", "http://example.com,https://example.org")
	t.Setenv("INSTANCE_URL", "")
	t.Setenv("DB_HOST", "db.local")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "sfuser")
	t.Setenv("DB_PASSWORD", "sfpass")
	t.Setenv("DB_NAME", "sfdb")
	t.Setenv("DB_SSLMODE", "require")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port() != 9090 {
		t.Fatalf("expected 9090, got %d", cfg.Port())
	}
	if cfg.InstanceURL() != "http://localhost:9090" {
		t.Fatalf("expected http://localhost:9090, got %s", cfg.InstanceURL())
	}
	t.Setenv("INSTANCE_URL", "https://custom.my.salesforce.com")
	cfg2, err := NewConfig()
	if err != nil || cfg2.InstanceURL() != "https://custom.my.salesforce.com" {
		t.Fatalf("expected custom instance URL, got %v %v", cfg2, err)
	}
	if cfg.DBHost() != "db.local" || cfg.DBPort() != 5433 || cfg.DBUser() != "sfuser" || cfg.DBPassword() != "sfpass" || cfg.DBName() != "sfdb" || cfg.DBSSLMode() != "require" {
		t.Fatalf("unexpected db overrides")
	}
}

func TestNewConfig_FileAndErrors(t *testing.T) {
	tmp := t.TempDir()
	validFile := filepath.Join(tmp, "config.yaml")
	content := `
service:
  port: 7070
  instance_url: "http://localhost:7070"
database:
  host: "yamlhost"
  port: 5434
`
	if err := os.WriteFile(validFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_FILE", validFile)
	t.Setenv("PORT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("PG_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("PG_PORT", "")
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port() != 7070 || cfg.DBHost() != "yamlhost" || cfg.DBPort() != 5434 {
		t.Fatalf("unexpected file values: port=%d host=%s dbport=%d", cfg.Port(), cfg.DBHost(), cfg.DBPort())
	}

	// Missing file
	t.Setenv("CONFIG_FILE", filepath.Join(tmp, "missing.yaml"))
	if _, err := NewConfig(); err == nil {
		t.Fatal("expected error for missing file")
	}

	// Invalid YAML
	badFile := filepath.Join(tmp, "bad.yaml")
	_ = os.WriteFile(badFile, []byte("service: [unclosed"), 0600)
	t.Setenv("CONFIG_FILE", badFile)
	if _, err := NewConfig(); err == nil {
		t.Fatal("expected error for invalid yaml")
	}

	// Invalid PORT
	t.Setenv("CONFIG_FILE", "")
	t.Setenv("PORT", "not-an-int")
	if _, err := NewConfig(); err == nil {
		t.Fatal("expected error for invalid PORT")
	}

	// Invalid DB_PORT
	t.Setenv("PORT", "8080")
	t.Setenv("DB_PORT", "not-an-int")
	if _, err := NewConfig(); err == nil {
		t.Fatal("expected error for invalid DB_PORT")
	}
}
