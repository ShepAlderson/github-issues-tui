package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Issue represents a GitHub issue
type Issue struct {
	Number       int      `json:"number"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	State        string   `json:"state"`
	Author       string   `json:"author"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
	ClosedAt     string   `json:"closed_at"`
	CommentCount int      `json:"comment_count"`
	Labels       []string `json:"labels"`
	Assignees    []string `json:"assignees"`
}

// Comment represents a GitHub issue comment
type Comment struct {
	ID          int    `json:"id"`
	IssueNumber int    `json:"issue_number"`
	Body        string `json:"body"`
	Author      string `json:"author"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// InitializeSchema creates the database schema if it doesn't exist
// Returns a database connection
func InitializeSchema(dbPath string) (*sql.DB, error) {
	// Ensure the path has the correct scheme for libsql
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Use the libsql:// scheme for local files
	connectionString := fmt.Sprintf("file:%s", absPath)

	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create issues table
	createIssuesTable := `
	CREATE TABLE IF NOT EXISTS issues (
		repo TEXT NOT NULL,
		number INTEGER NOT NULL,
		title TEXT NOT NULL,
		body TEXT,
		state TEXT NOT NULL,
		author TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		closed_at TEXT,
		comment_count INTEGER DEFAULT 0,
		labels TEXT,
		assignees TEXT,
		PRIMARY KEY (repo, number)
	);
	CREATE INDEX IF NOT EXISTS idx_issues_state ON issues(repo, state);
	CREATE INDEX IF NOT EXISTS idx_issues_updated ON issues(repo, updated_at);
	`

	if _, err := db.Exec(createIssuesTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create issues table: %w", err)
	}

	// Create comments table
	createCommentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
		repo TEXT NOT NULL,
		id INTEGER NOT NULL,
		issue_number INTEGER NOT NULL,
		body TEXT,
		author TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		PRIMARY KEY (repo, id),
		FOREIGN KEY (repo, issue_number) REFERENCES issues(repo, number) ON DELETE CASCADE
	);
	CREATE INDEX IF NOT EXISTS idx_comments_issue ON comments(repo, issue_number);
	`

	if _, err := db.Exec(createCommentsTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create comments table: %w", err)
	}

	// Create sync_metadata table for tracking last sync time
	createSyncMetadataTable := `
	CREATE TABLE IF NOT EXISTS sync_metadata (
		repo TEXT NOT NULL PRIMARY KEY,
		last_sync_at TEXT NOT NULL
	);
	`

	if _, err := db.Exec(createSyncMetadataTable); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create sync_metadata table: %w", err)
	}

	return db, nil
}

