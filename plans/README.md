# GoWebMail - Email Testing Tool

A complete email testing tool similar to MailSlurper, built with Go and designed for development environments. Captures all outgoing SMTP emails and provides a modern web interface for viewing, searching, and managing captured messages.

## 🎯 Project Overview

GoWebMail is a lightweight, zero-configuration email testing tool that helps developers test email functionality without sending real emails. It acts as an SMTP server that captures all emails and provides a web interface to view them.

### Key Features

- ✅ **SMTP Server**: Accepts all incoming mail without authentication
- ✅ **Web Interface**: Modern, responsive UI for viewing emails
- ✅ **REST API**: Complete API for programmatic access
- ✅ **Real-time Updates**: WebSocket support for instant notifications
- ✅ **Full-text Search**: Fast search across all email content
- ✅ **Attachment Support**: View and download email attachments
- ✅ **HTML Email Rendering**: Safe HTML email preview with sanitization
- ✅ **Docker Support**: Easy deployment with Docker and docker-compose
- ✅ **Single Binary**: No external dependencies required
- ✅ **Cross-platform**: Works on Linux, macOS, and Windows

## 📋 Planning Documents

This repository contains comprehensive planning documentation for the GoWebMail project:

### 1. [Architecture Document](architecture.md)
Complete system architecture including:
- System component diagram
- Database schema design
- API endpoint specifications
- WebSocket message formats
- Frontend component structure
- Security considerations
- Performance optimization strategies

### 2. [Implementation Guide](implementation-guide.md)
Detailed phase-by-phase implementation plan:
- Phase 1: Project Foundation (Configuration, Structure)
- Phase 2: Storage Layer (SQLite, Models)
- Phase 3: Email Parsing (MIME, Sanitization)
- Phase 4: SMTP Server (Email Capture)
- Phase 5: REST API (HTTP Endpoints)
- Phase 6: WebSocket Support (Real-time Updates)
- Phase 7: Frontend Development (UI/UX)
- Phase 8: Integration & Polish
- Phase 9: Docker & Documentation

Each phase includes:
- Detailed technical specifications
- Code examples and patterns
- File structure and organization
- Testing strategies
- Success criteria

### 3. [API Reference](api-reference.md)
Complete API documentation including:
- All REST endpoints with examples
- WebSocket message formats
- Request/response formats
- Error codes and handling
- Usage examples in multiple languages
- Integration examples for testing frameworks
- CI/CD integration guides

## 🏗️ Architecture Overview

```
┌─────────────────┐
│   Application   │
│  Under Test     │
└────────┬────────┘
         │ SMTP
         ▼
┌─────────────────────────────────────────┐
│           GoWebMail System              │
│                                         │
│  ┌──────────┐    ┌──────────────┐     │
│  │   SMTP   │───▶│   Storage    │     │
│  │  Server  │    │   (SQLite)   │     │
│  └──────────┘    └──────┬───────┘     │
│                         │              │
│  ┌──────────┐    ┌──────▼───────┐     │
│  │   REST   │───▶│   Storage    │     │
│  │   API    │    │   Service    │     │
│  └────┬─────┘    └──────────────┘     │
│       │                                │
│  ┌────▼─────┐                         │
│  │WebSocket │                         │
│  │   Hub    │                         │
│  └────┬─────┘                         │
│       │                                │
└───────┼────────────────────────────────┘
        │ HTTP/WS
        ▼
┌─────────────────┐
│  Web Interface  │
│  (Browser)      │
└─────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- Go 1.25 or later
- SQLite 3
- Modern web browser

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/gowebmail.git
cd gowebmail

# Build the application
go build -o gowebmail ./cmd/gowebmail

# Run the application
./gowebmail
```

### Using Docker

```bash
# Build Docker image
docker build -t gowebmail -f docker/Dockerfile .

# Run with docker-compose
docker-compose -f docker/docker-compose.yml up
```

### Configuration

Create a `gowebmail.yml` configuration file:

