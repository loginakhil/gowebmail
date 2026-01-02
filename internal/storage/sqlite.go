package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

// SQLiteStorage implements the Storage interface using SQLite
type SQLiteStorage struct {
	db      *sql.DB
	logger  zerolog.Logger
	hasFTS5 bool
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string, logger zerolog.Logger) (*SQLiteStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1) // SQLite works best with single connection
	db.SetMaxIdleConns(1)

	storage := &SQLiteStorage{
		db:     db,
		logger: logger,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	logger.Info().Str("path", dbPath).Msg("SQLite storage initialized")

	return storage, nil
}

// initSchema initializes the database schema
func (s *SQLiteStorage) initSchema() error {
	// Create base schema
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}

	// Try to create FTS5 schema (optional)
	if _, err := s.db.Exec(fts5Schema); err != nil {
		s.logger.Warn().Err(err).Msg("FTS5 not available, full-text search will use LIKE-based fallback")
		s.hasFTS5 = false
	} else {
		s.logger.Info().Msg("FTS5 full-text search enabled")
		s.hasFTS5 = true
	}

	return nil
}

// SaveEmail saves an email to the database
func (s *SQLiteStorage) SaveEmail(email *Email) (int64, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Marshal JSON fields
	toJSON, _ := json.Marshal(email.To)
	ccJSON, _ := json.Marshal(email.CC)
	bccJSON, _ := json.Marshal(email.BCC)
	headersJSON, _ := json.Marshal(email.Headers)

	// Insert email
	result, err := tx.Exec(`
		INSERT INTO emails (
			message_id, from_address, to_addresses, cc_addresses, bcc_addresses,
			subject, body_plain, body_html, headers, size, received_at, read
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		email.MessageID, email.From, string(toJSON), string(ccJSON), string(bccJSON),
		email.Subject, email.BodyPlain, email.BodyHTML, string(headersJSON),
		email.Size, email.ReceivedAt, email.Read,
	)
	if err != nil {
		return 0, err
	}

	emailID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert attachments
	for _, att := range email.Attachments {
		if attWithData, ok := interface{}(&att).(*Attachment); ok {
			_, err = tx.Exec(`
				INSERT INTO attachments (email_id, filename, content_type, size, data)
				VALUES (?, ?, ?, ?, ?)
			`, emailID, att.Filename, att.ContentType, att.Size, attWithData.Data)
			if err != nil {
				return 0, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return emailID, nil
}

// GetEmail retrieves an email by ID
func (s *SQLiteStorage) GetEmail(id int64) (*Email, error) {
	var email Email
	var toJSON, ccJSON, bccJSON, headersJSON string

	err := s.db.QueryRow(`
		SELECT id, message_id, from_address, to_addresses, cc_addresses, bcc_addresses,
		       subject, body_plain, body_html, headers, size, received_at, read
		FROM emails WHERE id = ?
	`, id).Scan(
		&email.ID, &email.MessageID, &email.From, &toJSON, &ccJSON, &bccJSON,
		&email.Subject, &email.BodyPlain, &email.BodyHTML, &headersJSON,
		&email.Size, &email.ReceivedAt, &email.Read,
	)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON fields
	json.Unmarshal([]byte(toJSON), &email.To)
	json.Unmarshal([]byte(ccJSON), &email.CC)
	json.Unmarshal([]byte(bccJSON), &email.BCC)
	json.Unmarshal([]byte(headersJSON), &email.Headers)

	// Get attachments metadata
	rows, err := s.db.Query(`
		SELECT id, filename, content_type, size
		FROM attachments WHERE email_id = ?
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var att AttachmentMeta
		if err := rows.Scan(&att.ID, &att.Filename, &att.ContentType, &att.Size); err != nil {
			return nil, err
		}
		email.Attachments = append(email.Attachments, att)
	}

	return &email, nil
}

