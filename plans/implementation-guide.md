# GoWebMail - Implementation Guide

## Development Phases

This guide breaks down the implementation into logical phases that can be executed sequentially. Each phase builds upon the previous one and results in a working, testable component.

---

## Phase 1: Project Foundation

### 1.1 Directory Structure Setup

Create the complete project structure:

```
gowebmail/
├── cmd/
│   └── gowebmail/
│       └── main.go
├── internal/
│   ├── config/
│   ├── smtp/
│   ├── storage/
│   ├── api/
│   ├── email/
│   └── retention/
├── web/
│   ├── css/
│   ├── js/
│   └── assets/
├── docs/
├── docker/
├── configs/
├── scripts/
└── data/
```

### 1.2 Go Module Configuration

Update [`go.mod`](go.mod) with required dependencies:

```go
module gowebmail

go 1.25

require (
    github.com/foxcpp/maddy v0.7.1
    github.com/mattn/go-sqlite3 v1.14.18
    github.com/gorilla/mux v1.8.1
    github.com/gorilla/websocket v1.5.1
    github.com/microcosm-cc/bluemonday v1.0.26
    github.com/rs/zerolog v1.31.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/emersion/go-message v0.18.0
)
```

### 1.3 Configuration System

**File**: `internal/config/config.go`

Define configuration structures:

```go
type Config struct {
    SMTP      SMTPConfig      `yaml:"smtp"`
    HTTP      HTTPConfig      `yaml:"http"`
    Storage   StorageConfig   `yaml:"storage"`
    Retention RetentionConfig `yaml:"retention"`
    Web       WebConfig       `yaml:"web"`
    Logging   LoggingConfig   `yaml:"logging"`
}

type SMTPConfig struct {
    Host           string        `yaml:"host"`
    Port           int           `yaml:"port"`
    MaxMessageSize int64         `yaml:"max_message_size"`
    Timeout        time.Duration `yaml:"timeout"`
}

type HTTPConfig struct {
    Host         string        `yaml:"host"`
    Port         int           `yaml:"port"`
    ReadTimeout  time.Duration `yaml:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout"`
}

type StorageConfig struct {
    Type string `yaml:"type"`
    Path string `yaml:"path"`
}

type RetentionConfig struct {
    Enabled         bool          `yaml:"enabled"`
    MaxAge          time.Duration `yaml:"max_age"`
    MaxCount        int           `yaml:"max_count"`
    CleanupInterval time.Duration `yaml:"cleanup_interval"`
}

type WebConfig struct {
    Enabled bool       `yaml:"enabled"`
    Auth    AuthConfig `yaml:"auth"`
}