```yaml
smtp:
  host: "0.0.0.0"
  port: 1025

http:
  host: "0.0.0.0"
  port: 8080

storage:
  type: "sqlite"
  path: "./data/gowebmail.db"

retention:
  enabled: true
  max_age: "7d"
  max_count: 1000
```

## 📊 Project Structure

```
gowebmail/
├── cmd/
│   └── gowebmail/          # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── smtp/               # SMTP server implementation
│   ├── storage/            # Database layer
│   ├── api/                # REST API and WebSocket
│   ├── email/              # Email parsing and sanitization
│   └── retention/          # Retention policy enforcement
├── web/
│   ├── css/                # Stylesheets
│   ├── js/                 # JavaScript modules
│   └── index.html          # Main UI
├── docs/                   # Additional documentation
├── docker/                 # Docker configuration
├── configs/                # Example configurations
├── plans/                  # Planning documents (this directory)
│   ├── architecture.md
│   ├── implementation-guide.md
│   └── api-reference.md
└── data/                   # SQLite database (created at runtime)
```

## 🔧 Technology Stack

### Backend
- **Language**: Go 1.25
- **SMTP**: `github.com/emersion/go-smtp`
- **Database**: SQLite with FTS5
- **HTTP Router**: `github.com/gorilla/mux`
- **WebSocket**: `github.com/gorilla/websocket`
- **Email Parsing**: `github.com/emersion/go-message`
- **HTML Sanitization**: `github.com/microcosm-cc/bluemonday`
- **Logging**: `github.com/rs/zerolog`

### Frontend
- **Framework**: Vanilla JavaScript (ES6+)
- **Styling**: Modern CSS (Grid, Flexbox)
- **Real-time**: WebSocket API
- **No build step required**

### Infrastructure
- **Container**: Docker
- **Orchestration**: docker-compose
- **Database**: SQLite (embedded)

## 📝 Development Phases

The project is organized into 9 development phases:

1. **Foundation** - Project structure, configuration system
2. **Storage** - Database schema, SQLite implementation
3. **Email Parsing** - MIME parsing, HTML sanitization
4. **SMTP Server** - Email capture and storage
5. **REST API** - HTTP endpoints and middleware
6. **WebSocket** - Real-time update system
7. **Frontend** - Web interface development
8. **Integration** - System integration and polish
9. **Deployment** - Docker support and documentation

See [Implementation Guide](implementation-guide.md) for detailed phase breakdowns.

## 🎨 User Interface

The web interface provides:

- **Email List**: Inbox-style view with sorting and filtering
- **Preview Pane**: Split view for reading emails
- **Search**: Full-text search across all emails
- **Filters**: Filter by sender, recipient, date, subject
- **HTML Rendering**: Safe HTML email preview
- **Attachments**: View and download attachments
- **Real-time Updates**: Instant notification of new emails
- **Responsive Design**: Works on desktop and mobile

## 🔌 API Integration

### Send Email (SMTP)

```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText('Test email body')
msg['Subject'] = 'Test Email'
msg['From'] = 'sender@example.com'
msg['To'] = 'recipient@example.com'

with smtplib.SMTP('localhost', 1025) as server:
    server.send_message(msg)
```

### Retrieve Email (REST API)

```bash
curl http://localhost:8080/api/emails
```

### Real-time Updates (WebSocket)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'email.new') {
    console.log('New email:', message.data);
  }
};
```

## 🧪 Testing Integration

### Python (pytest)

```python
def test_email_delivery():
    # Send email via SMTP
    send_email('test@example.com', 'Test Subject', 'Test Body')
    
    # Verify via API
    response = requests.get('http://localhost:8080/api/emails?limit=1')
    assert response.json()['data']['emails'][0]['subject'] == 'Test Subject'
```

### JavaScript (Jest)

```javascript
test('email delivery', async () => {
  await sendEmail('test@example.com', 'Test Subject', 'Test Body');
  
  const response = await fetch('http://localhost:8080/api/emails?limit=1');
  const data = await response.json();
  
  expect(data.data.emails[0].subject).toBe('Test Subject');
});
```

## 🐳 Docker Deployment

### Docker Compose

```yaml
version: '3.8'

