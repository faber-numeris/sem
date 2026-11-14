package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type IAppConfig interface {
	IServiceConfig
	IDatabaseConfig
}

type AppConfig struct {
	ServiceConfig  `yaml:"service"`
	DatabaseConfig `yaml:"database"`
}

func NewConfig() (IAppConfig, error) {
	cfg := &AppConfig{
		ServiceConfig: ServiceConfig{
			Port_:           8080,
			AllowedOrigins_: []string{"http://localhost:*", "https://*"},
			InstanceURL_:    "http://localhost:8080",
		},
		DatabaseConfig: DatabaseConfig{
			DBHost_:     "localhost",
			DBPort_:     5432,
			DBUser_:     "postgres",
			DBPassword_: "postgres",
			DBName_:     "postgres",
			DBSSLMode_:  "disable",
		},
	}

	if cfgPath := os.Getenv("CONFIG_FILE"); cfgPath != "" {
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", cfgPath, err)
		}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file %s: %w", cfgPath, err)
		}
	}

	if v := os.Getenv("PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT value %q: %w", v, err)
		}
		cfg.Port_ = port
	}
	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		cfg.AllowedOrigins_ = strings.Split(v, ",")
	}
	if v := os.Getenv("INSTANCE_URL"); v != "" {
		cfg.InstanceURL_ = v
	} else if cfg.InstanceURL_ == "http://localhost:8080" && cfg.Port_ != 8080 {
		cfg.InstanceURL_ = fmt.Sprintf("http://localhost:%d", cfg.Port_)
	}

	if v := firstEnv("DB_HOST", "PG_HOST"); v != "" {
		cfg.DBHost_ = v
	}
	if v := firstEnv("DB_PORT", "PG_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid DB_PORT value %q: %w", v, err)
		}
		cfg.DBPort_ = port
	}
	if v := firstEnv("DB_USER", "PG_USER"); v != "" {
		cfg.DBUser_ = v
	}
	if v := firstEnv("DB_PASSWORD", "PG_PASSWORD"); v != "" {
		cfg.DBPassword_ = v
	}
	if v := firstEnv("DB_NAME", "PG_DB"); v != "" {
		cfg.DBName_ = v
	}
	if v := os.Getenv("DB_SSLMODE"); v != "" {
		cfg.DBSSLMode_ = v
	}

	return cfg, nil
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
