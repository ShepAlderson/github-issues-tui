package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/git/github-issues-tui/internal/db"
)

func TestNewCommentsModel(t *testing.T) {
	// Create a test database with an issue and comments
	dbPath := "/tmp/ghissues_comments_test.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue
	testIssue := &db.Issue{
		Number:       42,
		Title:        "Test Issue with Comments",
		Body:         "Test issue body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-20T14:45:00Z",
		CommentCount: 3,
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Insert test comments
	comments := []*db.Comment{
		{
			ID:        1,
			IssueNum:  42,
			Body:      "First comment",
			Author:    "user1",
			CreatedAt: "2024-01-15T11:00:00Z",
		},
		{
			ID:        2,
			IssueNum:  42,
			Body:      "# Second comment\n\nWith markdown!",
			Author:    "user2",
			CreatedAt: "2024-01-15T12:00:00Z",
		},
		{
			ID:        3,
			IssueNum:  42,
			Body:      "Third comment",
			Author:    "user1",
			CreatedAt: "2024-01-15T13:00:00Z",
		},
	}

	for _, comment := range comments {
		if err := database.StoreComment(comment); err != nil {
			t.Fatalf("Failed to store comment: %v", err)
		}
	}

	// Test creating comments model
	testIssueFull, _ := database.GetIssue(42)
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	// Verify issue loaded
	if model.issue == nil {
		t.Fatal("Expected issue to be loaded")
	}

	if model.issue.Number != 42 {
		t.Errorf("Expected issue number 42, got %d", model.issue.Number)
	}

	// Verify comments loaded
	if len(model.comments) != 3 {
		t.Errorf("Expected 3 comments, got %d", len(model.comments))
	}

	// Verify markdown toggle state
	if model.showRendered {
		t.Error("Expected showRendered to be false initially")
	}
}

func TestNewCommentsModel_NoComments(t *testing.T) {
	// Create a test database with an issue but no comments
	dbPath := "/tmp/ghissues_comments_empty.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue without comments
	testIssue := &db.Issue{
		Number:    42,
		Title:     "Test Issue No Comments",
		Body:      "Test issue body",
		State:     "open",
		Author:    "testuser",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-20T14:45:00Z",
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Get the issue from DB to verify it works
	testIssueFull, _ := database.GetIssue(42)

	// Test creating comments model
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	// Verify no comments loaded
	if len(model.comments) != 0 {
		t.Errorf("Expected 0 comments, got %d", len(model.comments))
	}
}

func TestCommentsModel_MarkdownToggle(t *testing.T) {
	// Create a test database with an issue and comment
	dbPath := "/tmp/ghissues_comments_toggle.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue
	testIssue := &db.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     "open",
		Author:    "testuser",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Insert test comment with markdown
	comment := &db.Comment{
		ID:        1,
		IssueNum:  1,
		Body:      "# Header\n\n**Bold comment** and *italic comment*",
		Author:    "commenter",
		CreatedAt: "2024-01-15T11:00:00Z",
	}

	if err := database.StoreComment(comment); err != nil {
		t.Fatalf("Failed to store comment: %v", err)
	}

	testIssueFull, _ := database.GetIssue(1)
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	// Initial state should show raw markdown
	if model.showRendered {
		t.Error("Expected raw markdown initially")
	}

	// Test pressing 'm' to toggle to rendered view
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(*CommentsModel)

	if !model.showRendered {
		t.Error("Expected rendered markdown after pressing 'm'")
	}

	// Test pressing 'm' again to toggle back to raw
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(*CommentsModel)

	if model.showRendered {
		t.Error("Expected raw markdown after second press of 'm'")
	}
}

func TestCommentsModel_Quit(t *testing.T) {
	dbPath := "/tmp/ghissues_comments_quit.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	testIssue := &db.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "Test body",
		State:     "open",
		Author:    "testuser",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	testIssueFull, _ := database.GetIssue(1)
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	// Test quitting with 'q'
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(*CommentsModel)

	if !model.quitting {
		t.Error("Expected model to be quitting after pressing 'q'")
	}
	if cmd == nil {
		t.Error("Expected quit command after pressing 'q'")
	}

	// Reset and test quitting with Ctrl+C
	model.quitting = false
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model = updated.(*CommentsModel)

	if !model.quitting {
		t.Error("Expected model to be quitting after Ctrl+C")
	}
	if cmd == nil {
		t.Error("Expected quit command after Ctrl+C")
	}

	// Reset and test quitting with 'esc'
	model.quitting = false
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(*CommentsModel)

	if !model.quitting {
		t.Error("Expected model to be quitting after pressing 'esc'")
	}
	if cmd == nil {
		t.Error("Expected quit command after pressing 'esc'")
	}
}

func TestCommentsModel_HeaderContent(t *testing.T) {
	// Create a test database with an issue and comments
	dbPath := "/tmp/ghissues_comments_header.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue
	testIssue := &db.Issue{
		Number:       42,
		Title:        "Test Issue with Comments",
		Body:         "Test issue body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-20T14:45:00Z",
		CommentCount: 2,
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Insert test comments
	comments := []*db.Comment{
		{
			ID:        1,
			IssueNum:  42,
			Body:      "First comment body",
			Author:    "user1",
			CreatedAt: "2024-01-15T11:00:00Z",
		},
		{
			ID:        2,
			IssueNum:  42,
			Body:      "Second comment body with markdown",
			Author:    "user2",
			CreatedAt: "2024-01-15T12:00:00Z",
		},
	}

	for _, comment := range comments {
		if err := database.StoreComment(comment); err != nil {
			t.Fatalf("Failed to store comment: %v", err)
		}
	}

	testIssueFull, _ := database.GetIssue(42)
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	view := model.View()

	// Check header contains issue number and title
	if !strings.Contains(view, "Comments for #42") {
		t.Error("Expected issue number (#42) in header")
	}
	if !strings.Contains(view, "Test Issue with Comments") {
		t.Error("Expected issue title in header")
	}

	// Check view contains all comments
	if !strings.Contains(view, "user1") {
		t.Error("Expected author 'user1' in view")
	}
	if !strings.Contains(view, "user2") {
		t.Error("Expected author 'user2' in view")
	}

	// Check comment bodies are present
	if !strings.Contains(view, "First comment body") {
		t.Error("Expected first comment body in view")
	}
	if !strings.Contains(view, "Second comment body with markdown") {
		t.Error("Expected second comment body in view")
	}
}

func TestCommentsModel_ChronologicalOrder(t *testing.T) {
	// Create a test database with an issue and comments out of order
	dbPath := "/tmp/ghissues_comments_order.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue
	testIssue := &db.Issue{
		Number:       1,
		Title:        "Test Issue",
		Body:         "Test issue body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-15T10:30:00Z",
		CommentCount: 3,
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Insert comments in non-chronological order
	comments := []*db.Comment{
		{
			ID:        3,
			IssueNum:  1,
			Body:      "Most recent comment",
			Author:    "user3",
			CreatedAt: "2024-01-15T15:00:00Z", // Last
		},
		{
			ID:        1,
			IssueNum:  1,
			Body:      "Oldest comment",
			Author:    "user1",
			CreatedAt: "2024-01-15T11:00:00Z", // First
		},
		{
			ID:        2,
			IssueNum:  1,
			Body:      "Middle comment",
			Author:    "user2",
			CreatedAt: "2024-01-15T13:00:00Z", // Middle
		},
	}

	for _, comment := range comments {
		if err := database.StoreComment(comment); err != nil {
			t.Fatalf("Failed to store comment: %v", err)
		}
	}

	testIssueFull, _ := database.GetIssue(1)
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	// Verify comments are sorted chronologically (oldest first)
	if len(model.comments) != 3 {
		t.Fatalf("Expected 3 comments, got %d", len(model.comments))
	}

	// Check that comments are in chronological order
	if model.comments[0].CreatedAt != "2024-01-15T11:00:00Z" {
		t.Errorf("Expected first comment at 11:00, got %s", model.comments[0].CreatedAt)
	}
	if model.comments[1].CreatedAt != "2024-01-15T13:00:00Z" {
		t.Errorf("Expected second comment at 13:00, got %s", model.comments[1].CreatedAt)
	}
	if model.comments[2].CreatedAt != "2024-01-15T15:00:00Z" {
		t.Errorf("Expected third comment at 15:00, got %s", model.comments[2].CreatedAt)
	}
}

func TestCommentsModel_Scroll(t *testing.T) {
	// Create a test database with an issue and many comments
	dbPath := "/tmp/ghissues_comments_scroll.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue
	testIssue := &db.Issue{
		Number:       1,
		Title:        "Test Issue with Many Comments",
		Body:         "Test issue body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-15T10:30:00Z",
		CommentCount: 50,
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Insert many comments to enable scrolling
	for i := 1; i <= 50; i++ {
		comment := &db.Comment{
			ID:        int64(i),
			IssueNum:  1,
			Body:      "This is a comment body that will make the content scrollable\n\nWith multiple lines of text to ensure scrolling works properly",
			Author:    "user",
			CreatedAt: "2024-01-15T10:30:00Z",
		}
		if err := database.StoreComment(comment); err != nil {
			t.Fatalf("Failed to store comment: %v", err)
		}
	}

	testIssueFull, _ := database.GetIssue(1)
	model, err := NewCommentsModel(dbPath, testIssueFull)
	if err != nil {
		t.Fatalf("Failed to create comments model: %v", err)
	}
	defer model.Close()

	// Set viewport size to enable scrolling
	model.viewport.Width = 80
	model.viewport.Height = 15
	model.viewport.SetContent(model.getContent())

	// Initial scroll position should be 0
	if model.viewport.YOffset != 0 {
		t.Errorf("Expected initial scroll position 0, got %d", model.viewport.YOffset)
	}

	// Test scrolling down with arrow key
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(*CommentsModel)

	if model.viewport.YOffset == 0 {
		t.Error("Expected scroll position to change after pressing down arrow")
	}

	// Test scrolling up with arrow key
	prevScroll := model.viewport.YOffset
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(*CommentsModel)

	if model.viewport.YOffset >= prevScroll {
		t.Error("Expected scroll position to decrease after pressing up arrow")
	}

	// Test scrolling down with 'j'
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(*CommentsModel)

	if model.viewport.YOffset == 0 {
		t.Error("Expected scroll position to change after pressing 'j'")
	}

	// Test scrolling up with 'k'
	prevScroll = model.viewport.YOffset
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(*CommentsModel)

	if model.viewport.YOffset >= prevScroll {
		t.Error("Expected scroll position to decrease after pressing 'k'")
	}
}