type AuthConfig struct {
    Enabled  bool   `yaml:"enabled"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

type LoggingConfig struct {
    Level  string `yaml:"level"`
    Format string `yaml:"format"`
    Output string `yaml:"output"`
}
```

**File**: `internal/config/defaults.go`

Provide sensible defaults:

```go
func DefaultConfig() *Config {
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
            MaxAge:          7 * 24 * time.Hour,
            MaxCount:        1000,
            CleanupInterval: 1 * time.Hour,
        },
        Web: WebConfig{
            Enabled: true,
            Auth: AuthConfig{
                Enabled: false,
            },
        },
        Logging: LoggingConfig{
            Level:  "info",
            Format: "json",
            Output: "stdout",
        },
    }
}
```

**Deliverable**: Configuration loading from YAML file with environment variable overrides

---

## Phase 2: Storage Layer

### 2.1 Data Models

**File**: `internal/storage/models.go`

```go
type Email struct {
    ID           int64
    MessageID    string
    From         string
    To           []string
    CC           []string
    BCC          []string
    Subject      string
    BodyPlain    string
    BodyHTML     string
    Headers      map[string][]string
    Attachments  []AttachmentMeta
    Size         int64
    ReceivedAt   time.Time
    Read         bool
}

type AttachmentMeta struct {
    ID          int64
    Filename    string
    ContentType string
    Size        int64
}

type Attachment struct {
    AttachmentMeta
    Data []byte
}

type EmailFilter struct {
    From    string
    To      string
    Subject string
    Since   *time.Time
    Until   *time.Time
}
```

### 2.2 Storage Interface

**File**: `internal/storage/storage.go`

```go
type Storage interface {
    // Email operations
    SaveEmail(email *Email) (int64, error)
    GetEmail(id int64) (*Email, error)
    ListEmails(filter *EmailFilter, limit, offset int) ([]*Email, int64, error)
    SearchEmails(query string, limit, offset int) ([]*Email, int64, error)
    DeleteEmail(id int64) error
    DeleteAllEmails() error
    GetEmailCount() (int64, error)
    
    // Attachment operations
    GetAttachment(id int64) (*Attachment, error)
    
    // Retention operations
    DeleteOldEmails(before time.Time) (int64, error)
    DeleteExcessEmails(maxCount int) (int64, error)
    
    // Lifecycle
    Close() error
}
```

### 2.3 SQLite Implementation

**File**: `internal/storage/sqlite.go`

Key implementation points:

1. **Database Initialization**:
   - Create tables if not exist
   - Create indexes
   - Enable WAL mode for better concurrency
   - Set pragmas for performance

2. **Email Storage**:
   - Store email metadata in `emails` table
   - Store attachments in `attachments` table
   - Use transactions for atomicity
   - JSON encoding for arrays and maps

3. **Search Implementation**:
   - Option 1: SQLite FTS5 virtual table
   - Option 2: Simple LIKE queries for basic search
   - Recommend FTS5 for better performance

4. **Query Optimization**:
   - Use prepared statements
   - Implement connection pooling
   - Add appropriate indexes

**File**: `internal/storage/migrations.go`

```go
const schema = `
CREATE TABLE IF NOT EXISTS emails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    message_id TEXT UNIQUE,
    from_address TEXT NOT NULL,
    to_addresses TEXT NOT NULL,
    cc_addresses TEXT,
    bcc_addresses TEXT,
    subject TEXT,
    body_plain TEXT,
    body_html TEXT,
    headers TEXT NOT NULL,
    size INTEGER,
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    read BOOLEAN DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_emails_from ON emails(from_address);
CREATE INDEX IF NOT EXISTS idx_emails_received ON emails(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_emails_subject ON emails(subject);

CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    content_type TEXT,
    size INTEGER,
    data BLOB,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_attachments_email ON attachments(email_id);

-- FTS5 virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS emails_fts USING fts5(
    subject,
    from_address,
    to_addresses,
    body_plain,
    content='emails',
    content_rowid='id'
);

-- Triggers to keep FTS table in sync
CREATE TRIGGER IF NOT EXISTS emails_ai AFTER INSERT ON emails BEGIN
    INSERT INTO emails_fts(rowid, subject, from_address, to_addresses, body_plain)
    VALUES (new.id, new.subject, new.from_address, new.to_addresses, new.body_plain);
END;

CREATE TRIGGER IF NOT EXISTS emails_ad AFTER DELETE ON emails BEGIN
    DELETE FROM emails_fts WHERE rowid = old.id;
END;
`
```

**Deliverable**: Fully functional storage layer with SQLite backend

---

## Phase 3: Email Parsing

### 3.1 Email Parser

**File**: `internal/email/parser.go`

Use `github.com/emersion/go-message` for MIME parsing:

```go
type Parser struct{}

func (p *Parser) Parse(r io.Reader) (*storage.Email, error) {
    // Parse MIME message
    // Extract headers
    // Parse multipart bodies
    // Extract attachments
    // Calculate size
    // Return Email struct
}

func (p *Parser) parseHeaders(header mail.Header) map[string][]string
func (p *Parser) parseBody(entity *message.Entity) (plain, html string, attachments []storage.Attachment)
func (p *Parser) extractAttachment(entity *message.Entity) (*storage.Attachment, error)
```

Key features:
- Handle multipart/alternative (plain + HTML)
- Handle multipart/mixed (attachments)
- Decode quoted-printable and base64
- Extract inline images
- Handle nested multipart

### 3.2 HTML Sanitizer

**File**: `internal/email/sanitizer.go`

Use `github.com/microcosm-cc/bluemonday`:

```go
type Sanitizer struct {
    policy *bluemonday.Policy
}

func NewSanitizer() *Sanitizer {
    p := bluemonday.UGCPolicy()
    
    // Allow safe HTML tags
    p.AllowElements("p", "br", "strong", "em", "u", "h1", "h2", "h3", "h4", "h5", "h6")
    p.AllowElements("ul", "ol", "li", "blockquote", "pre", "code")
    p.AllowElements("table", "thead", "tbody", "tr", "th", "td")
    p.AllowElements("a", "img")
    
    // Allow safe attributes
    p.AllowAttrs("href").OnElements("a")
    p.AllowAttrs("src", "alt", "title").OnElements("img")
    p.AllowAttrs("class").Globally()
    
    // Block external resources
    p.RequireNoReferrerOnLinks(true)
    
    return &Sanitizer{policy: p}
}

func (s *Sanitizer) Sanitize(html string) string {
    return s.policy.Sanitize(html)
}
```

**Deliverable**: Email parsing and sanitization utilities

---

## Phase 4: SMTP Server (Maddy Integration)

### 4.1 Research Maddy Integration

After researching Maddy's architecture, there are two approaches:

**Option A: Custom SMTP Server** (Recommended)
- Build lightweight SMTP server using Go's `net/smtp` or `emersion/go-smtp`
- Simpler, more control, easier to maintain
- Maddy is complex and designed for production mail servers

**Option B: Maddy Integration**
- Use Maddy as library with custom delivery module
- More complex, but leverages Maddy's SMTP implementation

**Recommendation**: Use Option A with `github.com/emersion/go-smtp`

### 4.2 SMTP Server Implementation

**File**: `internal/smtp/server.go`

```go
type Server struct {
    config  *config.SMTPConfig
    storage storage.Storage
    parser  *email.Parser
    logger  zerolog.Logger
    server  *smtp.Server
}

func NewServer(cfg *config.SMTPConfig, storage storage.Storage, logger zerolog.Logger) *Server {
    s := &Server{
        config:  cfg,
        storage: storage,
        parser:  email.NewParser(),
        logger:  logger,
    }
    
    // Create SMTP server
    s.server = smtp.NewServer(s)
    s.server.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
    s.server.Domain = "gowebmail.local"
    s.server.MaxMessageBytes = int(cfg.MaxMessageSize)
    s.server.MaxRecipients = 100
    s.server.AllowInsecureAuth = true
    s.server.AuthDisabled = true
    
    return s
}

func (s *Server) Start() error {
    s.logger.Info().
        Str("addr", s.server.Addr).
        Msg("Starting SMTP server")
    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info().Msg("Shutting down SMTP server")
    return s.server.Shutdown(ctx)
}
```

**File**: `internal/smtp/handler.go`

```go
// Implement smtp.Backend interface
func (s *Server) NewSession(c *smtp.Conn) (smtp.Session, error) {
    return &Session{
        server: s,
        logger: s.logger,
    }, nil
}

type Session struct {
    server *Server
    logger zerolog.Logger
    from   string
    to     []string
}

func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
    s.from = from
    s.logger.Debug().Str("from", from).Msg("MAIL FROM")
    return nil
}

func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
    s.to = append(s.to, to)
    s.logger.Debug().Str("to", to).Msg("RCPT TO")
    return nil
}

func (s *Session) Data(r io.Reader) error {
    // Parse email
    email, err := s.server.parser.Parse(r)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to parse email")
        return err
    }
    
    // Set envelope data
    email.From = s.from
    email.To = s.to
    email.ReceivedAt = time.Now()
    
    // Save to storage
    id, err := s.server.storage.SaveEmail(email)
    if err != nil {
        s.logger.Error().Err(err).Msg("Failed to save email")
        return err
    }
    
    s.logger.Info().
        Int64("id", id).
        Str("from", s.from).
        Strs("to", s.to).
        Str("subject", email.Subject).
        Msg("Email received")
    
    // Broadcast to WebSocket clients (via channel)
    s.server.notifyNewEmail(email)
    
    return nil
}

func (s *Session) Reset() {
    s.from = ""
    s.to = nil
}

func (s *Session) Logout() error {
    return nil
}
```

**Deliverable**: Working SMTP server that accepts and stores emails

---

## Phase 5: REST API

### 5.1 HTTP Server Setup

**File**: `internal/api/server.go`

```go
type Server struct {
    config   *config.HTTPConfig
    storage  storage.Storage
    router   *mux.Router
    logger   zerolog.Logger
    wsHub    *WebSocketHub
    server   *http.Server
}

func NewServer(cfg *config.HTTPConfig, storage storage.Storage, logger zerolog.Logger) *Server {
    s := &Server{
        config:  cfg,
        storage: storage,
        router:  mux.NewRouter(),
        logger:  logger,
        wsHub:   NewWebSocketHub(logger),
    }
    
    s.setupRoutes()
    s.setupMiddleware()
    
    s.server = &http.Server{
        Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        Handler:      s.router,
        ReadTimeout:  cfg.ReadTimeout,
        WriteTimeout: cfg.WriteTimeout,
    }
    
    return s
}

func (s *Server) Start() error {
    // Start WebSocket hub
    go s.wsHub.Run()
    
    s.logger.Info().
        Str("addr", s.server.Addr).
        Msg("Starting HTTP server")
    
    return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
    s.logger.Info().Msg("Shutting down HTTP server")
    s.wsHub.Shutdown()
    return s.server.Shutdown(ctx)
}
```

### 5.2 Routes

**File**: `internal/api/routes.go`

```go
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
    s.router.HandleFunc("/ws", s.handleWebSocket)
    
    // Static files (web UI)
    s.router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))
}
```

### 5.3 Middleware

**File**: `internal/api/middleware.go`

```go
func (s *Server) setupMiddleware() {
    s.router.Use(s.loggingMiddleware)
    s.router.Use(s.corsMiddleware)
    s.router.Use(s.recoveryMiddleware)
    
    // Optional auth middleware
    if s.config.Auth.Enabled {
        s.router.Use(s.authMiddleware)
    }
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap response writer to capture status code
        wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}
        
        next.ServeHTTP(wrapped, r)
        
        s.logger.Info().
            Str("method", r.Method).
            Str("path", r.URL.Path).
            Int("status", wrapped.statusCode).
            Dur("duration", time.Since(start)).
            Msg("HTTP request")
    })
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

### 5.4 API Handlers

**File**: `internal/api/handlers.go`

```go
type APIResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   *APIError   `json:"error,omitempty"`
}

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

func (s *Server) handleListEmails(w http.ResponseWriter, r *http.Request) {
    // Parse query parameters
    limit := parseIntParam(r, "limit", 50, 1, 100)
    offset := parseIntParam(r, "offset", 0, 0, math.MaxInt)
    
    // Build filter
    filter := &storage.EmailFilter{
        From:    r.URL.Query().Get("from"),
        To:      r.URL.Query().Get("to"),
        Subject: r.URL.Query().Get("subject"),
    }
    
    // Parse date filters
    if since := r.URL.Query().Get("since"); since != "" {
        t, _ := time.Parse(time.RFC3339, since)
        filter.Since = &t
    }
    
    // Get emails
    emails, total, err := s.storage.ListEmails(filter, limit, offset)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
        return
    }
    
    s.sendSuccess(w, map[string]interface{}{
        "emails": emails,
        "total":  total,
        "limit":  limit,
        "offset": offset,
    })
}

