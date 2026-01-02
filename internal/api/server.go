package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"gowebmail/internal/config"
	"gowebmail/internal/storage"
)

// Server represents the HTTP API server
type Server struct {
	config  *config.Config
	storage storage.Storage
	router  *mux.Router
	logger  zerolog.Logger
	wsHub   *WebSocketHub
	server  *http.Server
}

// NewServer creates a new HTTP API server
func NewServer(cfg *config.Config, store storage.Storage, logger zerolog.Logger) *Server {
	s := &Server{
		config:  cfg,
		storage: store,
		router:  mux.NewRouter(),
		logger:  logger,
		wsHub:   NewWebSocketHub(logger),
	}

	s.setupRoutes()
	s.setupMiddleware()

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port),
		Handler:      s.router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Email endpoints
	api.HandleFunc("/emails", s.handleListEmails).Methods("GET")
	api.HandleFunc("/emails/{id:[0-9]+}", s.handleGetEmail).Methods("GET")
	api.HandleFunc("/emails/{id:[0-9]+}", s.handleDeleteEmail).Methods("DELETE")
	api.HandleFunc("/emails", s.handleDeleteAllEmails).Methods("DELETE")
	api.HandleFunc("/emails/search", s.handleSearchEmails).Methods("GET")
	api.HandleFunc("/emails/{id:[0-9]+}/raw", s.handleGetEmailRaw).Methods("GET")
	api.HandleFunc("/emails/{id:[0-9]+}/html", s.handleGetEmailHTML).Methods("GET")
	api.HandleFunc("/emails/{id:[0-9]+}/attachments/{aid:[0-9]+}", s.handleGetAttachment).Methods("GET")

	// Stats endpoint
	api.HandleFunc("/stats", s.handleGetStats).Methods("GET")

	// Health check
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// WebSocket
	s.router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		s.wsHub.ServeWS(w, r)
	})

	// Static files (web UI)
	s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))
}

// setupMiddleware configures middleware
func (s *Server) setupMiddleware() {
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
	s.router.Use(s.recoveryMiddleware)

	// Optional auth middleware
	if s.config.Web.Auth.Enabled {
		s.router.Use(s.authMiddleware)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.wsHub.Run()

	s.logger.Info().
		Str("addr", s.server.Addr).
		Msg("Starting HTTP server")

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info().Msg("Shutting down HTTP server")
	s.wsHub.Shutdown()
	return s.server.Shutdown(ctx)
}

// BroadcastNewEmail broadcasts a new email notification via WebSocket
func (s *Server) BroadcastNewEmail(email *storage.Email) {
	s.wsHub.Broadcast(&WebSocketMessage{
		Type: "email.new",
		Data: map[string]interface{}{
			"id":         email.ID,
			"from":       email.From,
			"to":         email.To,
			"subject":    email.Subject,
			"receivedAt": email.ReceivedAt,
		},
	})
}
