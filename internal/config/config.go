package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server     ServerConfig
	Database   DatabaseConfig
	Redis      RedisConfig
	JWT        JWTConfig
	R2         R2Config
	KIE        KIEConfig
	OpenRouter OpenRouterConfig
	Webhook    WebhookConfig
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Port string
	Env  string // development, staging, production
}

// DatabaseConfig holds database-related configuration.
type DatabaseConfig struct {
	URL string // Neon PostgreSQL connection string
}

// RedisConfig holds Redis-related configuration for Asynq.
type RedisConfig struct {
	URL string
}

// JWTConfig holds JWT-related configuration.
type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

// R2Config holds Cloudflare R2-related configuration.
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

// KIEConfig holds KIE API configuration.
type KIEConfig struct {
	APIKey  string
	BaseURL string
}

// OpenRouterConfig holds OpenRouter API configuration.
type OpenRouterConfig struct {
	APIKey string
}

// WebhookConfig holds webhook-related configuration.
type WebhookConfig struct {
	BaseURL string
}

// Load reads configuration from environment variables and .env file.
func Load() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	// Read .env file if it exists (ignore error if not found)
	_ = viper.ReadInConfig()

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_ENV", "development")
	viper.SetDefault("JWT_EXPIRY", "24h")

	// Parse JWT expiry duration
	jwtExpiry, err := time.ParseDuration(viper.GetString("JWT_EXPIRY"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
			Env:  viper.GetString("SERVER_ENV"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
		Redis: RedisConfig{
			URL: viper.GetString("REDIS_URL"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("JWT_SECRET"),
			Expiry: jwtExpiry,
		},
		R2: R2Config{
			AccountID:       viper.GetString("R2_ACCOUNT_ID"),
			AccessKeyID:     viper.GetString("R2_ACCESS_KEY_ID"),
			SecretAccessKey: viper.GetString("R2_SECRET_ACCESS_KEY"),
			BucketName:      viper.GetString("R2_BUCKET_NAME"),
			PublicURL:       viper.GetString("R2_PUBLIC_URL"),
		},
		KIE: KIEConfig{
			APIKey:  viper.GetString("KIE_API_KEY"),
			BaseURL: viper.GetString("KIE_BASE_URL"),
		},
		OpenRouter: OpenRouterConfig{
			APIKey: viper.GetString("OPENROUTER_API_KEY"),
		},
		Webhook: WebhookConfig{
			BaseURL: viper.GetString("WEBHOOK_BASE_URL"),
		},
	}

	return cfg, nil
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// IsStaging returns true if the environment is staging.
func (c *Config) IsStaging() bool {
	return c.Server.Env == "staging"
}
