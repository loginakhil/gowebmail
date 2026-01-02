# GoWebMail - API Reference

## Base URL

```
http://localhost:8080/api
```

## Response Format

All API responses follow this structure:

### Success Response
```json
{
  "success": true,
  "data": {
    // Response data here
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message"
  }
}
```

## Error Codes

| Code | Description |
|------|-------------|
| `NOT_FOUND` | Resource not found |
| `STORAGE_ERROR` | Database operation failed |
| `INVALID_REQUEST` | Invalid request parameters |
| `INTERNAL_ERROR` | Internal server error |

---

## Endpoints

### 1. List Emails

Get a paginated list of emails with optional filtering.

**Endpoint**: `GET /api/emails`

**Query Parameters**:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `limit` | integer | 50 | Number of results (max: 100) |
| `offset` | integer | 0 | Pagination offset |
| `from` | string | - | Filter by sender email |
| `to` | string | - | Filter by recipient email |
| `subject` | string | - | Filter by subject (partial match) |
| `since` | string | - | Filter by date (ISO 8601 format) |
| `until` | string | - | Filter by date (ISO 8601 format) |

**Example Request**:
```bash
curl "http://localhost:8080/api/emails?limit=10&from=test@example.com"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "emails": [
      {
        "id": 1,
        "messageId": "<abc123@example.com>",
        "from": "sender@example.com",
        "to": ["recipient@example.com"],
        "cc": [],
        "bcc": [],
        "subject": "Test Email",
        "bodyPlain": "This is a test email",
        "bodyHTML": "<p>This is a test email</p>",
        "headers": {
          "Content-Type": ["text/plain; charset=utf-8"],
          "Date": ["Thu, 02 Jan 2026 15:30:00 +0000"]
        },
        "attachments": [],
        "size": 1024,
        "receivedAt": "2026-01-02T15:30:00Z",
        "read": false
      }
    ],
    "total": 42,
    "limit": 10,
    "offset": 0
  }
}
```

---

### 2. Get Email

Get details of a specific email by ID.

**Endpoint**: `GET /api/emails/{id}`

**Path Parameters**:
- `id` (integer): Email ID

**Example Request**:
```bash
curl "http://localhost:8080/api/emails/1"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "id": 1,
    "messageId": "<abc123@example.com>",
    "from": "sender@example.com",
    "to": ["recipient@example.com"],
    "subject": "Test Email",
    "bodyPlain": "This is a test email",
    "bodyHTML": "<p>This is a test email</p>",
    "headers": {
      "Content-Type": ["text/plain; charset=utf-8"]
    },
    "attachments": [
      {
        "id": 1,
        "filename": "document.pdf",
        "contentType": "application/pdf",
        "size": 51200
      }
    ],
    "size": 52224,
    "receivedAt": "2026-01-02T15:30:00Z",
    "read": false
  }
}
```

---

### 3. Delete Email

Delete a specific email by ID.

**Endpoint**: `DELETE /api/emails/{id}`

**Path Parameters**:
- `id` (integer): Email ID

**Example Request**:
```bash
curl -X DELETE "http://localhost:8080/api/emails/1"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "deleted": 1
  }
}
```

---

### 4. Delete All Emails

Delete all emails from the database.

**Endpoint**: `DELETE /api/emails`

**Example Request**:
```bash
curl -X DELETE "http://localhost:8080/api/emails"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "deleted": 42
  }
}
```

---

### 5. Search Emails

Search emails using full-text search.

**Endpoint**: `GET /api/emails/search`

**Query Parameters**:

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `q` | string | - | Search query (required) |
| `limit` | integer | 50 | Number of results (max: 100) |
| `offset` | integer | 0 | Pagination offset |

**Example Request**:
```bash
curl "http://localhost:8080/api/emails/search?q=invoice&limit=10"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "emails": [
      {
        "id": 5,
        "from": "billing@example.com",
        "subject": "Your Invoice #12345",
        "receivedAt": "2026-01-02T14:00:00Z"
      }
    ],
    "total": 3,
    "limit": 10,
    "offset": 0
  }
}
```

---

### 6. Get Raw Email

Get the raw email source (RFC 822 format).

**Endpoint**: `GET /api/emails/{id}/raw`

**Path Parameters**:
- `id` (integer): Email ID

**Example Request**:
```bash
curl "http://localhost:8080/api/emails/1/raw"
```

**Example Response**:
```
From: sender@example.com
To: recipient@example.com
Subject: Test Email
Date: Thu, 02 Jan 2026 15:30:00 +0000
Content-Type: text/plain; charset=utf-8

This is a test email
```

---

### 7. Get HTML Email Body

Get the sanitized HTML body of an email.

**Endpoint**: `GET /api/emails/{id}/html`

**Path Parameters**:
- `id` (integer): Email ID

