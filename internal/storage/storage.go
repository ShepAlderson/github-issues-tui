package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// Issue represents a GitHub issue
type Issue struct {
	Number    int
	Title     string
	Body      string
	Author    string
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt  *time.Time
	Comments  int
	Labels    string // Comma-separated label names
	Assignees string // Comma-separated assignee usernames
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID          int
	IssueNumber int
	Body        string
	Author      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// InitializeDatabase creates the database file and initializes the schema
func InitializeDatabase(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create issues table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS issues (
			number INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			body TEXT,
			author TEXT NOT NULL,
			state TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			closed_at TEXT,
			comments INTEGER NOT NULL DEFAULT 0,
			labels TEXT,
			assignees TEXT
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create issues table: %w", err)
	}

	// Create comments table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY,
			issue_number INTEGER NOT NULL,
			body TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create comments table: %w", err)
	}

	// Create metadata table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata table: %w", err)
	}

	return db, nil
}

// StoreIssue stores or updates an issue in the database
func StoreIssue(db *sql.DB, issue *Issue) error {
	closedAt := ""
	if issue.ClosedAt != nil {
		closedAt = issue.ClosedAt.Format(time.RFC3339)
	}

	query := `
		INSERT INTO issues (number, title, body, author, state, created_at, updated_at, closed_at, comments, labels, assignees)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(number) DO UPDATE SET
			title = excluded.title,
			body = excluded.body,
			author = excluded.author,
			state = excluded.state,
			updated_at = excluded.updated_at,
			closed_at = excluded.closed_at,
			comments = excluded.comments,
			labels = excluded.labels,
			assignees = excluded.assignees
	`

	_, err := db.Exec(query,
		issue.Number,
		issue.Title,
		issue.Body,
		issue.Author,
		issue.State,
		issue.CreatedAt.Format(time.RFC3339),
		issue.UpdatedAt.Format(time.RFC3339),
		closedAt,
		issue.Comments,
		issue.Labels,
		issue.Assignees,
	)

	if err != nil {
		return fmt.Errorf("failed to store issue: %w", err)
	}

	return nil
}

// StoreComment stores or updates a comment in the database
func StoreComment(db *sql.DB, comment *Comment) error {
	query := `
		INSERT INTO comments (id, issue_number, body, author, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			body = excluded.body,
			author = excluded.author,
			updated_at = excluded.updated_at
	`

	_, err := db.Exec(query,
		comment.ID,
		comment.IssueNumber,
		comment.Body,
		comment.Author,
		comment.CreatedAt.Format(time.RFC3339),
		comment.UpdatedAt.Format(time.RFC3339),
	)

	if err != nil {
		return fmt.Errorf("failed to store comment: %w", err)
	}

	return nil
}

// GetIssues retrieves all issues from the database
func GetIssues(db *sql.DB) ([]Issue, error) {
	query := `
		SELECT number, title, body, author, state, created_at, updated_at, closed_at, comments, labels, assignees
		FROM issues
		ORDER BY number
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []Issue
	for rows.Next() {
		var issue Issue
		var closedAt sql.NullString
		var createdAtStr, updatedAtStr string

		err := rows.Scan(
			&issue.Number,
			&issue.Title,
			&issue.Body,
			&issue.Author,
			&issue.State,
			&createdAtStr,
			&updatedAtStr,
			&closedAt,
			&issue.Comments,
			&issue.Labels,
			&issue.Assignees,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		issue.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		issue.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		if closedAt.Valid {
			parsedTime, err := time.Parse(time.RFC3339, closedAt.String)
			if err == nil {
				issue.ClosedAt = &parsedTime
			}
		}

		issues = append(issues, issue)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}

// GetCommentsForIssue retrieves all comments for a specific issue
func GetCommentsForIssue(db *sql.DB, issueNumber int) ([]Comment, error) {
	query := `
		SELECT id, issue_number, body, author, created_at, updated_at
		FROM comments
		WHERE issue_number = ?
		ORDER BY id
	`

	rows, err := db.Query(query, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		var createdAtStr, updatedAtStr string
		err := rows.Scan(
			&comment.ID,
			&comment.IssueNumber,
			&comment.Body,
			&comment.Author,
			&createdAtStr,
			&updatedAtStr,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comment.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		comment.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse updated_at: %w", err)
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// UpdateLastSync updates the last sync timestamp in the metadata table
func UpdateLastSync(db *sql.DB, syncTime time.Time) error {
	query := `
		INSERT INTO metadata (key, value) VALUES ('last_sync', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`

	_, err := db.Exec(query, syncTime.Format(time.RFC3339))
	if err != nil {
		return fmt.Errorf("failed to update last sync: %w", err)
	}

	return nil
}

// GetLastSync retrieves the last sync timestamp from the metadata table
func GetLastSync(db *sql.DB) (time.Time, error) {
	var value string
	err := db.QueryRow("SELECT value FROM metadata WHERE key = 'last_sync'").Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil
		}
		return time.Time{}, fmt.Errorf("failed to get last sync: %w", err)
	}

	parsedTime, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last sync time: %w", err)
	}

	return parsedTime, nil
}
