package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gowebmail/internal/api"
	"gowebmail/internal/config"
	"gowebmail/internal/retention"
	"gowebmail/internal/smtp"
	"gowebmail/internal/storage"

	"github.com/rs/zerolog"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "gowebmail.yml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("GoWebMail %s\n", version)
		fmt.Printf("  Commit:   %s\n", commit)
		fmt.Printf("  Built:    %s\n", date)
		fmt.Printf("  Built by: %s\n", builtBy)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)
	logger.Info().
		Str("version", version).
		Str("commit", commit).
		Str("date", date).
		Msg("Starting GoWebMail")

	// Initialize storage
	store, err := storage.NewSQLiteStorage(cfg.Storage.Path, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize storage")
	}
	defer store.Close()

	// Create HTTP server
	httpServer := api.NewServer(cfg, store, logger)

	// Create SMTP server
	smtpServer := smtp.NewServer(&cfg.SMTP, store, logger)

	// Set callback for new emails to broadcast via WebSocket
	smtpServer.SetNewMailCallback(func(email *storage.Email) {
		httpServer.BroadcastNewEmail(email)
	})

	// Start retention policy manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.Retention.Enabled {
		retentionMgr := retention.NewManager(&cfg.Retention, store, logger)
		go retentionMgr.Start(ctx)
	}

	// Start servers in goroutines
	go func() {
		if err := smtpServer.Start(); err != nil {
			logger.Fatal().Err(err).Msg("SMTP server failed")
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil {
			logger.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	logger.Info().
		Int("smtp_port", cfg.SMTP.Port).
		Int("http_port", cfg.HTTP.Port).
		Msg("GoWebMail started successfully")

	// Wait for shutdown signal
	waitForShutdown(smtpServer, httpServer, logger)
}

// setupLogger configures the logger based on configuration
func setupLogger(cfg config.LoggingConfig) zerolog.Logger {
	// Set log level
	level := zerolog.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output io.Writer = os.Stdout
	if cfg.Output != "stdout" && cfg.Output != "" {
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			output = file
		}
	}

	// Configure format
	if cfg.Format == "text" {
		output = zerolog.ConsoleWriter{Out: output, TimeFormat: time.RFC3339}
	}

	return zerolog.New(output).With().Timestamp().Logger()
}

// waitForShutdown waits for a shutdown signal and gracefully shuts down servers
func waitForShutdown(smtpServer *smtp.Server, httpServer *api.Server, logger zerolog.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info().Msg("Shutdown signal received")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown servers gracefully
	logger.Info().Msg("Shutting down SMTP server...")
	if err := smtpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("SMTP server shutdown error")
	}

	logger.Info().Msg("Shutting down HTTP server...")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("HTTP server shutdown error")
	}

	logger.Info().Msg("Shutdown complete")
}
