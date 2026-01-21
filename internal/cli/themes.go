package cli

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/shepbook/github-issues-tui/internal/config"
	"github.com/shepbook/github-issues-tui/internal/tui"
	"github.com/spf13/cobra"
)

func newThemesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "themes",
		Short: "Manage color themes",
		Long:  "Preview and select color themes for the TUI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfgMgr := config.NewManager()

			// Check if config exists
			exists, err := cfgMgr.Exists()
			if err != nil {
				return fmt.Errorf("failed to check config: %w", err)
			}

			if !exists {
				fmt.Println("No configuration found. Please run 'ghissues' first to create a config.")
				fmt.Println("Or run 'ghissues config' to create a configuration.")
				return nil
			}

			// Load current config
			cfg, err := cfgMgr.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create theme manager
			tm := tui.NewThemeManager()

			// Show current theme
			currentTheme, exists := tm.GetThemeByName(cfg.Display.Theme)
			if !exists {
				fmt.Printf("Current theme '%s' not found. Using 'default'.\n", cfg.Display.Theme)
				cfg.Display.Theme = "default"
				currentTheme, _ = tm.GetThemeByName("default")
			}

			fmt.Printf("Current theme: %s (%s)\n\n", currentTheme.Name, currentTheme.Description)

			// Get all available themes
			themeNames := tm.GetThemeNames()
			themeItems := make([]string, 0, len(themeNames))
			for _, name := range themeNames {
				theme, _ := tm.GetThemeByName(name)
				themeItems = append(themeItems, fmt.Sprintf("%s - %s", name, theme.Description))
			}

			// Create interactive selector
			prompt := promptui.Select{
				Label: "Select a theme to preview or apply",
				Items: themeItems,
				Size:  10,
				Searcher: func(input string, index int) bool {
					item := themeItems[index]
					return strings.Contains(strings.ToLower(item), strings.ToLower(input))
				},
				StartInSearchMode: true,
			}

			index, _, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("theme selection cancelled: %w", err)
			}

			// Get selected theme name
			selectedName := themeNames[index]
			selectedTheme, _ := tm.GetThemeByName(selectedName)

			fmt.Printf("\nSelected: %s (%s)\n", selectedTheme.Name, selectedTheme.Description)

			// Show theme preview
			showThemePreview(selectedTheme)

			// Ask if user wants to apply the theme
			applyPrompt := promptui.Prompt{
				Label:     fmt.Sprintf("Apply theme '%s'", selectedTheme.Name),
				IsConfirm: true,
			}

			_, err = applyPrompt.Run()
			if err != nil {
				// User chose not to apply
				fmt.Println("Theme not applied.")
				return nil
			}

			// Update config with selected theme
			cfg.Display.Theme = selectedTheme.Name
			if err := cfgMgr.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Theme '%s' applied successfully!\n", selectedTheme.Name)
			fmt.Println("Run 'ghissues' to see the new theme in action.")

			return nil
		},
	}

	return cmd
}


// showThemePreview displays a more detailed theme preview
func showThemePreview(theme tui.Theme) {
	fmt.Println("\nTheme Preview:")
	fmt.Println("═══════════════════════════════════════════════════")

	// Create color blocks using ASCII art
	colorBlock := func(label string) string {
		return fmt.Sprintf("  %-15s ████████████████████", label)
	}

	fmt.Println(colorBlock("Primary Text"))
	fmt.Println(colorBlock("Secondary Text"))
	fmt.Println(colorBlock("Highlight"))
	fmt.Println(colorBlock("Success"))
	fmt.Println(colorBlock("Warning"))
	fmt.Println(colorBlock("Error"))
	fmt.Println(colorBlock("Info"))

	fmt.Println("\nThis theme will affect:")
	fmt.Println("  • Issue list headers and selection")
	fmt.Println("  • Issue detail titles and labels")
	fmt.Println("  • Comment author names and dates")
	fmt.Println("  • Status messages and error displays")
	fmt.Println("  • Help overlay and footer text")
	fmt.Println("═══════════════════════════════════════════════════")
}