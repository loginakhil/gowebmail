package storage

// schema contains the SQL schema for the database
const schema = `
-- Emails table
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

-- Indexes for emails table
CREATE INDEX IF NOT EXISTS idx_emails_from ON emails(from_address);
CREATE INDEX IF NOT EXISTS idx_emails_received ON emails(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_emails_subject ON emails(subject);

-- Attachments table
CREATE TABLE IF NOT EXISTS attachments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    content_type TEXT,
    size INTEGER,
    data BLOB,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE CASCADE
);

-- Index for attachments table
CREATE INDEX IF NOT EXISTS idx_attachments_email ON attachments(email_id);
`

// fts5Schema contains the FTS5 schema (optional, only if FTS5 is available)
const fts5Schema = `
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

CREATE TRIGGER IF NOT EXISTS emails_au AFTER UPDATE ON emails BEGIN
    DELETE FROM emails_fts WHERE rowid = old.id;
    INSERT INTO emails_fts(rowid, subject, from_address, to_addresses, body_plain)
    VALUES (new.id, new.subject, new.from_address, new.to_addresses, new.body_plain);
END;
`
