// Package config loads and validates the BFF configuration from the
// environment (12-factor). Validation is fail-fast: an invalid or missing
// security-critical value panics the process at startup rather than at
// request time.
package config

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config holds all runtime configuration. Secrets are redacted by LogValue so
// the whole struct can be logged safely at startup.
type Config struct {
	Env  string `env:"ENV" envDefault:"dev"`
	Port int    `env:"PORT" envDefault:"8080"`

	JWTSecret   string        `env:"JWT_SECRET,required"`
	JWTExpiry   time.Duration `env:"JWT_EXPIRY" envDefault:"1h"`
	JWTIssuer   string        `env:"JWT_ISSUER" envDefault:"cloud-log-access-service"`
	JWTAudience string        `env:"JWT_AUDIENCE" envDefault:"cloud-log-access-service"`

	CORSAllowedOrigins []string `env:"CORS_ALLOWED_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000"`

	// Rate limits, keyed on the real client IP (X-Real-IP behind nginx).
	RateLimitRPM     int `env:"RATE_LIMIT_RPM" envDefault:"120"`
	LoginRateLimitPM int `env:"LOGIN_RATE_LIMIT_PM" envDefault:"5"`

	AWSRegion          string   `env:"AWS_REGION" envDefault:"us-east-1"`
	AWSAccessKeyID     string   `env:"AWS_ACCESS_KEY_ID" envDefault:"test"`
	AWSSecretAccessKey string   `env:"AWS_SECRET_ACCESS_KEY" envDefault:"test"`
	S3InternalEndpoint string   `env:"S3_INTERNAL_ENDPOINT" envDefault:"http://localhost:4566"`
	S3PublicEndpoint   string   `env:"S3_PUBLIC_ENDPOINT" envDefault:"http://localhost:4566"`
	S3UsePathStyle     bool     `env:"S3_USE_PATH_STYLE" envDefault:"true"`
	S3Buckets          []string `env:"S3_BUCKETS" envSeparator:"," envDefault:"production-logs"`

	// GCS / fake-gcs-server (registered only when GCS_INTERNAL_ENDPOINT is set).
	GCSInternalEndpoint string   `env:"GCS_INTERNAL_ENDPOINT"`
	GCSPublicEndpoint   string   `env:"GCS_PUBLIC_ENDPOINT" envDefault:"http://localhost:4443"`
	GCSBuckets          []string `env:"GCS_BUCKETS" envSeparator:"," envDefault:"gcs-prod-logs"`

	// Azure Blob / Azurite (registered only when AZURE_INTERNAL_ENDPOINT is set).
	AzureAccount          string   `env:"AZURE_STORAGE_ACCOUNT" envDefault:"devstoreaccount1"`
	AzureKey              string   `env:"AZURE_STORAGE_KEY"`
	AzureInternalEndpoint string   `env:"AZURE_INTERNAL_ENDPOINT"`
	AzurePublicEndpoint   string   `env:"AZURE_PUBLIC_ENDPOINT" envDefault:"http://localhost:10000/devstoreaccount1"`
	AzureContainers       []string `env:"AZURE_CONTAINERS" envSeparator:"," envDefault:"app-logs-prod"`

	// Pre-signed URL TTL bounds.
	PresignMinTTL     time.Duration `env:"PRESIGN_MIN_TTL" envDefault:"60s"`
	PresignMaxTTL     time.Duration `env:"PRESIGN_MAX_TTL" envDefault:"15m"`
	PresignDefaultTTL time.Duration `env:"PRESIGN_DEFAULT_TTL" envDefault:"5m"`

	// HTTP server timeouts. WriteTimeout is intentionally absent: downloads
	// stream, so per-route context deadlines bound handlers instead.
	ReadHeaderTimeout time.Duration `env:"READ_HEADER_TIMEOUT" envDefault:"5s"`
	ReadTimeout       time.Duration `env:"READ_TIMEOUT" envDefault:"15s"`
	IdleTimeout       time.Duration `env:"IDLE_TIMEOUT" envDefault:"120s"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

// minJWTSecretBytes is the floor for HS256 signing material (256-bit security).
const minJWTSecretBytes = 32

// Load parses the environment into a Config and validates it. It returns an
// error (never a partially-valid Config) so callers can fail fast.
func Load() (*Config, error) {
	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	// Byte count, not rune count: HMAC is fed the raw bytes.
	if n := len([]byte(cfg.JWTSecret)); n < minJWTSecretBytes {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d bytes, got %d", minJWTSecretBytes, n)
	}
	// Fail fast: an Azure endpoint without a key boots fine but produces invalid
	// SAS signatures at request time.
	if cfg.AzureInternalEndpoint != "" && cfg.AzureKey == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_KEY is required when AZURE_INTERNAL_ENDPOINT is set")
	}
	return &cfg, nil
}

// IsProd reports whether the service is running in a production-like mode
// (drives JSON vs text logging).
func (c Config) IsProd() bool {
	return c.Env == "prod" || c.Env == "production"
}

// LogValue redacts the JWT secret so the config can be safely logged.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("env", c.Env),
		slog.Int("port", c.Port),
		slog.String("jwt_secret", "[REDACTED]"),
		slog.String("aws_secret_access_key", "[REDACTED]"),
		slog.String("azure_storage_key", "[REDACTED]"),
		slog.Duration("jwt_expiry", c.JWTExpiry),
		slog.String("jwt_issuer", c.JWTIssuer),
		slog.Any("cors_allowed_origins", c.CORSAllowedOrigins),
	)
}
