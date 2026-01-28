package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/shepbook/ghissues/internal/config"
)

// MainModel represents the main application state
type MainModel struct {
	config *config.Config
}

func NewMainModel(cfg *config.Config) MainModel {
	return MainModel{
		config: cfg,
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m MainModel) View() string {

	msg := "✨ ghissues is configured!\n\n"
	if m.config != nil && m.config.Default.Repository != "" {
		msg += fmt.Sprintf("Repository: %s\n", m.config.Default.Repository)
	}
	msg += "\nThe full TUI will be available in a future user story.\n"
	msg += "\nPress 'q' or Ctrl+C to quit.\n"
	return msg
}

func main() {
	// Handle subcommands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			if err := runConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "version", "-v", "--version":
			fmt.Println("ghissues version 0.1.0")
			os.Exit(0)
		case "help", "-h", "--help":
			printHelp()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
			printHelp()
			os.Exit(1)
		}
	}

	// Check if config exists, run setup if not
	if !config.Exists() {
		fmt.Println("🚀 First-time setup required!")
		fmt.Println()

		cfg, err := config.RunSetup()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup cancelled or failed: %v\n", err)
			os.Exit(1)
		}

		// Re-run with the new config
		p := tea.NewProgram(NewMainModel(cfg))
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load existing config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Run main application
	p := tea.NewProgram(NewMainModel(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

func runConfig() error {
	fmt.Println("🚀 Re-running first-time setup...")
	fmt.Println()

	_, err := config.RunSetup()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("✅ Configuration updated successfully!")
	return nil
}

func printHelp() {
	help := `ghissues - A terminal UI for GitHub issues

Usage:
  ghissues              Run the application (setup if first run)
  ghissues config       Configure repository and authentication
  ghissues help         Show this help message
  ghissues version      Show version

Configuration:
  The configuration is stored at ~/.config/ghissues/config.toml

First-Time Setup:
  On first run, ghissues will prompt you for:
  1. The GitHub repository (owner/repo format)
  2. Authentication method (environment variable or config file token)

Keybindings (when TUI is ready):
  j, ↓    Move down
  k, ↑    Move up
  ?       Show help
  q       Quit
`
	fmt.Println(help)
}