func (s *Server) handleGetEmail(w http.ResponseWriter, r *http.Request) {
    id := parseIDParam(r)
    
    email, err := s.storage.GetEmail(id)
    if err != nil {
        if errors.Is(err, storage.ErrNotFound) {
            s.sendError(w, http.StatusNotFound, "NOT_FOUND", "Email not found")
        } else {
            s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
        }
        return
    }
    
    s.sendSuccess(w, email)
}

func (s *Server) handleDeleteEmail(w http.ResponseWriter, r *http.Request) {
    id := parseIDParam(r)
    
    err := s.storage.DeleteEmail(id)
    if err != nil {
        s.sendError(w, http.StatusInternalServerError, "STORAGE_ERROR", err.Error())
        return
    }
    
    // Notify WebSocket clients
    s.wsHub.Broadcast(&WebSocketMessage{
        Type: "email.deleted",
        Data: map[string]interface{}{"id": id},
    })
    
    s.sendSuccess(w, map[string]interface{}{"deleted": id})
}

func (s *Server) sendSuccess(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(APIResponse{
        Success: true,
        Data:    data,
    })
}

func (s *Server) sendError(w http.ResponseWriter, status int, code, message string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
        },
    })
}
```

**Deliverable**: Complete REST API with all endpoints

---

## Phase 6: WebSocket Support

### 6.1 WebSocket Hub

**File**: `internal/api/websocket.go`

```go
type WebSocketHub struct {
    clients    map[*WebSocketClient]bool
    broadcast  chan *WebSocketMessage
    register   chan *WebSocketClient
    unregister chan *WebSocketClient
    logger     zerolog.Logger
    mu         sync.RWMutex
}

