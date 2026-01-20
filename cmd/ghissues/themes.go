package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/themes"
)

// runThemes displays available themes and allows the user to select one
func runThemes() error {
	// Load existing config if it exists
	var cfg *config.Config
	var err error
	if config.Exists() {
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	} else {
		cfg = &config.Config{
			Database: config.Database{},
			Display:  config.Display{},
		}
	}

	// Display available themes
	fmt.Println("Available Themes:")
	fmt.Println()
	fmt.Println(themes.List())
	fmt.Println()

	// Get current theme
	currentTheme := cfg.Display.Theme
	if currentTheme == "" {
		currentTheme = config.DefaultTheme()
	}

	// Prompt for theme selection
	fmt.Printf("Current theme: %s\n\n", currentTheme)
	fmt.Println("Enter a theme name to preview/change (or press Enter to keep current): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	input = strings.TrimSpace(input)

	if input != "" {
		// Validate theme name
		if !themes.IsValid(input) {
			return fmt.Errorf("invalid theme: %s", input)
		}

		// Update config
		cfg.Display.Theme = input

		// Save config
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("\nTheme set to: %s\n", input)
	} else {
		fmt.Printf("\nTheme unchanged: %s\n", currentTheme)
	}

	return nil
}

// GetCurrentTheme returns the current theme from config or default
func GetCurrentTheme() themes.Theme {
	cfg, err := config.Load()
	if err != nil {
		return themes.Default
	}

	themeName := cfg.Display.Theme
	if themeName == "" {
		themeName = config.DefaultTheme()
	}

	return themes.Get(themeName)
}