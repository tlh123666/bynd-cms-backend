package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr          string
	BYNDAPIBaseURL    string
	InternalToken     string
	JWTSecret         string
	AdminUsername     string
	AdminPassword     string
	AdminPasswordHash string
	AllowedOrigins    []string
	RequestTimeout    time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:          env("HTTP_ADDR", ":8082"),
		BYNDAPIBaseURL:    strings.TrimRight(env("BYND_API_BASE_URL", "http://179.198.108.246:8081"), "/"),
		InternalToken:     strings.TrimSpace(os.Getenv("CMS_INTERNAL_TOKEN")),
		JWTSecret:         env("CMS_JWT_SECRET", "development-only-change-this-secret"),
		AdminUsername:     env("CMS_ADMIN_USERNAME", "Admin"),
		AdminPassword:     env("CMS_ADMIN_PASSWORD", "123456"),
		AdminPasswordHash: strings.TrimSpace(os.Getenv("CMS_ADMIN_PASSWORD_HASH")),
		AllowedOrigins:    splitCSV(env("CMS_ALLOWED_ORIGINS", "http://localhost:3006")),
		RequestTimeout:    15 * time.Second,
	}
	if cfg.BYNDAPIBaseURL == "" {
		return Config{}, fmt.Errorf("BYND_API_BASE_URL is required")
	}
	if cfg.InternalToken == "" {
		return Config{}, fmt.Errorf("CMS_INTERNAL_TOKEN is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("CMS_JWT_SECRET must contain at least 32 characters")
	}
	if cfg.AdminUsername == "" || (cfg.AdminPassword == "" && cfg.AdminPasswordHash == "") {
		return Config{}, fmt.Errorf("CMS administrator credentials are required")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	items := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}
