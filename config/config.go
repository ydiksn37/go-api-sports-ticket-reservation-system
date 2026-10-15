package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

// アプリケーション全体の設定をまとめた構造体
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}

type ServerConfig struct {
	Port string
	Host string
}

// データベース接続に関する設定
type DatabaseConfig struct {
	URL string
}

// JWT 認証に関する設定
type AuthConfig struct {
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	CookieDomain string
	CookieSecure bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			URL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:55432/road_to_dena?sslmode=disable"),
		},
		Auth: AuthConfig{
			JWTSecret:    getEnv("JWT_SECRET", "local-dev-secret"),
			JWTIssuer:    getEnv("JWT_ISSUER", "road-to-dena"),
			JWTAudience:  getEnv("JWT_AUDIENCE", "road-to-dena-client"),
			AccessTTL:    15 * time.Minute,
			RefreshTTL:   7 * 24 * time.Hour,
			CookieDomain: getEnv("COOKIE_DOMAIN", ""),
			CookieSecure: false,
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
