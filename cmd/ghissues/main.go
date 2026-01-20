package main

import (
	"fmt"
	"os"

	"github.com/shepbook/ghissues/internal/config"
)

func main() {
	// Check for 'config' subcommand
	if len(os.Args) > 1 && os.Args[1] == "config" {
		if err := runSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if config exists, if not run setup
	if !config.Exists() {
		fmt.Println("Welcome to ghissues! Let's set up your configuration.")
		fmt.Println()

		if err := runSetup(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Configuration complete. Run 'ghissues' to start the app.")
}
