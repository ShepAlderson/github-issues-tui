package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shepbook/ghissues/internal/github"
	_ "github.com/tursodatabase/go-libsql"
)

// Store provides database operations for issues
type Store struct {
	db *sql.DB
}

// NewStore creates a new database store and initializes the schema
func NewStore(path string) (*Store, error) {
	db, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &Store{db: db}

	if err := store.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS issues (
			number INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			body TEXT,
			author TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			comment_count INTEGER DEFAULT 0,
			labels TEXT,
			assignees TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id TEXT PRIMARY KEY,
			issue_number INTEGER NOT NULL,
			body TEXT,
			author TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_issue ON comments(issue_number)`,
		`CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues(updated_at)`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute schema statement: %w", err)
		}
	}

	return nil
}

// SaveIssue saves or updates a single issue
func (s *Store) SaveIssue(ctx context.Context, issue *github.Issue) error {
	labelsJSON, err := json.Marshal(issue.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	assigneesJSON, err := json.Marshal(issue.Assignees)
	if err != nil {
		return fmt.Errorf("failed to marshal assignees: %w", err)
	}

	query := `
	INSERT OR REPLACE INTO issues (number, title, body, author, created_at, updated_at, comment_count, labels, assignees)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query,
		issue.Number,
		issue.Title,
		issue.Body,
		issue.Author.Login,
		issue.CreatedAt,
		issue.UpdatedAt,
		issue.CommentCount,
		string(labelsJSON),
		string(assigneesJSON),
	)

	return err
}

// SaveIssues saves multiple issues in a transaction
func (s *Store) SaveIssues(ctx context.Context, issues []github.Issue) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR REPLACE INTO issues (number, title, body, author, created_at, updated_at, comment_count, labels, assignees)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, issue := range issues {
		labelsJSON, err := json.Marshal(issue.Labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels: %w", err)
		}

		assigneesJSON, err := json.Marshal(issue.Assignees)
		if err != nil {
			return fmt.Errorf("failed to marshal assignees: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			issue.Number,
			issue.Title,
			issue.Body,
			issue.Author.Login,
			issue.CreatedAt,
			issue.UpdatedAt,
			issue.CommentCount,
			string(labelsJSON),
			string(assigneesJSON),
		)
		if err != nil {
			return fmt.Errorf("failed to save issue %d: %w", issue.Number, err)
		}
	}

	return tx.Commit()
}

// GetIssue retrieves an issue by number
func (s *Store) GetIssue(ctx context.Context, number int) (*github.Issue, error) {
	query := `
	SELECT number, title, body, author, created_at, updated_at, comment_count, labels, assignees
	FROM issues WHERE number = ?
	`

	row := s.db.QueryRowContext(ctx, query, number)
	return s.scanIssue(row)
}

