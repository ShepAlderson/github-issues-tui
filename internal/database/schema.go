package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

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

// IssueDetail represents a full issue with all its details
type IssueDetail struct {
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	Author       string    `json:"author"`
	AuthorURL    string    `json:"author_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	CommentCount int       `json:"comment_count"`
	State        string    `json:"state"`
	Labels       []string  `json:"labels"`
	Assignees    []string  `json:"assignees"`
}

// Comment represents a comment on an issue
type Comment struct {
	ID        int64     `json:"id"`
	Author    string    `json:"author"`
	AuthorURL string    `json:"author_url"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetIssueByNumber retrieves a full issue by its number
func (dm *DBManager) GetIssueByNumber(number int) (*IssueDetail, error) {
	query := `
		SELECT number, title, body, author, author_url, created_at, updated_at,
		       comment_count, state, labels, assignees
		FROM issues
		WHERE number = ?
	`

	var labelsJSON, assigneesJSON string
	issue := &IssueDetail{}

	err := dm.db.QueryRow(query, number).Scan(
		&issue.Number,
		&issue.Title,
		&issue.Body,
		&issue.Author,
		&issue.AuthorURL,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.CommentCount,
		&issue.State,
		&labelsJSON,
		&assigneesJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("issue #%d not found", number)
		}
		return nil, fmt.Errorf("failed to query issue #%d: %w", number, err)
	}

	// Parse JSON arrays for labels and assignees
	if labelsJSON != "" {
		if err := json.Unmarshal([]byte(labelsJSON), &issue.Labels); err != nil {
			return nil, fmt.Errorf("failed to parse labels for issue #%d: %w", number, err)
		}
	}

	if assigneesJSON != "" {
		if err := json.Unmarshal([]byte(assigneesJSON), &issue.Assignees); err != nil {
			return nil, fmt.Errorf("failed to parse assignees for issue #%d: %w", number, err)
		}
	}

	return issue, nil
}

// GetCommentsByIssueNumber retrieves all comments for an issue
func (dm *DBManager) GetCommentsByIssueNumber(issueNumber int) ([]Comment, error) {
	query := `
		SELECT id, author, author_url, body, created_at, updated_at
		FROM comments
		WHERE issue_number = ?
		ORDER BY created_at ASC
	`

	rows, err := dm.db.Query(query, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments for issue #%d: %w", issueNumber, err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		err := rows.Scan(
			&comment.ID,
			&comment.Author,
			&comment.AuthorURL,
			&comment.Body,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment row: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comment rows: %w", err)
	}

	return comments, nil
}

// GetAllIssues retrieves all issues for the list view
func (dm *DBManager) GetAllIssues() ([]struct {
	Number       int
	Title        string
	Author       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CommentCount int
}, error) {
	query := `
		SELECT number, title, author, created_at, updated_at, comment_count
		FROM issues
		ORDER BY updated_at DESC
	`

	rows, err := dm.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []struct {
		Number       int
		Title        string
		Author       string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		CommentCount int
	}

	for rows.Next() {
		var issue struct {
			Number       int
			Title        string
			Author       string
			CreatedAt    time.Time
			UpdatedAt    time.Time
			CommentCount int
		}
		err := rows.Scan(
				&issue.Number,
				&issue.Title,
				&issue.Author,
				&issue.CreatedAt,
				&issue.UpdatedAt,
				&issue.CommentCount,
			)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue row: %w", err)
		}
		issues = append(issues, issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issue rows: %w", err)
	}

	return issues, nil
}

// GetDB returns the underlying database connection
func (dm *DBManager) GetDB() *sql.DB {
	return dm.db
}

// GetLastSyncTime retrieves the last sync timestamp from metadata
func (dm *DBManager) GetLastSyncTime() (time.Time, error) {
	query := "SELECT value FROM metadata WHERE key = 'last_sync'"
	var value string
	err := dm.db.QueryRow(query).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			// No sync recorded yet, return zero time
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to query last sync time: %w", err)
	}

	// Parse the timestamp
	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse sync time %q: %w", value, err)
	}

	return parsedTime, nil
}

// UpsertIssue inserts or updates an issue in the database
func (dm *DBManager) UpsertIssue(issue *IssueDetail) error {
	// Convert labels and assignees to JSON
	labelsJSON, err := json.Marshal(issue.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	assigneesJSON, err := json.Marshal(issue.Assignees)
	if err != nil {
		return fmt.Errorf("failed to marshal assignees: %w", err)
	}

	query := `
		INSERT OR REPLACE INTO issues (
			number, title, body, author, author_url, created_at, updated_at,
			comment_count, state, labels, assignees
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = dm.db.Exec(query,
		issue.Number,
		issue.Title,
		issue.Body,
		issue.Author,
		issue.AuthorURL,
		issue.CreatedAt,
		issue.UpdatedAt,
		issue.CommentCount,
		issue.State,
		string(labelsJSON),
		string(assigneesJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert issue #%d: %w", issue.Number, err)
	}

	return nil
}

// UpsertComment inserts or updates a comment in the database
func (dm *DBManager) UpsertComment(comment *Comment, issueNumber int) error {
	query := `
		INSERT OR REPLACE INTO comments (
			id, issue_number, author, author_url, body, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := dm.db.Exec(query,
		comment.ID,
		issueNumber,
		comment.Author,
		comment.AuthorURL,
		comment.Body,
		comment.CreatedAt,
		comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert comment %d for issue #%d: %w", comment.ID, issueNumber, err)
	}

	return nil
}

// DeleteIssue removes an issue from the database
func (dm *DBManager) DeleteIssue(issueNumber int) error {
	// Comments will be deleted automatically due to ON DELETE CASCADE foreign key
	query := "DELETE FROM issues WHERE number = ?"
	_, err := dm.db.Exec(query, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to delete issue #%d: %w", issueNumber, err)
	}
	return nil
}