type WebSocketClient struct {
    hub  *WebSocketHub
    conn *websocket.Conn
    send chan *WebSocketMessage
}

type WebSocketMessage struct {
    Type string                 `json:"type"`
    Data map[string]interface{} `json:"data"`
}

func NewWebSocketHub(logger zerolog.Logger) *WebSocketHub {
    return &WebSocketHub{
        clients:    make(map[*WebSocketClient]bool),
        broadcast:  make(chan *WebSocketMessage, 256),
        register:   make(chan *WebSocketClient),
        unregister: make(chan *WebSocketClient),
        logger:     logger,
    }
}

func (h *WebSocketHub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
            h.logger.Debug().Msg("WebSocket client connected")
            
        case client := <-h.unregister:
            h.mu.Lock()
            if _, ok := h.clients[client]; ok {
                delete(h.clients, client)
                close(client.send)
            }
            h.mu.Unlock()
            h.logger.Debug().Msg("WebSocket client disconnected")
            
        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                    delete(h.clients, client)
                }
            }
            h.mu.RUnlock()
        }
    }
}

func (h *WebSocketHub) Broadcast(message *WebSocketMessage) {
    h.broadcast <- message
}
```

### 6.2 WebSocket Handler

```go
var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins for development
    },
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        s.logger.Error().Err(err).Msg("WebSocket upgrade failed")
        return
    }
    
    client := &WebSocketClient{
        hub:  s.wsHub,
        conn: conn,
        send: make(chan *WebSocketMessage, 256),
    }
    
    client.hub.register <- client
    
    // Start goroutines for reading and writing
    go client.writePump()
    go client.readPump()
}

