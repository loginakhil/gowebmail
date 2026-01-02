package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"strings"

	"github.com/emersion/go-message"
	"gowebmail/internal/storage"
)

// Parser handles email parsing
type Parser struct{}

// NewParser creates a new email parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses an email from a reader
func (p *Parser) Parse(r io.Reader) (*storage.Email, error) {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read email: %w", err)
	}

	// Parse message
	msg, err := mail.ReadMessage(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse email: %w", err)
	}

	email := &storage.Email{
		Headers: make(map[string][]string),
	}

	// Parse headers
	p.parseHeaders(msg.Header, email)

	// Parse body
	entity, err := message.Read(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse MIME: %w", err)
	}

	attachments, err := p.parseBody(entity, email)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}

	// Convert attachments to metadata
	for _, att := range attachments {
		email.Attachments = append(email.Attachments, storage.AttachmentMeta{
			Filename:    att.Filename,
			ContentType: att.ContentType,
			Size:        att.Size,
		})
	}

	// Calculate size
	email.Size = int64(len(data))

	return email, nil
}

// parseHeaders extracts headers from the email
func (p *Parser) parseHeaders(header mail.Header, email *storage.Email) {
	// Copy all headers
	for key, values := range header {
		email.Headers[key] = values
	}

	// Extract common headers
	email.MessageID = header.Get("Message-ID")
	email.Subject = p.decodeHeader(header.Get("Subject"))

	// From address
	if from := header.Get("From"); from != "" {
		if addr, err := mail.ParseAddress(from); err == nil {
			email.From = addr.Address
		} else {
			email.From = from
		}
	}

	// To addresses
	if to := header.Get("To"); to != "" {
		email.To = p.parseAddressList(to)
	}

	// CC addresses
	if cc := header.Get("Cc"); cc != "" {
		email.CC = p.parseAddressList(cc)
	}

	// BCC addresses
	if bcc := header.Get("Bcc"); bcc != "" {
		email.BCC = p.parseAddressList(bcc)
	}
}

// parseAddressList parses a comma-separated list of email addresses
func (p *Parser) parseAddressList(addrs string) []string {
	addresses, err := mail.ParseAddressList(addrs)
	if err != nil {
		// If parsing fails, split by comma
		parts := strings.Split(addrs, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}

	result := make([]string, len(addresses))
	for i, addr := range addresses {
		result[i] = addr.Address
	}
	return result
}

// decodeHeader decodes MIME encoded-word headers
func (p *Parser) decodeHeader(header string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(header)
	if err != nil {
		return header
	}
	return decoded
}

// parseBody parses the email body and extracts text and attachments
func (p *Parser) parseBody(entity *message.Entity, email *storage.Email) ([]*storage.Attachment, error) {
	var attachments []*storage.Attachment

	mediaType, _, err := entity.Header.ContentType()
	if err != nil {
		mediaType = "text/plain"
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		// Handle multipart
		mr := entity.MultipartReader()
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			atts, err := p.parsePart(part, email)
			if err != nil {
				return nil, err
			}
			attachments = append(attachments, atts...)
		}
	} else {
		// Handle single part
		atts, err := p.parsePart(entity, email)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, atts...)
	}

	return attachments, nil
}

// parsePart parses a single MIME part
func (p *Parser) parsePart(entity *message.Entity, email *storage.Email) ([]*storage.Attachment, error) {
	var attachments []*storage.Attachment

	mediaType, params, err := entity.Header.ContentType()
	if err != nil {
		mediaType = "text/plain"
		params = nil
	}

	// Check if it's an attachment
	disposition, dispParams, _ := entity.Header.ContentDisposition()
	isAttachment := disposition == "attachment" || (disposition == "inline" && dispParams["filename"] != "")

	if isAttachment {
		// Handle attachment
		filename := dispParams["filename"]
		if filename == "" {
			filename = params["name"]
		}
		if filename == "" {
			filename = "attachment"
		}

		data, err := io.ReadAll(entity.Body)
		if err != nil {
			return nil, err
		}

		// Decode if needed
		encoding := entity.Header.Get("Content-Transfer-Encoding")
		data = p.decodeContent(data, encoding)

		attachments = append(attachments, &storage.Attachment{
			AttachmentMeta: storage.AttachmentMeta{
				Filename:    filename,
				ContentType: mediaType,
				Size:        int64(len(data)),
			},
			Data: data,
		})
	} else if strings.HasPrefix(mediaType, "text/") {
		// Handle text content
		data, err := io.ReadAll(entity.Body)
		if err != nil {
			return nil, err
		}

		// Decode if needed
		encoding := entity.Header.Get("Content-Transfer-Encoding")
		data = p.decodeContent(data, encoding)

		text := string(data)

		if mediaType == "text/plain" {
			email.BodyPlain = text
		} else if mediaType == "text/html" {
			email.BodyHTML = text
		}
	} else if strings.HasPrefix(mediaType, "multipart/") {
		// Handle nested multipart
		mr := entity.MultipartReader()
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			atts, err := p.parsePart(part, email)
			if err != nil {
				return nil, err
			}
			attachments = append(attachments, atts...)
		}
	}

	return attachments, nil
}

// decodeContent decodes content based on transfer encoding
func (p *Parser) decodeContent(data []byte, encoding string) []byte {
	encoding = strings.ToLower(strings.TrimSpace(encoding))

	switch encoding {
	case "base64":
		decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
		n, err := base64.StdEncoding.Decode(decoded, data)
		if err == nil {
			return decoded[:n]
		}
	case "quoted-printable":
		reader := quotedprintable.NewReader(bytes.NewReader(data))
		decoded, err := io.ReadAll(reader)
		if err == nil {
			return decoded
		}
	}

	return data
}