**Example Request**:
```bash
curl "http://localhost:8080/api/emails/1/html"
```

**Example Response**:
```html
<p>This is a test email</p>
<p>With <strong>HTML</strong> formatting</p>
```

---

### 8. Download Attachment

Download an email attachment.

**Endpoint**: `GET /api/emails/{id}/attachments/{aid}`

**Path Parameters**:
- `id` (integer): Email ID
- `aid` (integer): Attachment ID

**Example Request**:
```bash
curl "http://localhost:8080/api/emails/1/attachments/1" -o document.pdf
```

**Response**: Binary file with appropriate Content-Type and Content-Disposition headers

---

### 9. Get Statistics

Get email statistics.

**Endpoint**: `GET /api/stats`

**Example Request**:
```bash
curl "http://localhost:8080/api/stats"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "totalEmails": 42,
    "totalSize": 5242880,
    "todayCount": 12,
    "oldestEmail": "2026-01-01T10:00:00Z",
    "newestEmail": "2026-01-02T15:30:00Z"
  }
}
```

---

### 10. Health Check

Check if the API is running.

**Endpoint**: `GET /api/health`

**Example Request**:
```bash
curl "http://localhost:8080/api/health"
```

**Example Response**:
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "1.0.0",
    "uptime": 3600
  }
}
```

---

## WebSocket API

### Connection

**Endpoint**: `ws://localhost:8080/ws`

**Example (JavaScript)**:
```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected to GoWebMail');
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

### Message Types

#### 1. New Email

Sent when a new email is received.

```json
{
  "type": "email.new",
  "data": {
    "id": 1,
    "from": "sender@example.com",
    "to": ["recipient@example.com"],
    "subject": "Test Email",
    "receivedAt": "2026-01-02T15:30:00Z"
  }
}
```

#### 2. Email Deleted

Sent when an email is deleted.

```json
{
  "type": "email.deleted",
  "data": {
    "id": 1
  }
}
```

#### 3. All Emails Cleared

Sent when all emails are deleted.

```json
{
  "type": "emails.cleared",
  "data": {}
}
```

---

## Usage Examples

### Example 1: Send and Retrieve Email

**Step 1**: Send email via SMTP
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

**Step 2**: Retrieve email via API
```bash
curl "http://localhost:8080/api/emails?limit=1"
```

---

### Example 2: Search for Emails

```bash
# Search for emails containing "invoice"
curl "http://localhost:8080/api/emails/search?q=invoice"

# Search with pagination
curl "http://localhost:8080/api/emails/search?q=invoice&limit=10&offset=0"
```

---

### Example 3: Filter Emails

```bash
# Get emails from specific sender
curl "http://localhost:8080/api/emails?from=billing@example.com"

# Get emails from today
TODAY=$(date -u +"%Y-%m-%dT00:00:00Z")
curl "http://localhost:8080/api/emails?since=$TODAY"

# Combine filters
curl "http://localhost:8080/api/emails?from=billing@example.com&subject=invoice"
```

---

### Example 4: Delete Emails

```bash
# Delete specific email
curl -X DELETE "http://localhost:8080/api/emails/1"

# Delete all emails
curl -X DELETE "http://localhost:8080/api/emails"
```

---

### Example 5: Real-time Updates (JavaScript)

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

// Handle new emails
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  if (message.type === 'email.new') {
    console.log('New email received:', message.data);
    // Update UI with new email
    addEmailToList(message.data);
  }
  
  if (message.type === 'email.deleted') {
    console.log('Email deleted:', message.data.id);
    // Remove email from UI
    removeEmailFromList(message.data.id);
  }
};
```

---

### Example 6: Integration with Testing Framework

**Python (pytest)**:
```python
import smtplib
import requests
from email.mime.text import MIMEText

def test_email_delivery():
    # Send email
    msg = MIMEText('Test content')
    msg['Subject'] = 'Test Subject'
    msg['From'] = 'test@example.com'
    msg['To'] = 'recipient@example.com'
    
    with smtplib.SMTP('localhost', 1025) as server:
        server.send_message(msg)
    
    # Verify email was received
    response = requests.get('http://localhost:8080/api/emails?limit=1')
    data = response.json()
    
    assert data['success'] == True
    assert len(data['data']['emails']) > 0
    
    email = data['data']['emails'][0]
    assert email['subject'] == 'Test Subject'
    assert email['from'] == 'test@example.com'
```

**JavaScript (Jest)**:
```javascript
const nodemailer = require('nodemailer');
const axios = require('axios');

test('email delivery', async () => {
  // Send email
  const transporter = nodemailer.createTransport({
    host: 'localhost',
    port: 1025,
    secure: false
  });
  
  await transporter.sendMail({
    from: 'test@example.com',
    to: 'recipient@example.com',
    subject: 'Test Subject',
    text: 'Test content'
  });
  
  // Verify email was received
  const response = await axios.get('http://localhost:8080/api/emails?limit=1');
  
  expect(response.data.success).toBe(true);
  expect(response.data.data.emails).toHaveLength(1);
  
  const email = response.data.data.emails[0];
  expect(email.subject).toBe('Test Subject');
  expect(email.from).toBe('test@example.com');
});
```

