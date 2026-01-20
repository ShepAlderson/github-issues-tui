package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shepbook/github-issues-tui/internal/sync"
	"github.com/shepbook/github-issues-tui/internal/tui"
	"golang.org/x/term"
)

// checkTerminalCapabilities verifies that stdin is a terminal
// and returns an error with a clear message if it's not
func checkTerminalCapabilities() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("stdin is not a terminal. TUI applications require an interactive terminal environment.\n\nPlease run this command in a terminal (not redirected from a file or pipe).")
	}
	return nil
}

// getTUIOptions returns bubbletea program options based on environment variables
// This allows testing different initialization flags for debugging purposes
func getTUIOptions() []tea.ProgramOption {
	// Always include stdin input configuration (critical for keyboard input)
	options := []tea.ProgramOption{
		tea.WithInput(os.Stdin),
	}

	// Parse GHISSUES_TUI_OPTIONS environment variable for additional flags
	tuiOpts := os.Getenv("GHISSUES_TUI_OPTIONS")
	if tuiOpts == "" {
		// Default: use alt screen
		options = append(options, tea.WithAltScreen())
		return options
	}

	// Parse comma-separated options
	flags := strings.Split(tuiOpts, ",")
	useAltScreen := true

	for _, flag := range flags {
		flag = strings.TrimSpace(flag)
		switch flag {
		case "nomouse":
			// Disable mouse support (useful for debugging)
			options = append(options, tea.WithMouseCellMotion())
		case "noaltscreen":
			// Disable alt screen (useful for debugging - output stays in terminal)
			useAltScreen = false
		}
	}

	if useAltScreen {
		options = append(options, tea.WithAltScreen())
	}

	return options
}

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

	// Check terminal capabilities before launching TUI
	if err := checkTerminalCapabilities(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Launch TUI with minimal config
	columns := []string{"number", "title", "state", "author", "created"}
	model := tui.NewModel(issues, columns, "created", false, store, lastSyncTime, "default")
	tuiOptions := getTUIOptions()
	p := tea.NewProgram(model, tuiOptions...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
		os.Exit(1)
	}
}
