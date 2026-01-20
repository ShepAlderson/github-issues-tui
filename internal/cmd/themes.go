package cmd

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/git/github-issues-tui/internal/config"
)

// availableThemes returns a list of all available theme names
func availableThemes() []string {
	return []string{"default", "dracula", "gruvbox", "nord", "solarized-dark", "solarized-light"}
}

// RunThemesCommand runs the theme selection command
func RunThemesCommand(configPath string, input io.Reader, output io.Writer, nonInteractive bool, themeName string) error {
	// Load current config or create default
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		cfg = &config.Config{}
	}

	// If non-interactive mode, just set the theme
	if nonInteractive {
		if themeName == "" {
			return fmt.Errorf("theme name required in non-interactive mode")
		}

		// Validate theme name
		valid := false
		for _, t := range availableThemes() {
			if t == themeName {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid theme name: %s. Available themes: %s", themeName, strings.Join(availableThemes(), ", "))
		}

		cfg.Display.Theme = themeName
		if err := config.SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Fprintf(output, "Theme set to: %s\n", themeName)
		return nil
	}

	// Interactive mode - display themes and let user choose
	fmt.Fprintln(output, "=== Theme Selection ===")
	fmt.Fprintln(output)

	// Display all themes with color samples
	themes := availableThemes()
	for i, themeName := range themes {
		theme := config.GetTheme(themeName)

		// Create color sample bars
		accentBar := lipgloss.NewStyle().
			Background(theme.Accent).
			Width(5).
			Render(" ")

		headerBar := lipgloss.NewStyle().
			Background(theme.Header).
			Width(5).
			Render(" ")

		bgBar := lipgloss.NewStyle().
			Background(theme.Background).
			Width(5).
			Render(" ")

		successBar := lipgloss.NewStyle().
			Background(theme.Success).
			Width(5).
			Render(" ")

		errorBar := lipgloss.NewStyle().
			Background(theme.Error).
			Width(5).
			Render(" ")

		// Format theme preview line
		preview := fmt.Sprintf("%d. %-20s %s %s %s %s %s",
			i+1,
			themeName,
			accentBar,
			headerBar,
			bgBar,
			successBar,
			errorBar,
		)

		fmt.Fprintln(output, preview)
	}

	fmt.Fprintln(output)
	fmt.Fprintf(output, "Current theme: %s\n", getCurrentTheme(cfg))
	fmt.Fprintln(output)
	fmt.Fprint(output, "Select a theme (1-" + fmt.Sprintf("%d", len(themes)) + "), or press Enter to keep current: ")

	// Read user input
	reader := bufio.NewReader(input)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	choice = strings.TrimSpace(choice)

	// If user pressed Enter without input, keep current theme
	if choice == "" {
		fmt.Fprintln(output, "No theme selected. Keeping current theme.")
		return nil
	}

	// Parse choice
	var selectedTheme string
	for i, t := range themes {
		if fmt.Sprintf("%d", i+1) == choice {
			selectedTheme = t
			break
		}
	}

	if selectedTheme == "" {
		// Try direct theme name
		for _, t := range themes {
			if t == choice {
				selectedTheme = t
				break
			}
		}
	}

	if selectedTheme == "" {
		return fmt.Errorf("invalid selection: %s", choice)
	}

	// Save the selected theme
	cfg.Display.Theme = selectedTheme
	if err := config.SaveConfig(configPath, cfg); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Fprintf(output, "\nTheme saved: %s\n", selectedTheme)
	fmt.Fprintln(output)
	fmt.Fprintln(output, "The new theme will be applied the next time you run ghissues.")

	return nil
}

// getCurrentTheme returns the current theme name from config, or "default" if not set
func getCurrentTheme(cfg *config.Config) string {
	if cfg != nil && cfg.Display.Theme != "" {
		return cfg.Display.Theme
	}
	return "default"
}

// ListThemesCommand lists all available themes
func ListThemesCommand(output io.Writer) error {
	themes := availableThemes()

	fmt.Fprintln(output, "Available themes:")
	fmt.Fprintln(output)

	for _, themeName := range themes {
		theme := config.GetTheme(themeName)
		displayName := themeName

		// Add indicators for light/dark themes
		if strings.Contains(themeName, "dark") {
			displayName = fmt.Sprintf("%s (dark)", themeName)
		} else if strings.Contains(themeName, "light") {
			displayName = fmt.Sprintf("%s (light)", themeName)
		}

		fmt.Fprintf(output, "  - %s\n", displayName)
		fmt.Fprintf(output, "    Background: %v, Accent: %v\n", theme.Background, theme.Accent)
		fmt.Fprintln(output)
	}

	fmt.Fprintln(output, "Use 'ghissues themes' to preview and select a theme interactively.")
	fmt.Fprintln(output, "Or use 'ghissues --themes set <theme-name>' to set a theme directly.")

	return nil
}
