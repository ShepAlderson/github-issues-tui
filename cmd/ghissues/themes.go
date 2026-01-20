package main

import (
	"fmt"
	"os"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/theme"
)

func handleThemesCommand(configPath string) {
	// Load config to get current theme
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	currentTheme := cfg.Display.Theme
	if currentTheme == "" {
		currentTheme = "default"
	}

	// Display current theme
	fmt.Printf("Current theme: %s\n\n", currentTheme)

	// List all available themes
	themes := theme.GetAllThemes()
	fmt.Println("Available themes:")
	fmt.Println()

	for i, t := range themes {
		prefix := "  "
		if t == currentTheme {
			prefix = "* "
		}
		fmt.Printf("%s%d. %s\n", prefix, i+1, t)
	}

	fmt.Println()
	fmt.Println("To change theme, edit your config file:")
	fmt.Printf("  %s\n", configPath)
	fmt.Println()
	fmt.Println("Add or update the [display] section:")
	fmt.Println("  [display]")
	fmt.Println("  theme = \"dracula\"  # or: gruvbox, nord, solarized-dark, solarized-light")
}
