package storage

import "time"

// Storage defines the interface for email storage operations
type Storage interface {
	// Email operations
	SaveEmail(email *Email) (int64, error)
	GetEmail(id int64) (*Email, error)
	ListEmails(filter *EmailFilter, limit, offset int) (*EmailListResult, error)
	SearchEmails(query string, limit, offset int) (*EmailListResult, error)
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