// GetAllIssues retrieves all issues
func (s *Store) GetAllIssues(ctx context.Context) ([]github.Issue, error) {
	query := `
	SELECT number, title, body, author, created_at, updated_at, comment_count, labels, assignees
	FROM issues ORDER BY updated_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var issues []github.Issue
	for rows.Next() {
		issue, err := s.scanIssueRow(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, *issue)
	}

	return issues, rows.Err()
}

func (s *Store) scanIssue(row *sql.Row) (*github.Issue, error) {
	var issue github.Issue
	var author string
	var labelsJSON, assigneesJSON string

	err := row.Scan(
		&issue.Number,
		&issue.Title,
		&issue.Body,
		&author,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.CommentCount,
		&labelsJSON,
		&assigneesJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	issue.Author = github.User{Login: author}

	if labelsJSON != "" {
		if err := json.Unmarshal([]byte(labelsJSON), &issue.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}
	}

	if assigneesJSON != "" {
		if err := json.Unmarshal([]byte(assigneesJSON), &issue.Assignees); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignees: %w", err)
		}
	}

	return &issue, nil
}

func (s *Store) scanIssueRow(rows *sql.Rows) (*github.Issue, error) {
	var issue github.Issue
	var author string
	var labelsJSON, assigneesJSON string

	err := rows.Scan(
		&issue.Number,
		&issue.Title,
		&issue.Body,
		&author,
		&issue.CreatedAt,
		&issue.UpdatedAt,
		&issue.CommentCount,
		&labelsJSON,
		&assigneesJSON,
	)
	if err != nil {
		return nil, err
	}

	issue.Author = github.User{Login: author}

	if labelsJSON != "" {
		if err := json.Unmarshal([]byte(labelsJSON), &issue.Labels); err != nil {
			return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
		}
	}

	if assigneesJSON != "" {
		if err := json.Unmarshal([]byte(assigneesJSON), &issue.Assignees); err != nil {
			return nil, fmt.Errorf("failed to unmarshal assignees: %w", err)
		}
	}

	return &issue, nil
}

// SaveComments saves comments for an issue
func (s *Store) SaveComments(ctx context.Context, issueNumber int, comments []github.Comment) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing comments for this issue
	_, err = tx.ExecContext(ctx, "DELETE FROM comments WHERE issue_number = ?", issueNumber)
	if err != nil {
		return fmt.Errorf("failed to delete existing comments: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO comments (id, issue_number, body, author, created_at)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, comment := range comments {
		_, err = stmt.ExecContext(ctx,
			comment.ID,
			issueNumber,
			comment.Body,
			comment.Author.Login,
			comment.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to save comment %s: %w", comment.ID, err)
		}
	}

	return tx.Commit()
}

// GetComments retrieves comments for an issue
func (s *Store) GetComments(ctx context.Context, issueNumber int) ([]github.Comment, error) {
	query := `
	SELECT id, body, author, created_at
	FROM comments WHERE issue_number = ? ORDER BY created_at ASC
	`

	rows, err := s.db.QueryContext(ctx, query, issueNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []github.Comment
	for rows.Next() {
		var comment github.Comment
		var author string

		err := rows.Scan(&comment.ID, &comment.Body, &author, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}

		comment.Author = github.User{Login: author}
		comments = append(comments, comment)
	}

	return comments, rows.Err()
}

// ClearIssues removes all issues and comments
func (s *Store) ClearIssues(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM comments")
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, "DELETE FROM issues")
	return err
}

// GetLastSyncTime returns the last sync time, or zero time if never synced
func (s *Store) GetLastSyncTime(ctx context.Context) (time.Time, error) {
	var value string
	err := s.db.QueryRowContext(ctx, "SELECT value FROM metadata WHERE key = 'last_sync'").Scan(&value)

	if err == sql.ErrNoRows {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, value)
}

// SetLastSyncTime sets the last sync time
func (s *Store) SetLastSyncTime(ctx context.Context, t time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT OR REPLACE INTO metadata (key, value) VALUES ('last_sync', ?)",
		t.UTC().Format(time.RFC3339),
	)
	return err
}

// GetAllIssueNumbers returns all issue numbers currently in the database
func (s *Store) GetAllIssueNumbers(ctx context.Context) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT number FROM issues")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var numbers []int
	for rows.Next() {
		var num int
		if err := rows.Scan(&num); err != nil {
			return nil, err
		}
		numbers = append(numbers, num)
	}

	return numbers, rows.Err()
}

// DeleteIssues removes issues by their numbers, along with their comments
func (s *Store) DeleteIssues(ctx context.Context, numbers []int) error {
	if len(numbers) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, num := range numbers {
		// Delete comments first (foreign key)
		_, err := tx.ExecContext(ctx, "DELETE FROM comments WHERE issue_number = ?", num)
		if err != nil {
			return fmt.Errorf("failed to delete comments for issue %d: %w", num, err)
		}

		// Delete the issue
		_, err = tx.ExecContext(ctx, "DELETE FROM issues WHERE number = ?", num)
		if err != nil {
			return fmt.Errorf("failed to delete issue %d: %w", num, err)
		}
	}

	return tx.Commit()
}