// SaveIssue saves or updates an issue in the database
func SaveIssue(db *sql.DB, repo string, issue Issue) error {
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
		INSERT INTO issues (
			repo, number, title, body, state, author, created_at, updated_at, closed_at, comment_count, labels, assignees
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo, number) DO UPDATE SET
			title = excluded.title,
			body = excluded.body,
			state = excluded.state,
			author = excluded.author,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at,
			closed_at = excluded.closed_at,
			comment_count = excluded.comment_count,
			labels = excluded.labels,
			assignees = excluded.assignees
	`

	_, err = db.Exec(
		query,
		repo,
		issue.Number,
		issue.Title,
		issue.Body,
		issue.State,
		issue.Author,
		issue.CreatedAt,
		issue.UpdatedAt,
		issue.ClosedAt,
		issue.CommentCount,
		string(labelsJSON),
		string(assigneesJSON),
	)

	if err != nil {
		return fmt.Errorf("failed to save issue %d: %w", issue.Number, err)
	}

	return nil
}

// SaveComment saves or updates a comment in the database
func SaveComment(db *sql.DB, repo string, comment Comment) error {
	query := `
		INSERT INTO comments (
			repo, id, issue_number, body, author, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(repo, id) DO UPDATE SET
			body = excluded.body,
			author = excluded.author,
			created_at = excluded.created_at,
			updated_at = excluded.updated_at
	`

	_, err := db.Exec(
		query,
		repo,
		comment.ID,
		comment.IssueNumber,
		comment.Body,
		comment.Author,
		comment.CreatedAt,
		comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save comment %d: %w", comment.ID, err)
	}

	return nil
}

// GetIssueCount returns the number of issues for a repository
func GetIssueCount(db *sql.DB, repo string) (int, error) {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM issues WHERE repo = ?", repo)
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count issues: %w", err)
	}
	return count, nil
}

// GetCommentCount returns the number of comments for a repository
func GetCommentCount(db *sql.DB, repo string) (int, error) {
	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM comments WHERE repo = ?", repo)
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count comments: %w", err)
	}
	return count, nil
}

// parseLabels converts a JSON string to a slice of labels
func parseLabels(jsonData string) []string {
	if jsonData == "" {
		return []string{}
	}

	var labels []string
	if err := json.Unmarshal([]byte(jsonData), &labels); err != nil {
		return []string{}
	}
	return labels
}

// parseAssignees converts a JSON string to a slice of assignees
func parseAssignees(jsonData string) []string {
	if jsonData == "" {
		return []string{}
	}

	var assignees []string
	if err := json.Unmarshal([]byte(jsonData), &assignees); err != nil {
		return []string{}
	}
	return assignees
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

// ListIssue represents an issue for display in the list view
type ListIssue struct {
	Number       int
	Title        string
	Author       string
	CreatedAt    string
	UpdatedAt    string
	State        string
	CommentCount int
	Labels       []string
	Assignees    []string
}

// Validate checks if the issue has valid data
func (i ListIssue) Validate() error {
	if i.Number == 0 {
		return fmt.Errorf("issue number cannot be zero")
	}
	if i.Title == "" {
		return fmt.Errorf("issue title cannot be empty")
	}
	if i.Author == "" {
		return fmt.Errorf("issue author cannot be empty")
	}
	return nil
}

// ListIssues returns all issues for a repository
func ListIssues(db *sql.DB, repo string) ([]ListIssue, error) {
	return ListIssuesSorted(db, repo, "updated", true)
}

// ListIssuesSorted returns issues sorted by the specified field
func ListIssuesSorted(db *sql.DB, repo string, sortField string, descending bool) ([]ListIssue, error) {
	// Map sort field to column name
	var orderBy string
	switch sortField {
	case "number":
		orderBy = "number"
	case "created":
		orderBy = "created_at"
	case "updated", "":
		orderBy = "updated_at"
	case "comments":
		orderBy = "comment_count"
	default:
		orderBy = "updated_at"
	}

	// Build order direction
	direction := "ASC"
	if descending {
		direction = "DESC"
	}

	query := fmt.Sprintf(
		"SELECT number, title, author, created_at, updated_at, state, comment_count, labels, assignees FROM issues WHERE repo = ? ORDER BY %s %s",
		orderBy,
		direction,
	)

	rows, err := db.Query(query, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []ListIssue
	for rows.Next() {
		var issue ListIssue
		var labelsJSON, assigneesJSON string
		if err := rows.Scan(&issue.Number, &issue.Title, &issue.Author, &issue.CreatedAt, &issue.UpdatedAt, &issue.State, &issue.CommentCount, &labelsJSON, &assigneesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		issue.Labels = parseLabels(labelsJSON)
		issue.Assignees = parseAssignees(assigneesJSON)
		issues = append(issues, issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}

// ListIssuesByState returns issues filtered by state
func ListIssuesByState(db *sql.DB, repo string, state string) ([]ListIssue, error) {
	query := `SELECT number, title, author, created_at, updated_at, state, comment_count, labels, assignees
		FROM issues WHERE repo = ? AND state = ? ORDER BY updated_at DESC`

	rows, err := db.Query(query, repo, state)
	if err != nil {
		return nil, fmt.Errorf("failed to query issues: %w", err)
	}
	defer rows.Close()

	var issues []ListIssue
	for rows.Next() {
		var issue ListIssue
		var labelsJSON, assigneesJSON string
		if err := rows.Scan(&issue.Number, &issue.Title, &issue.Author, &issue.CreatedAt, &issue.UpdatedAt, &issue.State, &issue.CommentCount, &labelsJSON, &assigneesJSON); err != nil {
			return nil, fmt.Errorf("failed to scan issue: %w", err)
		}
		issue.Labels = parseLabels(labelsJSON)
		issue.Assignees = parseAssignees(assigneesJSON)
		issues = append(issues, issue)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issues: %w", err)
	}

	return issues, nil
}

// IssueDetail represents an issue with all details for the detail view
type IssueDetail struct {
	Number       int
	Title        string
	Body         string
	State        string
	Author       string
	CreatedAt    string
	UpdatedAt    string
	ClosedAt     string
	CommentCount int
	Labels       []string
	Assignees    []string
}

// GetIssueDetail returns a single issue with full details
func GetIssueDetail(db *sql.DB, repo string, number int) (*IssueDetail, error) {
	query := `SELECT number, title, body, state, author, created_at, updated_at, closed_at, comment_count, labels, assignees
		FROM issues WHERE repo = ? AND number = ?`

	row := db.QueryRow(query, repo, number)

	var detail IssueDetail
	var labelsJSON, assigneesJSON string
	err := row.Scan(&detail.Number, &detail.Title, &detail.Body, &detail.State, &detail.Author,
		&detail.CreatedAt, &detail.UpdatedAt, &detail.ClosedAt, &detail.CommentCount, &labelsJSON, &assigneesJSON)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("issue #%d not found in %s", number, repo)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query issue: %w", err)
	}

	detail.Labels = parseLabels(labelsJSON)
	detail.Assignees = parseAssignees(assigneesJSON)

	return &detail, nil
}

// GetCommentsForIssue returns all comments for a specific issue, sorted chronologically
func GetCommentsForIssue(db *sql.DB, repo string, issueNumber int) ([]Comment, error) {
	query := `SELECT id, issue_number, body, author, created_at, updated_at
		FROM comments
		WHERE repo = ? AND issue_number = ?
		ORDER BY created_at ASC`

	rows, err := db.Query(query, repo, issueNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.IssueNumber, &comment.Body, &comment.Author, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comments: %w", err)
	}

	return comments, nil
}

// FormatDate formats a date string for display
func FormatDate(dateStr string) string {
	if dateStr == "" {
		return ""
	}
	// Parse RFC3339 format and return just the date portion
	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02")
}

// FormatRelativeTime formats a time as a relative string (e.g., "5 minutes ago")
// It takes two time.Time parameters: the time to format and the current reference time
func FormatRelativeTime(t time.Time, now time.Time) string {
	duration := now.Sub(t)

	if duration < time.Minute {
		return "just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	// For days, use a simple approximation
	days := int(duration.Hours() / 24)

	if days < 7 {
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	if days < 30 {
		weeks := days / 7
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	if days < 365 {
		months := days / 30
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := days / 365
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// GetLastSyncTime returns the last sync timestamp for a repository
// Returns empty string if no sync has been performed
func GetLastSyncTime(db *sql.DB, repo string) (string, error) {
	var lastSync string
	row := db.QueryRow("SELECT last_sync_at FROM sync_metadata WHERE repo = ?", repo)
	err := row.Scan(&lastSync)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get last sync time: %w", err)
	}
	return lastSync, nil
}

// SaveLastSyncTime saves or updates the last sync timestamp for a repository
func SaveLastSyncTime(db *sql.DB, repo string, timestamp string) error {
	query := `
		INSERT INTO sync_metadata (repo, last_sync_at)
		VALUES (?, ?)
		ON CONFLICT(repo) DO UPDATE SET
			last_sync_at = excluded.last_sync_at
	`
	_, err := db.Exec(query, repo, timestamp)
	if err != nil {
		return fmt.Errorf("failed to save last sync time: %w", err)
	}
	return nil
}

// DeleteIssue removes an issue and its comments from the database
func DeleteIssue(db *sql.DB, repo string, issueNumber int) error {
	// Delete comments first (due to foreign key constraint)
	_, err := db.Exec("DELETE FROM comments WHERE repo = ? AND issue_number = ?", repo, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to delete comments: %w", err)
	}

	// Delete the issue
	_, err = db.Exec("DELETE FROM issues WHERE repo = ? AND number = ?", repo, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to delete issue: %w", err)
	}

	return nil
}

// DeleteCommentsForIssue removes all comments for a specific issue
func DeleteCommentsForIssue(db *sql.DB, repo string, issueNumber int) error {
	_, err := db.Exec("DELETE FROM comments WHERE repo = ? AND issue_number = ?", repo, issueNumber)
	if err != nil {
		return fmt.Errorf("failed to delete comments: %w", err)
	}
	return nil
}

// GetAllIssueNumbers returns all issue numbers for a repository
func GetAllIssueNumbers(db *sql.DB, repo string) ([]int, error) {
	rows, err := db.Query("SELECT number FROM issues WHERE repo = ?", repo)
	if err != nil {
		return nil, fmt.Errorf("failed to query issue numbers: %w", err)
	}
	defer rows.Close()

	var numbers []int
	for rows.Next() {
		var num int
		if err := rows.Scan(&num); err != nil {
			return nil, fmt.Errorf("failed to scan issue number: %w", err)
		}
		numbers = append(numbers, num)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating issue numbers: %w", err)
	}

	return numbers, nil
}
