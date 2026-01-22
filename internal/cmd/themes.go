package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/themes"
	"github.com/spf13/cobra"
)

// newThemesCmd creates the themes subcommand
func newThemesCmd() *cobra.Command {
	var setTheme string
	var previewTheme string

	cmd := &cobra.Command{
		Use:   "themes",
		Short: "List, preview, and set color themes",
		Long: `List available color themes, preview a theme, or set the active theme.

Without flags, lists all available themes and indicates the current one.
Use --preview to see a theme without changing the config.
Use --set to change the active theme.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := GetConfigPath()

			// Load config (or create empty one for display)
			var currentTheme config.Theme
			var cfg *config.Config
			if config.Exists(path) {
				var err error
				cfg, err = config.Load(path)
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}
				currentTheme = cfg.Display.Theme
			}

			// Handle preview
			if previewTheme != "" {
				theme := config.Theme(previewTheme)
				if err := config.ValidateTheme(theme); err != nil {
					return fmt.Errorf("invalid theme: %s", previewTheme)
				}
				return previewThemeOutput(cmd, theme)
			}

			// Handle set
			if setTheme != "" {
				theme := config.Theme(setTheme)
				if err := config.ValidateTheme(theme); err != nil {
					return fmt.Errorf("invalid theme: %s", setTheme)
				}

				if cfg == nil {
					// Create new config if it doesn't exist
					cfg = config.New()
				}
				cfg.Display.Theme = theme

				if err := config.Save(cfg, path); err != nil {
					return fmt.Errorf("failed to save config: %w", err)
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Theme set to: %s\n", theme.DisplayName())
				return nil
			}

			// List themes
			return listThemes(cmd, currentTheme)
		},
	}

	cmd.Flags().StringVar(&setTheme, "set", "", "Set the active theme")
	cmd.Flags().StringVar(&previewTheme, "preview", "", "Preview a theme without changing config")

	return cmd
}

// listThemes lists all available themes
func listThemes(cmd *cobra.Command, currentTheme config.Theme) error {
	fmt.Fprintln(cmd.OutOrStdout(), "Available themes:")
	fmt.Fprintln(cmd.OutOrStdout())

	for _, themeName := range config.AllThemes() {
		theme := themes.GetTheme(themeName)
		styles := theme.Styles()

		indicator := "  "
		if themeName == currentTheme || (currentTheme == "" && themeName == config.ThemeDefault) {
			indicator = "* "
		}

		// Show theme name with its primary color
		name := styles.Title.Render(themeName.DisplayName())
		current := ""
		if themeName == currentTheme || (currentTheme == "" && themeName == config.ThemeDefault) {
			current = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" (current)")
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s%s%s\n", indicator, name, current)
	}

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), "Use 'ghissues themes --preview <name>' to preview a theme")
	fmt.Fprintln(cmd.OutOrStdout(), "Use 'ghissues themes --set <name>' to set your theme")

	return nil
}

// previewThemeOutput shows a preview of the given theme
func previewThemeOutput(cmd *cobra.Command, themeName config.Theme) error {
	theme := themes.GetTheme(themeName)
	styles := theme.Styles()

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintf(cmd.OutOrStdout(), "Preview: %s\n", styles.Title.Render(themeName.DisplayName()))
	fmt.Fprintln(cmd.OutOrStdout(), "─────────────────────────────────")
	fmt.Fprintln(cmd.OutOrStdout())

	// Sample TUI preview
	fmt.Fprintln(cmd.OutOrStdout(), styles.Title.Render("GitHub Issues"))
	fmt.Fprintln(cmd.OutOrStdout())

	// Header
	fmt.Fprintln(cmd.OutOrStdout(), styles.Header.Render("  #     Title                              Author"))
	fmt.Fprintln(cmd.OutOrStdout(), styles.Muted.Render("─────────────────────────────────────────────────"))

	// Sample issues
	fmt.Fprintln(cmd.OutOrStdout(), styles.Selected.Render("> #123  Fix login button alignment         @alice"))
	fmt.Fprintln(cmd.OutOrStdout(), styles.Normal.Render("  #124  Add dark mode support              @bob"))
	fmt.Fprintln(cmd.OutOrStdout(), styles.Normal.Render("  #125  Improve performance                @carol"))

	fmt.Fprintln(cmd.OutOrStdout())

	// Labels sample
	fmt.Fprint(cmd.OutOrStdout(), styles.Muted.Render("Labels: "))
	fmt.Fprint(cmd.OutOrStdout(), styles.Label.Render("bug"))
	fmt.Fprint(cmd.OutOrStdout(), " ")
	fmt.Fprint(cmd.OutOrStdout(), styles.Label.Render("enhancement"))
	fmt.Fprintln(cmd.OutOrStdout())

	// Status bar sample
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), styles.Status.Render("3 issues | Updated ↓ | j/k: nav | q: quit"))

	// Error sample
	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintln(cmd.OutOrStdout(), styles.Error.Render("Error: Connection timed out"))
	fmt.Fprintln(cmd.OutOrStdout(), styles.Warning.Render("Note: Run 'ghissues sync' to refresh"))

	fmt.Fprintln(cmd.OutOrStdout())
	fmt.Fprintf(cmd.OutOrStdout(), "To use this theme: ghissues themes --set %s\n", themeName)

	return nil
}
