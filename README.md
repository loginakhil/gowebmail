# GoWebMail - Email Testing Tool

A complete email testing tool for development environments. Captures all outgoing SMTP emails and provides a modern web interface for viewing, searching, and managing captured messages.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)

## Features

- ✅ **SMTP Server**: Accepts all incoming mail without authentication on port 1025
- ✅ **Web Interface**: Modern, responsive UI for viewing emails
- ✅ **REST API**: Complete API for programmatic access
- ✅ **Real-time Updates**: WebSocket support for instant notifications
- ✅ **Full-text Search**: Fast search across all email content using SQLite FTS5
- ✅ **Attachment Support**: View and download email attachments
- ✅ **HTML Email Rendering**: Safe HTML email preview with sanitization
- ✅ **Docker Support**: Easy deployment with Docker and docker-compose
- ✅ **Single Binary**: No external dependencies required
- ✅ **Cross-platform**: Works on Linux, macOS, and Windows

## Quick Start

### Using Pre-built Binary

1. Download the latest release
2. Run the application:
```bash
./gowebmail
```

3. Open your browser to `http://localhost:8080`
4. Configure your application to send emails to `localhost:1025`

### Using Docker

```bash
docker-compose -f docker/docker-compose.yml up
```

### Building from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/gowebmail.git
cd gowebmail

# Build
go build -o gowebmail ./cmd/gowebmail

# Run
./gowebmail
```

## Configuration

Create a `gowebmail.yml` file (see [`configs/gowebmail.example.yml`](configs/gowebmail.example.yml)):

```yaml
smtp:
  host: "0.0.0.0"
  port: 1025
  max_message_size: 10485760  # 10MB

http:
  host: "0.0.0.0"
  port: 8080

storage:
  type: "sqlite"
  path: "./data/gowebmail.db"

retention:
  enabled: true
  max_age: "168h"        # 7 days
  max_count: 1000
  cleanup_interval: "1h"

web:
  enabled: true
  auth:
    enabled: false
    username: "admin"
    password: "changeme"

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Environment Variables

Override configuration with environment variables:

- `GOWEBMAIL_SMTP_PORT` - SMTP server port
- `GOWEBMAIL_HTTP_PORT` - HTTP server port
- `GOWEBMAIL_STORAGE_PATH` - Database file path
- `GOWEBMAIL_LOG_LEVEL` - Log level (debug, info, warn, error)
- `GOWEBMAIL_WEB_AUTH_ENABLED` - Enable web authentication
- `GOWEBMAIL_WEB_AUTH_USERNAME` - Web interface username
- `GOWEBMAIL_WEB_AUTH_PASSWORD` - Web interface password

## Usage

### Sending Test Emails

**Python:**
```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText('This is a test email')
msg['Subject'] = 'Test Email'
msg['From'] = 'sender@example.com'
msg['To'] = 'recipient@example.com'

with smtplib.SMTP('localhost', 1025) as server:
    server.send_message(msg)
```

**Node.js:**
```javascript
const nodemailer = require('nodemailer');

const transporter = nodemailer.createTransport({
    host: 'localhost',
    port: 1025,
    secure: false
});

await transporter.sendMail({
    from: 'sender@example.com',
    to: 'recipient@example.com',
    subject: 'Test Email',
    text: 'This is a test email'
});
```

**Go:**
```go
package main

import (
    "net/smtp"
)

func main() {
    from := "sender@example.com"
    to := []string{"recipient@example.com"}
    msg := []byte("Subject: Test Email\r\n\r\nThis is a test email")
    
    smtp.SendMail("localhost:1025", nil, from, to, msg)
}
```

### API Usage

**List Emails:**
```bash
curl http://localhost:8080/api/emails
```

**Search Emails:**
```bash
curl "http://localhost:8080/api/emails/search?q=invoice"
```

**Delete Email:**
```bash
curl -X DELETE http://localhost:8080/api/emails/1
```

**Delete All Emails:**
```bash
curl -X DELETE http://localhost:8080/api/emails
```

See [API Reference](plans/api-reference.md) for complete documentation.

## CI/CD Integration

### GitLab CI

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

