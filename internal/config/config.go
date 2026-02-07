package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	R2          R2Config
	KIE         KIEConfig
	OpenRouter  OpenRouterConfig
	Webhook     WebhookConfig
	CORS        CORSConfig
	Crypto      CryptoConfig
	YouTube     YouTubeConfig
	FrontendURL string // Frontend base URL for OAuth redirects (e.g. https://www.thinkclip.xyz)
}

// CORSConfig holds CORS-related configuration.
type CORSConfig struct {
	Origins []string // Comma-separated list of allowed origins
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
	BaseURL        string
	Secret         string        // Secret token for webhook authentication
	RateLimitRPS   int           // Rate limit requests per second
	RateLimitBurst int           // Rate limit burst size
	AllowedHosts   []string      // Allowed hosts for URL validation (SSRF prevention)
}

// CryptoConfig holds encryption-related configuration.
type CryptoConfig struct {
	EncryptionKey string // Base64-encoded 32-byte key for AES-256
}

// YouTubeConfig holds YouTube API configuration (optional).
type YouTubeConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
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
	viper.SetDefault("WEBHOOK_RATE_LIMIT_RPS", 10)
	viper.SetDefault("WEBHOOK_RATE_LIMIT_BURST", 20)
	viper.SetDefault("WEBHOOK_ALLOWED_HOSTS", "suno.ai,suno.com,audiopipe.suno.ai,cdn1.suno.ai,cdn2.suno.ai,kie.ai,cdn.kie.ai,storage.kie.ai,musicfile.kie.ai,s3.amazonaws.com,s3.us-east-1.amazonaws.com,s3.us-west-2.amazonaws.com,nanobananastorage.blob.core.windows.net,aiquickdraw.com")

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
			BaseURL:        viper.GetString("WEBHOOK_BASE_URL"),
			Secret:         viper.GetString("WEBHOOK_SECRET"),
			RateLimitRPS:   viper.GetInt("WEBHOOK_RATE_LIMIT_RPS"),
			RateLimitBurst: viper.GetInt("WEBHOOK_RATE_LIMIT_BURST"),
			AllowedHosts:   parseCommaSeparated(viper.GetString("WEBHOOK_ALLOWED_HOSTS")),
		},
		CORS: CORSConfig{
			Origins: parseCORSOrigins(viper.GetString("CORS_ORIGINS")),
		},
		Crypto: CryptoConfig{
			EncryptionKey: viper.GetString("ENCRYPTION_KEY"),
		},
		YouTube: YouTubeConfig{
			ClientID:     viper.GetString("YOUTUBE_CLIENT_ID"),
			ClientSecret: viper.GetString("YOUTUBE_CLIENT_SECRET"),
			RedirectURI:  viper.GetString("YOUTUBE_REDIRECT_URI"),
		},
		FrontendURL: strings.TrimRight(viper.GetString("FRONTEND_URL"), "/"),
	}

	return cfg, nil
}

// parseCORSOrigins parses comma-separated CORS origins string into a slice.
func parseCORSOrigins(originsStr string) []string {
	return parseCommaSeparated(originsStr)
}

// parseCommaSeparated parses comma-separated string into a slice.
func parseCommaSeparated(str string) []string {
	if str == "" {
		return []string{}
	}
	parts := strings.Split(str, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// Validate checks that all required configuration values are set.
// Returns an error describing all missing/invalid values.
func (c *Config) Validate() error {
	var errs []string

	if c.Database.URL == "" {
		errs = append(errs, "DATABASE_URL is required")
	}
	if c.Redis.URL == "" {
		errs = append(errs, "REDIS_URL is required")
	}
	if c.JWT.Secret == "" {
		errs = append(errs, "JWT_SECRET is required")
	} else if len(c.JWT.Secret) < 32 {
		errs = append(errs, "JWT_SECRET must be at least 32 characters")
	}
	if c.Crypto.EncryptionKey == "" {
		errs = append(errs, "ENCRYPTION_KEY is required")
	}

	// Webhook secret is required in production/staging
	if c.IsProduction() || c.IsStaging() {
		if c.Webhook.Secret == "" {
			errs = append(errs, "WEBHOOK_SECRET is required in production/staging")
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
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