func (c *WebSocketClient) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
        return nil
    })
    
    for {
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }
}

func (c *WebSocketClient) writePump() {
    ticker := time.NewTicker(30 * time.Second)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()
    
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteJSON(message); err != nil {
                return
            }
            
        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}
```

**Deliverable**: Real-time WebSocket communication

---

## Phase 7: Frontend Development

### 7.1 HTML Structure

**File**: `web/index.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GoWebMail - Email Testing Tool</title>
    <link rel="stylesheet" href="/css/main.css">
    <link rel="stylesheet" href="/css/email-list.css">
    <link rel="stylesheet" href="/css/email-preview.css">
</head>
<body>
    <div class="app">
        <!-- Header -->
        <header class="header">
            <h1>GoWebMail</h1>
            <div class="stats">
                <span id="email-count">0 emails</span>
            </div>
        </header>
        
        <!-- Toolbar -->
        <div class="toolbar">
            <div class="search-bar">
                <input type="text" id="search-input" placeholder="Search emails...">
                <button id="search-btn">Search</button>
                <button id="clear-filters-btn">Clear</button>
            </div>
            <div class="actions">
                <button id="refresh-btn">Refresh</button>
                <button id="delete-selected-btn">Delete Selected</button>
                <button id="delete-all-btn">Delete All</button>
            </div>
        </div>
        
        <!-- Main content -->
        <div class="main-content">
            <!-- Email list -->
            <div class="email-list-container">
                <div id="email-list" class="email-list"></div>
                <div id="loading" class="loading" style="display: none;">Loading...</div>
                <div id="empty-state" class="empty-state">No emails yet</div>
            </div>
            
            <!-- Email preview -->
            <div class="email-preview-container">
                <div id="email-preview" class="email-preview">
                    <div class="preview-empty">Select an email to view</div>
                </div>
            </div>
        </div>
    </div>
    
    <script type="module" src="/js/app.js"></script>