---

### Example 7: CI/CD Integration

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

**Docker Compose (for local testing)**:
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

---

## Rate Limiting

Currently, there is no rate limiting implemented as this is a development tool. If deploying in a shared environment, consider adding rate limiting middleware.

---

## Authentication

By default, the API does not require authentication. To enable basic authentication:

**Configuration**:
```yaml
web:
  auth:
    enabled: true
    username: "admin"
    password: "your-secure-password"
```

**Usage**:
```bash
curl -u admin:your-secure-password "http://localhost:8080/api/emails"
```

---

## CORS

CORS is enabled by default to allow access from any origin. This is suitable for development but should be restricted in production environments.

---

## Content Security Policy

When viewing HTML emails, the following CSP is applied:
- No JavaScript execution
- No external resource loading
- Sandboxed iframe rendering

---

## Best Practices

1. **Clean up regularly**: Use the delete all endpoint or configure retention policies
2. **Use search**: For large email volumes, use the search endpoint instead of filtering
3. **WebSocket for real-time**: Use WebSocket connections for real-time updates instead of polling
4. **Pagination**: Always use pagination for listing emails
5. **Error handling**: Always check the `success` field in responses

---

## Troubleshooting

### Email not appearing in API

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

---

## Performance Tips

1. **Limit results**: Use appropriate `limit` values (default: 50, max: 100)
2. **Use filters**: Filter at the database level instead of client-side
3. **Index usage**: Filters on `from`, `to`, `subject`, and `receivedAt` use indexes
4. **Attachment handling**: Large attachments are streamed, not loaded into memory
5. **Connection pooling**: Database connections are pooled for efficiency

---

## Security Considerations

⚠️ **Important**: GoWebMail is designed for development and testing environments only.

- No encryption by default
- No authentication by default (optional)
- Accepts all emails without validation
- HTML sanitization prevents XSS but emails should not be trusted
- Do not expose to public internet
- Do not use for production email

---

## API Client Libraries

### JavaScript/TypeScript

```javascript
class GoWebMailClient {
  constructor(baseURL = 'http://localhost:8080/api') {
    this.baseURL = baseURL;
  }
  
  async listEmails(params = {}) {
    const query = new URLSearchParams(params);
    const response = await fetch(`${this.baseURL}/emails?${query}`);
    return response.json();
  }
  
  async getEmail(id) {
    const response = await fetch(`${this.baseURL}/emails/${id}`);
    return response.json();
  }
  
  async deleteEmail(id) {
    const response = await fetch(`${this.baseURL}/emails/${id}`, {
      method: 'DELETE'
    });
    return response.json();
  }
  
  async searchEmails(query) {
    const response = await fetch(
      `${this.baseURL}/emails/search?q=${encodeURIComponent(query)}`
    );
    return response.json();
  }
}
```

### Python

```python
import requests

class GoWebMailClient:
    def __init__(self, base_url='http://localhost:8080/api'):
        self.base_url = base_url
    
    def list_emails(self, **params):
        response = requests.get(f'{self.base_url}/emails', params=params)
        return response.json()
    
    def get_email(self, email_id):
        response = requests.get(f'{self.base_url}/emails/{email_id}')
        return response.json()
    
    def delete_email(self, email_id):
        response = requests.delete(f'{self.base_url}/emails/{email_id}')
        return response.json()
    
    def search_emails(self, query):
        response = requests.get(
            f'{self.base_url}/emails/search',
            params={'q': query}
        )
        return response.json()
```

### Go

```go
package gowebmail

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Client struct {
    BaseURL string
    HTTP    *http.Client
}

func NewClient(baseURL string) *Client {
    return &Client{
        BaseURL: baseURL,
        HTTP:    &http.Client{},
    }
}

func (c *Client) ListEmails(params map[string]string) (*EmailListResponse, error) {
    req, _ := http.NewRequest("GET", c.BaseURL+"/emails", nil)
    q := req.URL.Query()
    for k, v := range params {
        q.Add(k, v)
    }
    req.URL.RawQuery = q.Encode()
    
    resp, err := c.HTTP.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var result EmailListResponse
    json.NewDecoder(resp.Body).Decode(&result)
    return &result, nil
}
```

---

## Changelog

### Version 1.0.0
- Initial release
- SMTP server with email capture
- REST API with full CRUD operations
- WebSocket support for real-time updates
- Web interface
- SQLite storage with FTS5 search
- Docker support
