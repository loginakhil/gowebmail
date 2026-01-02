package smtp

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/emersion/go-smtp"
	"github.com/rs/zerolog"

	"gowebmail/internal/config"
	"gowebmail/internal/email"
	"gowebmail/internal/storage"
)

// Server represents the SMTP server
type Server struct {
	config    *config.SMTPConfig
	storage   storage.Storage
	parser    *email.Parser
	logger    zerolog.Logger
	server    *smtp.Server
	onNewMail func(*storage.Email)
}

// NewServer creates a new SMTP server
func NewServer(cfg *config.SMTPConfig, store storage.Storage, logger zerolog.Logger) *Server {
	s := &Server{
		config:  cfg,
		storage: store,
		parser:  email.NewParser(),
		logger:  logger,
	}

	// Create SMTP server
	s.server = smtp.NewServer(s)
	s.server.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	s.server.Domain = "gowebmail.local"
	s.server.MaxMessageBytes = cfg.MaxMessageSize
	s.server.MaxRecipients = 100
	s.server.AllowInsecureAuth = true
	s.server.ReadTimeout = cfg.Timeout
	s.server.WriteTimeout = cfg.Timeout

	return s
}

// SetNewMailCallback sets the callback for new emails
func (s *Server) SetNewMailCallback(callback func(*storage.Email)) {
	s.onNewMail = callback
}

// Start starts the SMTP server
func (s *Server) Start() error {
	s.logger.Info().
		Str("addr", s.server.Addr).
		Msg("Starting SMTP server")
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the SMTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down SMTP server")
	return s.server.Shutdown(ctx)
}

// NewSession implements smtp.Backend interface
func (s *Server) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		server: s,
		logger: s.logger.With().
			Str("remote", c.Conn().RemoteAddr().String()).
			Logger(),
	}, nil
}

// Session represents an SMTP session
type Session struct {
	server *Server
	logger zerolog.Logger
	from   string
	to     []string
}

// AuthPlain implements smtp.Session interface (not used, auth disabled)
func (s *Session) AuthPlain(username, password string) error {
	return nil
}

// Mail implements smtp.Session interface
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.from = from
	s.logger.Debug().Str("from", from).Msg("MAIL FROM")
	return nil
}

// Rcpt implements smtp.Session interface
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.to = append(s.to, to)
	s.logger.Debug().Str("to", to).Msg("RCPT TO")
	return nil
}

// Data implements smtp.Session interface
func (s *Session) Data(r io.Reader) error {
	s.logger.Debug().Msg("Receiving email data")

	// Parse email
	email, err := s.server.parser.Parse(r)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to parse email")
		return fmt.Errorf("failed to parse email: %w", err)
	}

	// Set envelope data if not present in headers
	if email.From == "" {
		email.From = s.from
	}
	if len(email.To) == 0 {
		email.To = s.to
	}
	email.ReceivedAt = time.Now()

	// Save to storage
	id, err := s.server.storage.SaveEmail(email)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to save email")
		return fmt.Errorf("failed to save email: %w", err)
	}

	email.ID = id

	s.logger.Info().
		Int64("id", id).
		Str("from", email.From).
		Strs("to", email.To).
		Str("subject", email.Subject).
		Int64("size", email.Size).
		Msg("Email received and saved")

	// Notify callback
	if s.server.onNewMail != nil {
		go s.server.onNewMail(email)
	}

	return nil
}

// Reset implements smtp.Session interface
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

// Logout implements smtp.Session interface
func (s *Session) Logout() error {
	return nil
}
