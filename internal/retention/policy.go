package retention

import (
	"context"
	"time"

	"github.com/rs/zerolog"

	"gowebmail/internal/config"
	"gowebmail/internal/storage"
)

// Manager handles retention policy enforcement
type Manager struct {
	config  *config.RetentionConfig
	storage storage.Storage
	logger  zerolog.Logger
	stop    chan struct{}
	done    chan struct{}
}

// NewManager creates a new retention policy manager
func NewManager(cfg *config.RetentionConfig, store storage.Storage, logger zerolog.Logger) *Manager {
	return &Manager{
		config:  cfg,
		storage: store,
		logger:  logger,
		stop:    make(chan struct{}),
		done:    make(chan struct{}),
	}
}

// Start starts the retention policy enforcement
func (m *Manager) Start(ctx context.Context) {
	defer close(m.done)

	if !m.config.Enabled {
		m.logger.Info().Msg("Retention policy disabled")
		return
	}

	m.logger.Info().
		Dur("max_age", m.config.MaxAge).
		Int("max_count", m.config.MaxCount).
		Dur("cleanup_interval", m.config.CleanupInterval).
		Msg("Starting retention policy manager")

	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	m.cleanup()

	for {
		select {
		case <-ticker.C:
			m.cleanup()
		case <-m.stop:
			m.logger.Info().Msg("Retention policy manager stopped")
			return
		case <-ctx.Done():
			m.logger.Info().Msg("Retention policy manager context cancelled")
			return
		}
	}
}

// Stop stops the retention policy manager
func (m *Manager) Stop() {
	close(m.stop)
	<-m.done
}

// cleanup performs the cleanup operation
func (m *Manager) cleanup() {
	m.logger.Debug().Msg("Running retention policy cleanup")

	// Delete old emails
	if m.config.MaxAge > 0 {
		before := time.Now().Add(-m.config.MaxAge)
		deleted, err := m.storage.DeleteOldEmails(before)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to delete old emails")
		} else if deleted > 0 {
			m.logger.Info().
				Int64("count", deleted).
				Time("before", before).
				Msg("Deleted old emails")
		}
	}

	// Delete excess emails
	if m.config.MaxCount > 0 {
		deleted, err := m.storage.DeleteExcessEmails(m.config.MaxCount)
		if err != nil {
			m.logger.Error().Err(err).Msg("Failed to delete excess emails")
		} else if deleted > 0 {
			m.logger.Info().
				Int64("count", deleted).
				Int("max_count", m.config.MaxCount).
				Msg("Deleted excess emails")
		}
	}
}
