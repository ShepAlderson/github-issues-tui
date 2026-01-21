package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// DBManager manages database connections and operations
type DBManager struct {
	db *sql.DB
}

// NewDBManager creates a new database manager
func NewDBManager(dbPath string) (*DBManager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	return &DBManager{db: db}, nil
}

// Close closes the database connection
func (dm *DBManager) Close() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}

// InitializeSchema creates the necessary tables if they don't exist
func (dm *DBManager) InitializeSchema() error {
	// Create issues table
	issuesSQL := `
	CREATE TABLE IF NOT EXISTS issues (
		id INTEGER PRIMARY KEY,
		number INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		author TEXT NOT NULL,
		author_url TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		comment_count INTEGER DEFAULT 0,
		state TEXT NOT NULL,
		labels TEXT, -- JSON array of label names
		assignees TEXT -- JSON array of assignee logins
	);
	`
	if _, err := dm.db.Exec(issuesSQL); err != nil {
		return fmt.Errorf("failed to create issues table: %w", err)
	}

	// Create comments table
	commentsSQL := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY,
		issue_number INTEGER NOT NULL,
		author TEXT NOT NULL,
		author_url TEXT,
		body TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
	);
	`
	if _, err := dm.db.Exec(commentsSQL); err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	// Create metadata table for sync information
	metadataSQL := `
	CREATE TABLE IF NOT EXISTS metadata (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`
	if _, err := dm.db.Exec(metadataSQL); err != nil {
		return fmt.Errorf("failed to create metadata table: %w", err)
	}

	// Create indexes for performance
	indexSQL := `
	CREATE INDEX IF NOT EXISTS idx_issues_number ON issues(number);
	CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues(updated_at);
	CREATE INDEX IF NOT EXISTS idx_comments_issue_number ON comments(issue_number);
	CREATE INDEX IF NOT EXISTS idx_comments_created ON comments(created_at);
	`
	if _, err := dm.db.Exec(indexSQL); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// GetDB returns the underlying database connection
func (dm *DBManager) GetDB() *sql.DB {
	return dm.db
}