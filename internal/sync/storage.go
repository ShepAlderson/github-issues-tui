package sync

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// IssueStore handles storage and retrieval of issues in the database
type IssueStore struct {
	db *sql.DB
}

// NewIssueStore creates a new issue store
func NewIssueStore(dbPath string) (*IssueStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	store := &IssueStore{db: db}

	// Initialize schema
	if err := store.InitSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *IssueStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// InitSchema creates the necessary database tables if they don't exist
func (s *IssueStore) InitSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS issues (
			number INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			body TEXT,
			state TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			comment_count INTEGER DEFAULT 0
		);

		CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(state);
		CREATE INDEX IF NOT EXISTS idx_issues_author ON issues(author);

		CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY,
			issue_number INTEGER NOT NULL,
			body TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_comments_issue ON comments(issue_number);

		CREATE TABLE IF NOT EXISTS labels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_number INTEGER NOT NULL,
			name TEXT NOT NULL,
			FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE,
			UNIQUE(issue_number, name)
		);

		CREATE INDEX IF NOT EXISTS idx_labels_issue ON labels(issue_number);

		CREATE TABLE IF NOT EXISTS assignees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_number INTEGER NOT NULL,
			username TEXT NOT NULL,
			FOREIGN KEY (issue_number) REFERENCES issues(number) ON DELETE CASCADE,
			UNIQUE(issue_number, username)
		);

		CREATE INDEX IF NOT EXISTS idx_assignees_issue ON assignees(issue_number);
	`

	_, err := s.db.Exec(schema)
	return err
}

// StoreIssue stores or updates an issue in the database
func (s *IssueStore) StoreIssue(issue *Issue) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or replace issue
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO issues
		(number, title, body, state, author, created_at, updated_at, comment_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, issue.Number, issue.Title, issue.Body, issue.State, issue.Author,
		issue.CreatedAt, issue.UpdatedAt, issue.CommentCount)
	if err != nil {
		return fmt.Errorf("failed to store issue: %w", err)
	}

	// Delete existing labels and assignees (we'll re-insert them)
	_, err = tx.Exec("DELETE FROM labels WHERE issue_number = ?", issue.Number)
	if err != nil {
		return fmt.Errorf("failed to delete old labels: %w", err)
	}

	_, err = tx.Exec("DELETE FROM assignees WHERE issue_number = ?", issue.Number)
	if err != nil {
		return fmt.Errorf("failed to delete old assignees: %w", err)
	}

	// Insert labels
	for _, label := range issue.Labels {
		_, err = tx.Exec(`
			INSERT INTO labels (issue_number, name)
			VALUES (?, ?)
		`, issue.Number, label)
		if err != nil {
			return fmt.Errorf("failed to store label: %w", err)
		}
	}

	// Insert assignees
	for _, assignee := range issue.Assignees {
		_, err = tx.Exec(`
			INSERT INTO assignees (issue_number, username)
			VALUES (?, ?)
		`, issue.Number, assignee)
		if err != nil {
			return fmt.Errorf("failed to store assignee: %w", err)
		}
	}

	return tx.Commit()
}

// StoreComment stores or updates a comment in the database
func (s *IssueStore) StoreComment(comment *Comment) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO comments
		(id, issue_number, body, author, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, comment.ID, comment.IssueNumber, comment.Body, comment.Author,
		comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to store comment: %w", err)
	}

	return nil
}

// ClearData removes all data from the database
func (s *IssueStore) ClearData() error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete in reverse order of foreign key dependencies
	tables := []string{"comments", "labels", "assignees", "issues"}
	for _, table := range tables {
		_, err = tx.Exec("DELETE FROM " + table)
		if err != nil {
			return fmt.Errorf("failed to clear %s: %w", table, err)
		}
	}

	return tx.Commit()
}

// GetIssueCount returns the total number of issues in the database
func (s *IssueStore) GetIssueCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM issues").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count issues: %w", err)
	}
	return count, nil
}