</body>
</html>
```

### 7.2 JavaScript Architecture

**File**: `web/js/app.js`

```javascript
import { APIClient } from './api.js';
import { WebSocketClient } from './websocket.js';
import { EmailList } from './email-list.js';
import { EmailPreview } from './email-preview.js';
import { SearchBar } from './search.js';

class App {
    constructor() {
        this.api = new APIClient('/api');
        this.ws = new WebSocketClient('/ws');
        this.emailList = new EmailList(this.api);
        this.emailPreview = new EmailPreview(this.api);
        this.searchBar = new SearchBar();
        
        this.init();
    }
    
    init() {
        // Setup event listeners
        this.setupEventListeners();
        
        // Connect WebSocket
        this.ws.connect();
        this.ws.on('email.new', (data) => this.handleNewEmail(data));
        this.ws.on('email.deleted', (data) => this.handleEmailDeleted(data));
        this.ws.on('emails.cleared', () => this.handleEmailsCleared());
        
        // Load initial emails
        this.loadEmails();
    }
    
    setupEventListeners() {
        document.getElementById('refresh-btn').addEventListener('click', () => this.loadEmails());
        document.getElementById('delete-all-btn').addEventListener('click', () => this.deleteAllEmails());
        document.getElementById('search-btn').addEventListener('click', () => this.search());
        
        this.emailList.on('select', (email) => this.emailPreview.show(email));
    }
    
    async loadEmails() {
        const emails = await this.api.listEmails();
        this.emailList.render(emails);
        this.updateStats();
    }
    
    async search() {
        const query = this.searchBar.getQuery();
        const emails = await this.api.searchEmails(query);
        this.emailList.render(emails);
    }
    
    handleNewEmail(data) {
        this.emailList.prependEmail(data);
        this.updateStats();
    }
    
    async updateStats() {
        const stats = await this.api.getStats();
        document.getElementById('email-count').textContent = `${stats.count} emails`;
    }
}

