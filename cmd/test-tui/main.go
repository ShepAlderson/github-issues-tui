package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/github-issues-tui/internal/sync"
	"github.com/shepbook/github-issues-tui/internal/tui"
)

func main() {
	dbPath := "./test-db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	// Load issues from database
	store, err := sync.NewIssueStore(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	issues, err := store.LoadIssues()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to load issues: %v\n", err)
		os.Exit(1)
	}

	lastSyncTime, _ := store.GetLastSyncTime()

	// Launch TUI with minimal config
	columns := []string{"number", "title", "state", "author", "created"}
	model := tui.NewModel(issues, columns, "created", false, store, lastSyncTime, "default")
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
