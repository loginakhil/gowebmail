package config

import "time"

// Default returns the default configuration
func Default() *Config {
	return &Config{
		SMTP: SMTPConfig{
			Host:           "0.0.0.0",
			Port:           1025,
			MaxMessageSize: 10 * 1024 * 1024, // 10MB
			Timeout:        30 * time.Second,
		},
		HTTP: HTTPConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: "./data/gowebmail.db",
		},
		Retention: RetentionConfig{
			Enabled:         true,
			MaxAge:          7 * 24 * time.Hour, // 7 days
			MaxCount:        1000,
			CleanupInterval: 1 * time.Hour,
		},
		Web: WebConfig{
			Enabled: true,
			Auth: AuthConfig{
				Enabled:  false,
				Username: "admin",
				Password: "changeme",
			},
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		},
	}
}