// ListEmails retrieves a paginated list of emails with optional filtering
func (s *SQLiteStorage) ListEmails(filter *EmailFilter, limit, offset int) (*EmailListResult, error) {
	query := `
		SELECT id, message_id, from_address, to_addresses, cc_addresses, bcc_addresses,
		       subject, body_plain, body_html, headers, size, received_at, read
		FROM emails WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM emails WHERE 1=1"
	args := []interface{}{}

	// Apply filters
	if filter != nil {
		if filter.From != "" {
			query += " AND from_address LIKE ?"
			countQuery += " AND from_address LIKE ?"
			args = append(args, "%"+filter.From+"%")
		}
		if filter.To != "" {
			query += " AND to_addresses LIKE ?"
			countQuery += " AND to_addresses LIKE ?"
			args = append(args, "%"+filter.To+"%")
		}
		if filter.Subject != "" {
			query += " AND subject LIKE ?"
			countQuery += " AND subject LIKE ?"
			args = append(args, "%"+filter.Subject+"%")
		}
		if filter.Since != nil {
			query += " AND received_at >= ?"
			countQuery += " AND received_at >= ?"
			args = append(args, filter.Since)
		}
		if filter.Until != nil {
			query += " AND received_at <= ?"
			countQuery += " AND received_at <= ?"
			args = append(args, filter.Until)
		}
	}

	// Get total count
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Add ordering and pagination
	query += " ORDER BY received_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Execute query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails := []*Email{}
	for rows.Next() {
		var email Email
		var toJSON, ccJSON, bccJSON, headersJSON string

		err := rows.Scan(
			&email.ID, &email.MessageID, &email.From, &toJSON, &ccJSON, &bccJSON,
			&email.Subject, &email.BodyPlain, &email.BodyHTML, &headersJSON,
			&email.Size, &email.ReceivedAt, &email.Read,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		json.Unmarshal([]byte(toJSON), &email.To)
		json.Unmarshal([]byte(ccJSON), &email.CC)
		json.Unmarshal([]byte(bccJSON), &email.BCC)
		json.Unmarshal([]byte(headersJSON), &email.Headers)

		emails = append(emails, &email)
	}

	return &EmailListResult{
		Emails: emails,
		Total:  total,
	}, nil
}

// SearchEmails performs full-text search on emails
func (s *SQLiteStorage) SearchEmails(query string, limit, offset int) (*EmailListResult, error) {
	var sqlQuery string
	var countQuery string
	var args []interface{}

	if s.hasFTS5 {
		// Use FTS5 for search
		sqlQuery = `
			SELECT e.id, e.message_id, e.from_address, e.to_addresses, e.cc_addresses, e.bcc_addresses,
			       e.subject, e.body_plain, e.body_html, e.headers, e.size, e.received_at, e.read
			FROM emails e
			JOIN emails_fts fts ON e.id = fts.rowid
			WHERE emails_fts MATCH ?
			ORDER BY e.received_at DESC
			LIMIT ? OFFSET ?
		`
		countQuery = "SELECT COUNT(*) FROM emails_fts WHERE emails_fts MATCH ?"
		args = []interface{}{query, limit, offset}
	} else {
		// Fallback to LIKE-based search
		sqlQuery = `
			SELECT id, message_id, from_address, to_addresses, cc_addresses, bcc_addresses,
			       subject, body_plain, body_html, headers, size, received_at, read
			FROM emails
			WHERE subject LIKE ? OR from_address LIKE ? OR to_addresses LIKE ? OR body_plain LIKE ?
			ORDER BY received_at DESC
			LIMIT ? OFFSET ?
		`
		countQuery = `
			SELECT COUNT(*) FROM emails
			WHERE subject LIKE ? OR from_address LIKE ? OR to_addresses LIKE ? OR body_plain LIKE ?
		`
		searchPattern := "%" + query + "%"
		args = []interface{}{searchPattern, searchPattern, searchPattern, searchPattern, limit, offset}
	}

	rows, err := s.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	emails := []*Email{}
	for rows.Next() {
		var email Email
		var toJSON, ccJSON, bccJSON, headersJSON string

		err := rows.Scan(
			&email.ID, &email.MessageID, &email.From, &toJSON, &ccJSON, &bccJSON,
			&email.Subject, &email.BodyPlain, &email.BodyHTML, &headersJSON,
			&email.Size, &email.ReceivedAt, &email.Read,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		json.Unmarshal([]byte(toJSON), &email.To)
		json.Unmarshal([]byte(ccJSON), &email.CC)
		json.Unmarshal([]byte(bccJSON), &email.BCC)
		json.Unmarshal([]byte(headersJSON), &email.Headers)

		emails = append(emails, &email)
	}

	// Get total count for search
	var total int64
	if s.hasFTS5 {
		err = s.db.QueryRow(countQuery, query).Scan(&total)
	} else {
		searchPattern := "%" + query + "%"
		err = s.db.QueryRow(countQuery, searchPattern, searchPattern, searchPattern, searchPattern).Scan(&total)
	}
	if err != nil {
		total = int64(len(emails))
	}

	return &EmailListResult{
		Emails: emails,
		Total:  total,
	}, nil
}

// DeleteEmail deletes an email by ID
func (s *SQLiteStorage) DeleteEmail(id int64) error {
	result, err := s.db.Exec("DELETE FROM emails WHERE id = ?", id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// DeleteAllEmails deletes all emails
func (s *SQLiteStorage) DeleteAllEmails() error {
	_, err := s.db.Exec("DELETE FROM emails")
	return err
}

// GetEmailCount returns the total number of emails
func (s *SQLiteStorage) GetEmailCount() (int64, error) {
	var count int64
	err := s.db.QueryRow("SELECT COUNT(*) FROM emails").Scan(&count)
	return count, err
}

// GetAttachment retrieves an attachment by ID
func (s *SQLiteStorage) GetAttachment(id int64) (*Attachment, error) {
	var att Attachment
	err := s.db.QueryRow(`
		SELECT id, filename, content_type, size, data
		FROM attachments WHERE id = ?
	`, id).Scan(&att.ID, &att.Filename, &att.ContentType, &att.Size, &att.Data)

	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &att, nil
}

// DeleteOldEmails deletes emails older than the specified time
func (s *SQLiteStorage) DeleteOldEmails(before time.Time) (int64, error) {
	result, err := s.db.Exec("DELETE FROM emails WHERE received_at < ?", before)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// DeleteExcessEmails deletes emails exceeding the maximum count
func (s *SQLiteStorage) DeleteExcessEmails(maxCount int) (int64, error) {
	result, err := s.db.Exec(`
		DELETE FROM emails WHERE id IN (
			SELECT id FROM emails
			ORDER BY received_at DESC
			LIMIT -1 OFFSET ?
		)
	`, maxCount)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