services:
  gowebmail:
    image: gowebmail:latest
    ports:
      - "1025:1025"  # SMTP
      - "8080:8080"  # HTTP
    volumes:
      - ./data:/app/data
    environment:
      - GOWEBMAIL_LOG_LEVEL=info
```

### CI/CD Integration

**GitLab CI**:
```yaml
test:
  services:
    - name: gowebmail:latest
      alias: mailserver
  variables:
    SMTP_HOST: mailserver
    SMTP_PORT: 1025
  script:
    - npm test
```

**GitHub Actions**:
```yaml
services:
  mailserver:
    image: gowebmail:latest
    ports:
      - 1025:1025
      - 8080:8080
```

## 🔒 Security Considerations

⚠️ **Important**: GoWebMail is designed for development and testing environments only.

- Not suitable for production use
- No encryption by default
- Optional basic authentication
- Accepts all emails without validation
- Should not be exposed to public internet
- HTML emails are sanitized but should not be trusted

## 📈 Performance Targets

- **Memory Usage**: < 100MB (idle)
- **API Response Time**: < 100ms
- **Email Capacity**: 1000+ emails without degradation
- **Real-time Latency**: < 100ms for WebSocket updates
- **Binary Size**: < 20MB
- **Docker Image**: < 50MB

## 🎯 Success Criteria

- ✅ Zero-configuration startup
- ✅ Single binary deployment
- ✅ Cross-platform compatibility
- ✅ Real-time email updates
- ✅ Full-text search functionality
- ✅ HTML email rendering
- ✅ Attachment support
- ✅ REST API with complete CRUD
- ✅ WebSocket real-time updates
- ✅ Docker support
- ✅ Comprehensive documentation

## 📚 Documentation

- **[Architecture](architecture.md)**: System design and component interactions
- **[Implementation Guide](implementation-guide.md)**: Phase-by-phase development plan
- **[API Reference](api-reference.md)**: Complete API documentation with examples

## 🤝 Contributing

This is a planning repository. Implementation contributions should follow the architecture and implementation guide provided.

## 📄 License

GPL-3.0 License (to maintain compatibility with Maddy if used)

## 🙏 Acknowledgments

- Inspired by [MailSlurper](https://github.com/mailslurper/mailslurper)
- SMTP implementation based on [go-smtp](https://github.com/emersion/go-smtp)
- Email parsing using [go-message](https://github.com/emersion/go-message)

## 📞 Support

For questions or issues:
1. Check the [Implementation Guide](implementation-guide.md)
2. Review the [API Reference](api-reference.md)
3. Consult the [Architecture Document](architecture.md)

## 🗺️ Roadmap

### Phase 1 (Foundation)
- [x] Architecture design
- [x] Implementation planning
- [x] API specification
- [ ] Project structure setup
- [ ] Configuration system

### Phase 2 (Core Features)
- [ ] Storage layer implementation
- [ ] SMTP server
- [ ] REST API
- [ ] WebSocket support

### Phase 3 (User Interface)
- [ ] Frontend development
- [ ] Email list view
- [ ] Email preview
- [ ] Search functionality

### Phase 4 (Polish)
- [ ] Docker support
- [ ] Documentation
- [ ] Testing
- [ ] Performance optimization

### Future Enhancements
- [ ] Multiple storage backends (PostgreSQL, MySQL)
- [ ] Email forwarding/relay
- [ ] Export functionality (mbox, EML)
- [ ] Advanced filtering (regex, boolean)
- [ ] Email templates for testing
- [ ] API client libraries
- [ ] Prometheus metrics
- [ ] Multi-user support
- [ ] Dark mode UI

## 📊 Project Status

**Current Phase**: Planning & Architecture ✅

**Next Steps**:
1. Review and approve architecture
2. Set up project structure
3. Begin Phase 1 implementation
4. Implement storage layer
5. Build SMTP server

---

**Ready to implement?** Switch to Code mode to begin development following the implementation guide.