// Initialize app
new App();
```

**File**: `web/js/api.js`

```javascript
export class APIClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
    }
    
    async listEmails(params = {}) {
        const query = new URLSearchParams(params).toString();
        const response = await fetch(`${this.baseURL}/emails?${query}`);
        const data = await response.json();
        return data.success ? data.data.emails : [];
    }
    
    async getEmail(id) {
        const response = await fetch(`${this.baseURL}/emails/${id}`);
        const data = await response.json();
        return data.success ? data.data : null;
    }
    
    async deleteEmail(id) {
        const response = await fetch(`${this.baseURL}/emails/${id}`, {
            method: 'DELETE'
        });
        return response.ok;
    }
    
    async deleteAllEmails() {
        const response = await fetch(`${this.baseURL}/emails`, {
            method: 'DELETE'
        });
        return response.ok;
    }
    
    async searchEmails(query) {
        const response = await fetch(`${this.baseURL}/emails/search?q=${encodeURIComponent(query)}`);
        const data = await response.json();
        return data.success ? data.data.emails : [];
    }
    
    async getStats() {
        const response = await fetch(`${this.baseURL}/stats`);
        const data = await response.json();
        return data.success ? data.data : {};
    }
}
```

**File**: `web/js/websocket.js`

```javascript
export class WebSocketClient {
    constructor(url) {
        this.url = url;
        this.ws = null;
        this.listeners = {};
        this.reconnectDelay = 1000;
    }
    
    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsURL = `${protocol}//${window.location.host}${this.url}`;
        
        this.ws = new WebSocket(wsURL);
        
        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectDelay = 1000;
        };
        
        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.emit(message.type, message.data);
        };
        
        this.ws.onclose = () => {
            console.log('WebSocket disconnected, reconnecting...');
            setTimeout(() => this.connect(), this.reconnectDelay);
            this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };
    }
    
    on(event, callback) {
        if (!this.listeners[event]) {
            this.listeners[event] = [];
        }
        this.listeners[event].push(callback);
    }
    
    emit(event, data) {
        if (this.listeners[event]) {
            this.listeners[event].forEach(callback => callback(data));
        }
    }
}
```

### 7.3 CSS Styling

**File**: `web/css/main.css`

Modern, responsive design with:
- CSS Grid for layout
- Flexbox for components
- CSS variables for theming
- Responsive breakpoints
- Clean, minimal design

**Deliverable**: Complete, functional web interface

---

## Phase 8: Integration & Polish

### 8.1 Main Application

**File**: `cmd/gowebmail/main.go`

```go
func main() {
    // Load configuration
    cfg := loadConfig()
    
    // Setup logger
    logger := setupLogger(cfg.Logging)
    
    // Initialize storage
    storage, err := storage.NewSQLiteStorage(cfg.Storage.Path, logger)
    if err != nil {
        logger.Fatal().Err(err).Msg("Failed to initialize storage")
    }
    defer storage.Close()
    
    // Create SMTP server
    smtpServer := smtp.NewServer(&cfg.SMTP, storage, logger)
    
    // Create HTTP server
    httpServer := api.NewServer(&cfg.HTTP, storage, logger)
    
    // Start retention policy enforcer
    if cfg.Retention.Enabled {
        retentionMgr := retention.NewManager(&cfg.Retention, storage, logger)
        go retentionMgr.Start()
    }
    
    // Start servers
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
    
    // Wait for shutdown signal
    waitForShutdown(smtpServer, httpServer, logger)
}

func waitForShutdown(smtpServer *smtp.Server, httpServer *api.Server, logger zerolog.Logger) {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
    
    <-sigChan
    logger.Info().Msg("Shutdown signal received")
    
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    // Shutdown servers gracefully
    if err := smtpServer.Shutdown(ctx); err != nil {
        logger.Error().Err(err).Msg("SMTP server shutdown error")
    }
    
    if err := httpServer.Shutdown(ctx); err != nil {
        logger.Error().Err(err).Msg("HTTP server shutdown error")
    }
    
    logger.Info().Msg("Shutdown complete")
}
```

### 8.2 Retention Policy

**File**: `internal/retention/policy.go`

```go
type Manager struct {
    config  *config.RetentionConfig
    storage storage.Storage
    logger  zerolog.Logger
    stop    chan struct{}
}

