package sync

import (
	"database/sql"
	"fmt"
	"time"

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

		CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
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

// LoadIssues loads all issues from the database, sorted by updated_at DESC (most recent first)
func (s *IssueStore) LoadIssues() ([]*Issue, error) {
	query := `
		SELECT number, title, body, state, author, created_at, updated_at, comment_count
		FROM issues
		ORDER BY updated_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []*Issue
	for rows.Next() {
		issue := &Issue{}
		err := rows.Scan(
			&issue.Number,
			&issue.Title,
			&issue.Body,
			&issue.State,
			&issue.Author,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.CommentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}

		// Load labels for this issue
		labelRows, err := s.db.Query("SELECT name FROM labels WHERE issue_number = ? ORDER BY name", issue.Number)
		if err != nil {
			return nil, fmt.Errorf("failed to query labels: %w", err)
		}
		for labelRows.Next() {
			var label string
			if err := labelRows.Scan(&label); err != nil {
				labelRows.Close()
				return nil, fmt.Errorf("failed to scan label: %w", err)
			}
			issue.Labels = append(issue.Labels, label)
		}
		labelRows.Close()

		// Load assignees for this issue
		assigneeRows, err := s.db.Query("SELECT username FROM assignees WHERE issue_number = ? ORDER BY username", issue.Number)
		if err != nil {
			return nil, fmt.Errorf("failed to query assignees: %w", err)
		}
		for assigneeRows.Next() {
			var assignee string
			if err := assigneeRows.Scan(&assignee); err != nil {
				assigneeRows.Close()
				return nil, fmt.Errorf("failed to scan assignee: %w", err)
			}
			issue.Assignees = append(issue.Assignees, assignee)
		}
		assigneeRows.Close()

		issues = append(issues, issue)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}

// LoadComments loads all comments for a specific issue, sorted by created_at ASC (chronological order)
func (s *IssueStore) LoadComments(issueNumber int) ([]*Comment, error) {
	query := `
		SELECT id, issue_number, body, author, created_at, updated_at
		FROM comments
		WHERE issue_number = ?
		ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []*Comment
	for rows.Next() {
		comment := &Comment{}
		err := rows.Scan(
			&comment.ID,
			&comment.IssueNumber,
			&comment.Body,
			&comment.Author,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// GetLastSyncTime retrieves the last sync timestamp from metadata
func (s *IssueStore) GetLastSyncTime() (time.Time, error) {
	var timeStr string
	err := s.db.QueryRow("SELECT value FROM metadata WHERE key = ?", "last_sync_time").Scan(&timeStr)
	if err == sql.ErrNoRows {
		// No sync time recorded yet - return zero time
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to query last sync time: %w", err)
	}

	// Parse the time string
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last sync time: %w", err)
	}

	return t, nil
}

// SetLastSyncTime updates the last sync timestamp in metadata
func (s *IssueStore) SetLastSyncTime(t time.Time) error {
	timeStr := t.Format(time.RFC3339)
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO metadata (key, value)
		VALUES (?, ?)
	`, "last_sync_time", timeStr)
	if err != nil {
		return fmt.Errorf("failed to set last sync time: %w", err)
	}
	return nil
}

// RemoveClosedIssues removes all closed issues from the database
// This is used during incremental sync to clean up issues that have been closed
func (s *IssueStore) RemoveClosedIssues() error {
	_, err := s.db.Exec("DELETE FROM issues WHERE state = ?", "closed")
	if err != nil {
		return fmt.Errorf("failed to remove closed issues: %w", err)
	}
	return nil
}
