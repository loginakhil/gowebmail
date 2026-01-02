package storage

import (
	"errors"
	"time"
)

var (
	// ErrNotFound is returned when an email is not found
	ErrNotFound = errors.New("email not found")
	// ErrInvalidID is returned when an invalid ID is provided
	ErrInvalidID = errors.New("invalid email ID")
)

// Email represents an email message
type Email struct {
	ID          int64               `json:"id"`
	MessageID   string              `json:"messageId"`
	From        string              `json:"from"`
	To          []string            `json:"to"`
	CC          []string            `json:"cc,omitempty"`
	BCC         []string            `json:"bcc,omitempty"`
	Subject     string              `json:"subject"`
	BodyPlain   string              `json:"bodyPlain"`
	BodyHTML    string              `json:"bodyHTML"`
	Headers     map[string][]string `json:"headers"`
	Attachments []AttachmentMeta    `json:"attachments,omitempty"`
	Size        int64               `json:"size"`
	ReceivedAt  time.Time           `json:"receivedAt"`
	Read        bool                `json:"read"`
}

// AttachmentMeta represents attachment metadata
type AttachmentMeta struct {
	ID          int64  `json:"id"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int64  `json:"size"`
}

// Attachment represents a full attachment with data
type Attachment struct {
	AttachmentMeta
	Data []byte `json:"-"`
}

// EmailFilter represents filter criteria for listing emails
type EmailFilter struct {
	From    string
	To      string
	Subject string
	Since   *time.Time
	Until   *time.Time
}

// EmailListResult represents a paginated list of emails
type EmailListResult struct {
	Emails []*Email `json:"emails"`
	Total  int64    `json:"total"`
}