func (m *Manager) Start() {
    ticker := time.NewTicker(m.config.CleanupInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            m.cleanup()
        case <-m.stop:
            return
        }
    }
}

func (m *Manager) cleanup() {
    // Delete old emails
    if m.config.MaxAge > 0 {
        before := time.Now().Add(-m.config.MaxAge)
        deleted, err := m.storage.DeleteOldEmails(before)
        if err != nil {
            m.logger.Error().Err(err).Msg("Failed to delete old emails")
        } else if deleted > 0 {
            m.logger.Info().Int64("count", deleted).Msg("Deleted old emails")
        }
    }
    
    // Delete excess emails
    if m.config.MaxCount > 0 {
        deleted, err := m.storage.DeleteExcessEmails(m.config.MaxCount)
        if err != nil {
            m.logger.Error().Err(err).Msg("Failed to delete excess emails")
        } else if deleted > 0 {
            m.logger.Info().Int64("count", deleted).Msg("Deleted excess emails")
        }
    }
}
```

**Deliverable**: Complete, integrated application

---

## Phase 9: Docker & Documentation

### 9.1 Dockerfile

**File**: `docker/Dockerfile`

Multi-stage build for minimal image size

### 9.2 Docker Compose

**File**: `docker/docker-compose.yml`

Easy deployment configuration

### 9.3 Documentation

**File**: `README.md` - Project overview and quick start
**File**: `docs/API.md` - Complete API documentation
**File**: `docs/CONFIGURATION.md` - Configuration guide
**File**: `docs/USAGE.md` - Usage examples

**Deliverable**: Production-ready deployment and documentation

---

## Testing Strategy

### Unit Tests
- Storage layer: Test all CRUD operations
- Email parser: Test various email formats
- API handlers: Test request/response handling

### Integration Tests
- SMTP flow: Send email, verify storage
- API flow: Create, read, delete emails
- WebSocket: Test real-time updates

### Manual Testing
- Send test emails from various clients
- Test web interface in different browsers
- Verify Docker deployment

---

## Development Timeline

This is a complex project with multiple components. Here's a suggested development order:

1. **Week 1**: Foundation (Phases 1-2)
   - Project structure
   - Configuration system
   - Storage layer with SQLite

2. **Week 2**: Email Processing (Phase 3-4)
   - Email parser
   - SMTP server
   - Basic email capture

3. **Week 3**: API Layer (Phase 5-6)
   - REST API endpoints
   - WebSocket support
   - Integration with storage

4. **Week 4**: Frontend (Phase 7)
   - HTML/CSS layout
   - JavaScript components
   - Real-time updates

5. **Week 5**: Polish & Deploy (Phase 8-9)
   - Integration testing
   - Docker support
   - Documentation

---

## Key Technical Decisions

1. **SMTP Server**: Use `emersion/go-smtp` instead of embedding Maddy
   - Simpler integration
   - More control over behavior
   - Easier to maintain

2. **Storage**: SQLite with FTS5 for search
   - Zero configuration
   - Good performance for development tool
   - Full-text search built-in

3. **Frontend**: Vanilla JavaScript
   - No build step required
   - Lightweight
   - Easy to understand and modify

4. **Real-time**: WebSocket for updates
   - Better than polling
   - Low latency
   - Standard protocol

5. **Deployment**: Single binary + Docker
   - Easy distribution
   - No external dependencies
   - Works everywhere

---

## Success Criteria

- ✅ Captures all SMTP emails without configuration
- ✅ Web interface loads in < 1 second
- ✅ API responds in < 100ms
- ✅ Real-time updates appear instantly
- ✅ Handles 1000+ emails without slowdown
- ✅ Single binary deployment
- ✅ Works on Linux, macOS, Windows
- ✅ Docker image < 50MB
- ✅ Memory usage < 100MB (idle)
- ✅ Complete documentation
