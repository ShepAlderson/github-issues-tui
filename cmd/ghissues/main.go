package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/shepbook/ghissues/internal/config"
	"github.com/shepbook/ghissues/internal/setup"
)

const version = "0.1.0"

func main() {
	// Define flags
	var (
		showVersion bool
		showHelp    bool
		configPath  string
		repoFlag    string
	)

	flag.BoolVar(&showVersion, "version", false, "show version information")
	flag.BoolVar(&showHelp, "help", false, "show help information")
	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.StringVar(&repoFlag, "repo", "", "repository to use (owner/repo)")

	flag.Parse()

	// Handle version flag
	if showVersion {
		fmt.Printf("ghissues version %s\n", version)
		os.Exit(0)
	}

	// Handle help flag
	if showHelp {
		printHelp()
		os.Exit(0)
	}

	// Determine config path
	if configPath == "" {
		configPath = config.GetDefaultConfigPath()
	}

	// Get command arguments
	args := flag.Args()
	command := ""
	if len(args) > 0 {
		command = args[0]
	}

	// Handle config command
	if command == "config" {
		handleConfigCommand(configPath)
		return
	}

	// Check if config exists
	if !config.ConfigExists(configPath) {
		// Run first-time setup
		prompter := setup.NewPrompter()
		err := prompter.RunSetup(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		fmt.Fprintln(os.Stderr, "Run 'ghissues config' to reconfigure.")
		os.Exit(1)
	}

	// TODO: Implement main TUI application
	fmt.Printf("Configuration loaded for repository: %s\n", cfg.Default.Repository)
	fmt.Println("TUI application not yet implemented.")
}

func handleConfigCommand(configPath string) {
	prompter := setup.NewPrompter()
	err := prompter.RunSetup(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration failed: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Printf("ghissues - A terminal-based GitHub issues viewer\n\n")
	fmt.Printf("Usage:\n")
	fmt.Printf("  ghissues [flags]\n")
	fmt.Printf("  ghissues config\n\n")
	fmt.Printf("Flags:\n")
	fmt.Printf("  --config string   Path to config file (default: ~/.config/ghissues/config.toml)\n")
	fmt.Printf("  --repo string     Repository to use (owner/repo)\n")
	fmt.Printf("  --help            Show help information\n")
	fmt.Printf("  --version         Show version information\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  config            Run interactive configuration\n")
}
