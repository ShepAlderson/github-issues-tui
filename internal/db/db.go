package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// DB represents a database connection
type DB struct {
	conn *sql.DB
}

// Issue represents a GitHub issue
type Issue struct {
	Number       int
	Title        string
	Body         string
	State        string
	Author       string
	CreatedAt    string
	UpdatedAt    string
	CommentCount int
	Labels       []string
	Assignees    []string
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID        int64
	IssueNum  int
	Body      string
	Author    string
	CreatedAt string
}

// NewDB creates a new database connection
func NewDB(dbPath string) (*DB, error) {
	// Convert file paths to libsql URL format
	dbURL := dbPath
	if dbPath == ":memory:" {
		dbURL = "file::memory:?cache=shared"
	} else if !hasURLScheme(dbPath) {
		// If it's a regular file path without a scheme, add file:// prefix
		dbURL = "file:" + dbPath
	}

	conn, err := sql.Open("libsql", dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

// hasURLScheme checks if the path has a URL scheme
func hasURLScheme(path string) bool {
	return len(path) >= 5 && (path[:5] == "file:" || path[:5] == "http:" || path[:6] == "https:" || path[:4] == "ws:" || path[:5] == "wss:" || path[:9] == "libsql://")
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// createTables creates the necessary database tables
func (db *DB) createTables() error {
	// Create issues table
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS issues (
			number INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			body TEXT,
			state TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			comment_count INTEGER DEFAULT 0,
			labels TEXT, -- JSON array of strings
			assignees TEXT -- JSON array of strings
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create issues table: %w", err)
	}

	// Create comments table
	_, err = db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY,
			issue_num INTEGER NOT NULL,
			body TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (issue_num) REFERENCES issues (number) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	// Create index on comments for faster lookups
	_, err = db.conn.Exec(`
		CREATE INDEX IF NOT EXISTS idx_comments_issue_num ON comments (issue_num)
	`)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// StoreIssue stores an issue in the database
func (db *DB) StoreIssue(issue *Issue) error {
	// Marshal labels and assignees to JSON
	labelsJSON, err := json.Marshal(issue.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	assigneesJSON, err := json.Marshal(issue.Assignees)
	if err != nil {
		return fmt.Errorf("failed to marshal assignees: %w", err)
	}

	// Use INSERT OR REPLACE to handle both new inserts and updates
	_, err = db.conn.Exec(`
		INSERT OR REPLACE INTO issues (
			number, title, body, state, author, created_at, updated_at,
			comment_count, labels, assignees
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		issue.Number, issue.Title, issue.Body, issue.State, issue.Author,
		issue.CreatedAt, issue.UpdatedAt, issue.CommentCount,
		string(labelsJSON), string(assigneesJSON))

	if err != nil {
		return fmt.Errorf("failed to store issue: %w", err)
	}

	return nil
}

// GetIssue retrieves an issue by number
func (db *DB) GetIssue(number int) (*Issue, error) {
	var issue Issue
	var labelsJSON, assigneesJSON string

	err := db.conn.QueryRow(`
		SELECT number, title, body, state, author, created_at, updated_at,
		       comment_count, labels, assignees
		FROM issues
		WHERE number = ?
	`, number).Scan(
		&issue.Number, &issue.Title, &issue.Body, &issue.State,
		&issue.Author, &issue.CreatedAt, &issue.UpdatedAt,
		&issue.CommentCount, &labelsJSON, &assigneesJSON)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("issue not found: %d", number)
		}
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}

	// Unmarshal labels
	if err := json.Unmarshal([]byte(labelsJSON), &issue.Labels); err != nil {
		return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
	}

	// Unmarshal assignees
	if err := json.Unmarshal([]byte(assigneesJSON), &issue.Assignees); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assignees: %w", err)
	}

	return &issue, nil
}

// StoreComment stores a comment in the database
func (db *DB) StoreComment(comment *Comment) error {
	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO comments (id, issue_num, body, author, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, comment.ID, comment.IssueNum, comment.Body, comment.Author, comment.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to store comment: %w", err)
	}

	return nil
}

// GetComments retrieves all comments for an issue
func (db *DB) GetComments(issueNum int) ([]*Comment, error) {
	rows, err := db.conn.Query(`
		SELECT id, issue_num, body, author, created_at
		FROM comments
		WHERE issue_num = ?
		ORDER BY created_at ASC
	`, issueNum)

	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.IssueNum, &comment.Body,
			&comment.Author, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, &comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// GetAllOpenIssues retrieves all open issues
func (db *DB) GetAllOpenIssues() ([]*Issue, error) {
	rows, err := db.conn.Query(`
		SELECT number, title, body, state, author, created_at, updated_at,
		       comment_count, labels, assignees
		FROM issues
		WHERE state = 'open'
		ORDER BY created_at ASC
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to get open issues: %w", err)
	}
	defer rows.Close()

	var issues []*Issue
	for rows.Next() {
		var issue Issue
		var labelsJSON, assigneesJSON string

		err := rows.Scan(&issue.Number, &issue.Title, &issue.Body,
			&issue.State, &issue.Author, &issue.CreatedAt, &issue.UpdatedAt,
			&issue.CommentCount, &labelsJSON, &assigneesJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		// Unmarshal labels
		if err := json.Unmarshal([]byte(labelsJSON), &issue.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}

		// Unmarshal assignees
		if err := json.Unmarshal([]byte(assigneesJSON), &issue.Assignees); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignees: %w", err)
		}

		issues = append(issues, &issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}

// GetIssuesForDisplay retrieves issues with minimal fields for display
// defaulting to created_at ascending
func (db *DB) GetIssuesForDisplay() ([]*Issue, error) {
	return db.GetIssuesForDisplaySorted("created_at", false)
}

// GetIssuesForDisplaySorted retrieves issues with minimal fields for display
func (db *DB) GetIssuesForDisplaySorted(sortField string, descending bool) ([]*Issue, error) {
	// Validate sort field to prevent SQL injection
	validFields := map[string]bool{
		"updated_at":    true,
		"created_at":    true,
		"number":        true,
		"comment_count": true,
	}

	if !validFields[sortField] {
		// Default to updated_at descending if invalid field
		sortField = "updated_at"
		descending = true
	}

	// Build query with validated sort field
	orderDirection := "ASC"
	if descending {
		orderDirection = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT number, title, author, created_at, comment_count
		FROM issues
		WHERE state = 'open'
		ORDER BY %s %s
	`, sortField, orderDirection)

	rows, err := db.conn.Query(query)

	if err != nil {
		return nil, fmt.Errorf("failed to get issues for display: %w", err)
	}
	defer rows.Close()

	var issues []*Issue
	for rows.Next() {
		var issue Issue
		err := rows.Scan(&issue.Number, &issue.Title, &issue.Author, &issue.CreatedAt, &issue.CommentCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		issues = append(issues, &issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}

// ClearAllIssues removes all issues and comments from the database
func (db *DB) ClearAllIssues() error {
	// Delete comments first due to foreign key constraint
	_, err := db.conn.Exec(`DELETE FROM comments`)
	if err != nil {
		return fmt.Errorf("failed to clear comments: %w", err)
	}

	// Then delete issues
	_, err = db.conn.Exec(`DELETE FROM issues`)
	if err != nil {
		return fmt.Errorf("failed to clear issues: %w", err)
	}

	return nil
}
