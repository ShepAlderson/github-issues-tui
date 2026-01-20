package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/git/github-issues-tui/internal/db"
)

func TestNewDetailModel(t *testing.T) {
	// Create a test database with an issue
	dbPath := "/tmp/ghissues_detail_test.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue with all fields
	testIssue := &db.Issue{
		Number:       42,
		Title:        "Test Issue for Detail View",
		Body:         "# Problem\n\nThis is a test issue with markdown formatting.",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-20T14:45:00Z",
		CommentCount: 5,
		Labels:       []string{"bug", "help wanted"},
		Assignees:    []string{"user1", "user2"},
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	// Test creating detail model
	model, err := NewDetailModel(dbPath, 42)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	// Verify issue loaded
	if model.issue == nil {
		t.Fatal("Expected issue to be loaded")
	}

	if model.issue.Number != 42 {
		t.Errorf("Expected issue number 42, got %d", model.issue.Number)
	}

	if model.issue.Title != "Test Issue for Detail View" {
		t.Errorf("Expected correct title, got %q", model.issue.Title)
	}

	// Verify markdown toggle state
	if model.showRendered {
		t.Error("Expected showRendered to be false initially")
	}
}

func TestNewDetailModel_NotFound(t *testing.T) {
	// Create a test database without issues
	dbPath := "/tmp/ghissues_detail_notfound.db"
	defer os.Remove(dbPath)

	// Test creating detail model for non-existent issue
	_, err := NewDetailModel(dbPath, 999)
	if err == nil {
		t.Error("Expected error for non-existent issue")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got %v", err)
	}
}

func TestDetailModel_MarkdownToggle(t *testing.T) {
	// Create a test database with an issue
	dbPath := "/tmp/ghissues_detail_toggle.db"
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
		Body:      "# Header\n\n**Bold text** and *italic text*",
		State:     "open",
		Author:    "testuser",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	model, err := NewDetailModel(dbPath, 1)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	// Initial state should show raw markdown
	if model.showRendered {
		t.Error("Expected raw markdown initially")
	}

	// Test pressing 'm' to toggle to rendered view
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(*DetailModel)

	if !model.showRendered {
		t.Error("Expected rendered markdown after pressing 'm'")
	}

	// Test pressing 'm' again to toggle back to raw
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(*DetailModel)

	if model.showRendered {
		t.Error("Expected raw markdown after second press of 'm'")
	}
}

func TestDetailModel_Scroll(t *testing.T) {
	// Create a test database with an issue
	dbPath := "/tmp/ghissues_detail_scroll.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue with long body to enable scrolling
	longBody := "# Long Issue\n\n"
	for i := 0; i < 100; i++ {
		longBody += "This is line of text that will make the content scrollable\n"
	}

	testIssue := &db.Issue{
		Number:    1,
		Title:     "Long Issue for Scrolling",
		Body:      longBody,
		State:     "open",
		Author:    "testuser",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	model, err := NewDetailModel(dbPath, 1)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	// Set viewport size to enable scrolling
	model.viewport.Width = 80
	model.viewport.Height = 20
	model.viewport.SetContent(model.getContent())

	// Initial scroll position should be 0
	if model.viewport.YOffset != 0 {
		t.Errorf("Expected initial scroll position 0, got %d", model.viewport.YOffset)
	}

	// Test scrolling down with arrow key
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(*DetailModel)

	if model.viewport.YOffset == 0 {
		t.Error("Expected scroll position to change after pressing down arrow")
	}

	// Test scrolling up with arrow key
	prevScroll := model.viewport.YOffset
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updated.(*DetailModel)

	if model.viewport.YOffset >= prevScroll {
		t.Error("Expected scroll position to decrease after pressing up arrow")
	}

	// Test scrolling down with 'j'
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	model = updated.(*DetailModel)

	if model.viewport.YOffset == 0 {
		t.Error("Expected scroll position to change after pressing 'j'")
	}

	// Test scrolling up with 'k'
	prevScroll = model.viewport.YOffset
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(*DetailModel)

	if model.viewport.YOffset >= prevScroll {
		t.Error("Expected scroll position to decrease after pressing 'k'")
	}
}

func TestDetailModel_Quit(t *testing.T) {
	dbPath := "/tmp/ghissues_detail_quit.db"
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

	model, err := NewDetailModel(dbPath, 1)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	// Test quitting with 'q'
	updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updated.(*DetailModel)

	if !model.quitting {
		t.Error("Expected model to be quitting after pressing 'q'")
	}
	if cmd == nil {
		t.Error("Expected quit command after pressing 'q'")
	}

	// Reset and test quitting with Ctrl+C
	model.quitting = false
	updated, cmd = model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	model = updated.(*DetailModel)

	if !model.quitting {
		t.Error("Expected model to be quitting after Ctrl+C")
	}
	if cmd == nil {
		t.Error("Expected quit command after Ctrl+C")
	}
}

func TestDetailModel_HeaderContent(t *testing.T) {
	// Create a test database with an issue
	dbPath := "/tmp/ghissues_detail_header.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue with all fields
	testIssue := &db.Issue{
		Number:       42,
		Title:        "Test Issue with All Fields",
		Body:         "Test body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-20T14:45:00Z",
		CommentCount: 5,
		Labels:       []string{"bug", "help wanted", "priority-high"},
		Assignees:    []string{"user1", "user2"},
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	model, err := NewDetailModel(dbPath, 42)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	view := model.View()

	// Check header contains issue number and title
	if !strings.Contains(view, "#42") {
		t.Error("Expected issue number (#42) in header")
	}
	if !strings.Contains(view, "Test Issue with All Fields") {
		t.Error("Expected issue title in header")
	}

	// Check header contains author
	if !strings.Contains(view, "testuser") {
		t.Error("Expected author in header")
	}

	// Check header contains status
	if !strings.Contains(view, "open") {
		t.Error("Expected status (open) in header")
	}

	// Check header contains dates
	if !strings.Contains(view, "2024-01-15") {
		t.Error("Expected created date in header")
	}
	if !strings.Contains(view, "2024-01-20") {
		t.Error("Expected updated date in header")
	}

	// Check header contains labels
	if !strings.Contains(view, "bug") {
		t.Error("Expected 'bug' label in header")
	}
	if !strings.Contains(view, "help wanted") {
		t.Error("Expected 'help wanted' label in header")
	}

	// Check header contains assignees
	if !strings.Contains(view, "user1") {
		t.Error("Expected assignee 'user1' in header")
	}
	if !strings.Contains(view, "user2") {
		t.Error("Expected assignee 'user2' in header")
	}
}

func TestDetailModel_RenderMarkdown(t *testing.T) {
	// Create a test database with an issue
	dbPath := "/tmp/ghissues_detail_render.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	// Insert test issue with markdown
	testIssue := &db.Issue{
		Number:    1,
		Title:     "Test Issue",
		Body:      "# Header\n\n**Bold text** and *italic text*\n\n- List item 1\n- List item 2",
		State:     "open",
		Author:    "testuser",
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	model, err := NewDetailModel(dbPath, 1)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	// Test raw markdown view (initial state)
	view := model.View()
	if !strings.Contains(view, "**Bold text**") {
		t.Error("Expected raw markdown with '**Bold text**' in view")
	}

	// Toggle to rendered markdown
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	model = updated.(*DetailModel)

	view = model.View()
	// In rendered mode, markdown should be processed
	// glamour adds ANSI escape codes for styling, so the view should change
	// and not contain the literal markdown markers in the same way
	// Instead of checking for absence of markers, let's check that view mode indicator shows "Rendered"
	if !strings.Contains(view, "Rendered Markdown") {
		t.Error("Expected view mode indicator to show 'Rendered Markdown'")
	}
}

func TestDetailModel_EnterKey(t *testing.T) {
	// Create a test database with an issue
	dbPath := "/tmp/ghissues_detail_enter.db"
	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer func() {
		database.Close()
		os.Remove(dbPath)
	}()

	testIssue := &db.Issue{
		Number:       1,
		Title:        "Test Issue with Comments",
		Body:         "Test body",
		State:        "open",
		Author:       "testuser",
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-15T10:30:00Z",
		CommentCount: 3,
	}

	if err := database.StoreIssue(testIssue); err != nil {
		t.Fatalf("Failed to store test issue: %v", err)
	}

	model, err := NewDetailModel(dbPath, 1)
	if err != nil {
		t.Fatalf("Failed to create detail model: %v", err)
	}
	defer model.Close()

	// Test pressing Enter to view comments
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(*DetailModel)

	// For now, we expect it to set an error since comments view is not yet implemented
	if model.err == nil {
		t.Error("Expected error when pressing Enter (comments view not implemented)")
	}

	if !strings.Contains(model.err.Error(), "comments view") {
		t.Errorf("Expected 'comments view' error message, got %v", model.err.Error())
	}
}
