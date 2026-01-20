package cmd

import (
	"fmt"
	"io"

	"github.com/shepbook/git/github-issues-tui/internal/config"
	"github.com/shepbook/git/github-issues-tui/internal/prompt"
)

// RunConfigCommand runs the interactive configuration setup
func RunConfigCommand(configPath string, input io.Reader, output io.Writer) error {
	fmt.Fprintln(output, "=== GitHub Issues Configuration Setup ===")
	fmt.Fprintln(output)

	// Run interactive setup
	repository, token, err := prompt.RunInteractiveSetup(input, output)
	if err != nil {
		return fmt.Errorf("setup failed: %w", err)
	}

	// Create config struct
	cfg := &config.Config{
		Repository: repository,
		Token:      token,
	}
	// Set default display columns
	cfg.Display.Columns = config.GetDefaultDisplayColumns()
	// Set default sort preferences
	cfg.Display.Sort = config.GetDefaultSort()

	// Save configuration
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Fprintln(output)
	fmt.Fprintf(output, "Configuration saved to: %s\n", configPath)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "Next steps:")
	if token == "" {
		fmt.Fprintln(output, "- Ensure you have gh CLI installed and authenticated")
		fmt.Fprintln(output, "  Run: gh auth login")
	}
	fmt.Fprintln(output, "- Run 'ghissues' to launch the TUI")

	return nil
}
