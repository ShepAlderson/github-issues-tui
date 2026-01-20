package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// RunInteractiveSetup runs the full interactive setup process
func RunInteractiveSetup(input io.Reader, output io.Writer) (repository string, token string, err error) {
	// Create a single reader that will be reused
	reader := bufio.NewReader(input)

	// Prompt for repository
	repository, err = promptRepository(reader, output)
	if err != nil {
		return "", "", fmt.Errorf("failed to prompt for repository: %w", err)
	}

	// Get existing token from environment if available
	existingToken := getEnvToken()

	// Prompt for token selection
	method, err := promptTokenSelection(reader, output, existingToken)
	if err != nil {
		return "", "", fmt.Errorf("failed to prompt for token: %w", err)
	}

	// Get token based on selection
	switch method {
	case "env":
		token = existingToken
		if token == "" {
			return "", "", fmt.Errorf("no token found in GITHUB_TOKEN environment variable")
		}
	case "config":
		// Will be stored in config file
		token = ""
	case "manual":
		token, err = promptManualToken(reader, output)
		if err != nil {
			return "", "", fmt.Errorf("failed to prompt for manual token: %w", err)
		}
	default:
		return "", "", fmt.Errorf("invalid token selection: %s", method)
	}

	return repository, token, nil
}

// promptRepository prompts the user for a GitHub repository
func promptRepository(reader *bufio.Reader, output io.Writer) (string, error) {
	for {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Enter GitHub repository (owner/repo format, e.g., anthropic/claude):")
		fmt.Fprint(output, "> ")

		repo, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		repo = strings.TrimSpace(repo)
		if repo == "" {
			fmt.Fprintln(output, "Repository cannot be empty. Please try again.")
			continue
		}

		// Validate format: should contain exactly one slash
		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			fmt.Fprintln(output, "Invalid format. Use 'owner/repo' format (e.g., anthropic/claude).")
			continue
		}

		owner := strings.TrimSpace(parts[0])
		repoName := strings.TrimSpace(parts[1])

		if owner == "" || repoName == "" {
			fmt.Fprintln(output, "Both owner and repository name are required.")
			continue
		}

		return owner + "/" + repoName, nil
	}
}

// promptTokenSelection prompts the user to select how to provide their token
func promptTokenSelection(reader *bufio.Reader, output io.Writer, existingToken string) (string, error) {
	for {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Select authentication method:")

		opCount := 0
		if existingToken != "" {
			opCount++
			fmt.Fprintf(output, "%d. Use token from GITHUB_TOKEN environment variable\n", opCount)
		}
		opCount++
		fmt.Fprintf(output, "%d. Use gh CLI token (will check ~/.config/gh/hosts.yml)\n", opCount)
		opCount++
		fmt.Fprintf(output, "%d. Enter GitHub personal access token manually\n", opCount)
		fmt.Fprint(output, "> ")

		choice, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		choice = strings.TrimSpace(choice)

		// Map choice to selection based on what's available
		if choice == "1" {
			if existingToken != "" {
				return "env", nil
			}
			return "config", nil
		} else if choice == "2" {
			if existingToken != "" {
				return "config", nil
			}
			return "manual", nil
		} else if choice == "3" && existingToken != "" {
			return "manual", nil
		} else {
			fmt.Fprintln(output, "Invalid selection. Please try again.")
		}
	}
}

// promptManualToken prompts the user to enter a token manually
func promptManualToken(reader *bufio.Reader, output io.Writer) (string, error) {
	for {
		fmt.Fprintln(output)
		fmt.Fprintln(output, "Enter GitHub Personal Access Token:")
		fmt.Fprintln(output, "You can create one at https://github.com/settings/tokens")
		fmt.Fprint(output, "> ")

		token, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		token = strings.TrimSpace(token)
		if token == "" {
			fmt.Fprintln(output, "Token cannot be empty.")
			continue
		}

		return token, nil
	}
}

// getEnvToken returns the GITHUB_TOKEN from environment if set
func getEnvToken() string {
	return os.Getenv("GITHUB_TOKEN")
}
