package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	SMTP      SMTPConfig      `yaml:"smtp"`
	HTTP      HTTPConfig      `yaml:"http"`
	Storage   StorageConfig   `yaml:"storage"`
	Retention RetentionConfig `yaml:"retention"`
	Web       WebConfig       `yaml:"web"`
	Logging   LoggingConfig   `yaml:"logging"`
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	MaxMessageSize int64         `yaml:"max_message_size"`
	Timeout        time.Duration `yaml:"timeout"`
}

// HTTPConfig holds HTTP server configuration
type HTTPConfig struct {
	Host         string        `yaml:"host"`
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// RetentionConfig holds retention policy configuration
type RetentionConfig struct {
	Enabled         bool          `yaml:"enabled"`
	MaxAge          time.Duration `yaml:"max_age"`
	MaxCount        int           `yaml:"max_count"`
	CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

// WebConfig holds web interface configuration
type WebConfig struct {
	Enabled bool       `yaml:"enabled"`
	Auth    AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// Load loads configuration from file and applies environment variable overrides
func Load(path string) (*Config, error) {
	// Start with defaults
	cfg := Default()

	// Load from file if it exists
	if path != "" {
		if err := loadFromFile(path, cfg); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	return cfg, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			return nil
		}
		return err
	}

	return yaml.Unmarshal(data, cfg)
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(cfg *Config) {
	// SMTP overrides
	if v := os.Getenv("GOWEBMAIL_SMTP_HOST"); v != "" {
		cfg.SMTP.Host = v
	}
	if v := os.Getenv("GOWEBMAIL_SMTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.SMTP.Port = port
		}
	}

	// HTTP overrides
	if v := os.Getenv("GOWEBMAIL_HTTP_HOST"); v != "" {
		cfg.HTTP.Host = v
	}
	if v := os.Getenv("GOWEBMAIL_HTTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.HTTP.Port = port
		}
	}

	// Storage overrides
	if v := os.Getenv("GOWEBMAIL_STORAGE_PATH"); v != "" {
		cfg.Storage.Path = v
	}

	// Logging overrides
	if v := os.Getenv("GOWEBMAIL_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}

	// Web auth overrides
	if v := os.Getenv("GOWEBMAIL_WEB_AUTH_ENABLED"); v != "" {
		cfg.Web.Auth.Enabled = v == "true" || v == "1"
	}
	if v := os.Getenv("GOWEBMAIL_WEB_AUTH_USERNAME"); v != "" {
		cfg.Web.Auth.Username = v
	}
	if v := os.Getenv("GOWEBMAIL_WEB_AUTH_PASSWORD"); v != "" {
		cfg.Web.Auth.Password = v
	}
}
