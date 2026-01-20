package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"github.com/shepbook/ghissues/internal/github"
)

// Open opens a database connection and creates the schema if needed
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys (disabled by default in SQLite)
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if err := createSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

// createSchema creates the tables if they don't exist
func createSchema(db *sql.DB) error {
	// Create issues table
	issuesSQL := `
	CREATE TABLE IF NOT EXISTS issues (
		number INTEGER PRIMARY KEY,
		owner TEXT NOT NULL,
		repo TEXT NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		state TEXT NOT NULL,
		author_login TEXT NOT NULL,
		author_id INTEGER,
		author_email TEXT,
		author_name TEXT,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		comment_count INTEGER DEFAULT 0,
		html_url TEXT,
		synced_at TEXT NOT NULL
	);`

	if _, err := db.Exec(issuesSQL); err != nil {
		return fmt.Errorf("failed to create issues table: %w", err)
	}

	// Create labels table
	labelsSQL := `
	CREATE TABLE IF NOT EXISTS labels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		issue_number INTEGER NOT NULL,
		name TEXT NOT NULL,
		color TEXT,
		FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
	);`

	if _, err := db.Exec(labelsSQL); err != nil {
		return fmt.Errorf("failed to create labels table: %w", err)
	}

	// Create assignees table
	assigneesSQL := `
	CREATE TABLE IF NOT EXISTS assignees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		issue_number INTEGER NOT NULL,
		login TEXT NOT NULL,
		user_id INTEGER,
		FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
	);`

	if _, err := db.Exec(assigneesSQL); err != nil {
		return fmt.Errorf("failed to create assignees table: %w", err)
	}

	// Create comments table
	commentsSQL := `
	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY,
		issue_number INTEGER NOT NULL,
		body TEXT NOT NULL,
		author_login TEXT NOT NULL,
		author_id INTEGER,
		author_email TEXT,
		author_name TEXT,
		created_at TEXT NOT NULL,
		synced_at TEXT NOT NULL,
		FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
	);`

	if _, err := db.Exec(commentsSQL); err != nil {
		return fmt.Errorf("failed to create comments table: %w", err)
	}

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_issues_owner_repo ON issues(owner, repo);",
		"CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(state);",
		"CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues(updated_at);",
		"CREATE INDEX IF NOT EXISTS idx_labels_issue ON labels(issue_number);",
		"CREATE INDEX IF NOT EXISTS idx_assignees_issue ON assignees(issue_number);",
		"CREATE INDEX IF NOT EXISTS idx_comments_issue ON comments(issue_number);",
	}

	for _, idx := range indexes {
		if _, err := db.Exec(idx); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// UpsertIssue inserts or updates an issue in the database
func UpsertIssue(db *sql.DB, owner, repo string, issue *github.Issue) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `
	INSERT INTO issues (number, owner, repo, title, body, state, author_login, author_id, author_email, author_name, created_at, updated_at, comment_count, html_url, synced_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(number) DO UPDATE SET
		title = excluded.title,
		body = excluded.body,
		state = excluded.state,
		author_login = excluded.author_login,
		author_id = excluded.author_id,
		author_email = excluded.author_email,
		author_name = excluded.author_name,
		updated_at = excluded.updated_at,
		comment_count = excluded.comment_count,
		html_url = excluded.html_url,
		synced_at = excluded.synced_at;`

	_, err := db.Exec(query,
		issue.Number,
		owner,
		repo,
		issue.Title,
		issue.Body,
		issue.State,
		issue.Author.Login,
		issue.Author.ID,
		issue.Author.Email,
		issue.Author.Name,
		issue.CreatedAt.Format(time.RFC3339),
		issue.UpdatedAt.Format(time.RFC3339),
		issue.Comments,
		issue.HTMLURL,
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert issue: %w", err)
	}

	return nil
}

// DeleteLabels deletes all labels for an issue
func DeleteLabels(db *sql.DB, issueNumber int) error {
	_, err := db.Exec("DELETE FROM labels WHERE issue_number = ?", issueNumber)
	return err
}

// InsertLabel inserts a label for an issue
func InsertLabel(db *sql.DB, issueNumber int, label *github.Label) error {
	query := `INSERT INTO labels (issue_number, name, color) VALUES (?, ?, ?);`
	_, err := db.Exec(query, issueNumber, label.Name, label.Color)
	return err
}

// DeleteAssignees deletes all assignees for an issue
func DeleteAssignees(db *sql.DB, issueNumber int) error {
	_, err := db.Exec("DELETE FROM assignees WHERE issue_number = ?", issueNumber)
	return err
}

// InsertAssignee inserts an assignee for an issue
func InsertAssignee(db *sql.DB, issueNumber int, assignee *github.User) error {
	query := `INSERT INTO assignees (issue_number, login, user_id) VALUES (?, ?, ?);`
	_, err := db.Exec(query, issueNumber, assignee.Login, assignee.ID)
	return err
}

// UpsertComment inserts or updates a comment in the database
func UpsertComment(db *sql.DB, issueNumber int, comment *github.Comment) error {
	now := time.Now().UTC().Format(time.RFC3339)

	query := `
	INSERT INTO comments (id, issue_number, body, author_login, author_id, author_email, author_name, created_at, synced_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		body = excluded.body,
		author_login = excluded.author_login,
		author_id = excluded.author_id,
		author_email = excluded.author_email,
		author_name = excluded.author_name,
		created_at = excluded.created_at,
		synced_at = excluded.synced_at;`

	_, err := db.Exec(query,
		comment.ID,
		issueNumber,
		comment.Body,
		comment.Author.Login,
		comment.Author.ID,
		comment.Author.Email,
		comment.Author.Name,
		comment.CreatedAt.Format(time.RFC3339),
		now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert comment: %w", err)
	}

	return nil
}

// IssueExists checks if an issue exists in the database
func IssueExists(db *sql.DB, owner, repo string, number int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM issues WHERE owner = ? AND repo = ? AND number = ?;`
	err := db.QueryRow(query, owner, repo, number).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetIssueCount returns the count of issues for a repository
func GetIssueCount(db *sql.DB, owner, repo string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM issues WHERE owner = ? AND repo = ?;`
	err := db.QueryRow(query, owner, repo).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetCommentCount returns the count of comments for an issue
func GetCommentCount(db *sql.DB, issueNumber int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM comments WHERE issue_number = ?;`
	err := db.QueryRow(query, issueNumber).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}