### GitHub Actions

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mailserver:
        image: gowebmail:latest
        ports:
          - 1025:1025
          - 8080:8080
    
    steps:
      - uses: actions/checkout@v2
      - name: Run tests
        env:
          SMTP_HOST: localhost
          SMTP_PORT: 1025
        run: npm test
```

### Docker Compose (Local Testing)

```yaml
version: '3.8'

services:
  app:
    build: .
    environment:
      - SMTP_HOST=mailserver
      - SMTP_PORT=1025
    depends_on:
      - mailserver
  
  mailserver:
    image: gowebmail:latest
    ports:
      - "1025:1025"
      - "8080:8080"
```

## Architecture

GoWebMail consists of several key components:

- **SMTP Server**: Built using `emersion/go-smtp`, accepts all emails without authentication
- **Storage Layer**: SQLite database with FTS5 for full-text search
- **REST API**: HTTP endpoints for email management
- **WebSocket**: Real-time updates for new emails
- **Web UI**: Vanilla JavaScript frontend with no framework dependencies

See [Architecture Documentation](plans/architecture.md) for detailed design.

## Development

### Prerequisites

- Go 1.25 or later
- SQLite 3
- Modern web browser

### Project Structure

```
gowebmail/
├── cmd/gowebmail/          # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── smtp/               # SMTP server
│   ├── storage/            # Database layer
│   ├── api/                # REST API and WebSocket
│   ├── email/              # Email parsing and sanitization
│   └── retention/          # Retention policy
├── web/                    # Frontend files
├── docker/                 # Docker configuration
├── configs/                # Example configurations
├── plans/                  # Planning and documentation
└── docs/                   # Additional documentation
```

### Building

```bash
# Build for current platform
go build -o gowebmail ./cmd/gowebmail

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o gowebmail-linux ./cmd/gowebmail

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o gowebmail.exe ./cmd/gowebmail

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o gowebmail-mac ./cmd/gowebmail
```

### Running Tests

```bash
go test ./...
```

## Security Considerations

⚠️ **Important**: GoWebMail is designed for development and testing environments only.

- Not suitable for production use
- No encryption by default
- Optional basic authentication for web interface
- Accepts all emails without validation
- Should not be exposed to public internet
- HTML emails are sanitized but should not be trusted

## Performance

- **Memory Usage**: < 100MB (idle)
- **API Response Time**: < 100ms
- **Email Capacity**: 1000+ emails without degradation
- **Real-time Latency**: < 100ms for WebSocket updates
- **Binary Size**: < 20MB
- **Docker Image**: < 50MB

## Troubleshooting

### Emails not appearing

1. Check SMTP server is running: `curl http://localhost:8080/api/health`
2. Verify SMTP connection: `telnet localhost 1025`
3. Check logs for errors
4. Verify email was sent successfully

### WebSocket not connecting

1. Check browser console for errors
2. Verify WebSocket URL (ws:// vs wss://)
3. Check firewall settings
4. Verify server is running

### Search not working

1. Ensure FTS5 is enabled in SQLite
2. Check search query syntax
3. Verify emails exist in database
4. Check logs for errors

## Contributing

Contributions are welcome! Please see the planning documents in the [`plans/`](plans/) directory for architecture and implementation details.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Inspired by [MailSlurper](https://github.com/mailslurper/mailslurper)
- SMTP implementation using [go-smtp](https://github.com/emersion/go-smtp)
- Email parsing using [go-message](https://github.com/emersion/go-message)
- HTML sanitization using [bluemonday](https://github.com/microcosm-cc/bluemonday)

## Support

For questions or issues:
1. Check the [Implementation Guide](plans/implementation-guide.md)
2. Review the [API Reference](plans/api-reference.md)
3. Consult the [Architecture Document](plans/architecture.md)
4. Open an issue on GitHub

## Roadmap

- [ ] Multiple storage backends (PostgreSQL, MySQL)
- [ ] Email forwarding/relay capability
- [ ] Export functionality (mbox, EML format)
- [ ] Advanced filtering (regex, boolean operators)
- [ ] Email templates for testing
- [ ] API client libraries (Go, Python, JavaScript)
- [ ] Prometheus metrics
- [ ] Multi-user support
- [ ] Dark mode UI

---

**Made with ❤️ for developers who need to test email functionality**
# gowebmail
