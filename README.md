# GitHub Issues TUI (ghissues)

A terminal-based user interface (TUI) application for reviewing GitHub issues offline. The application fetches open issues from GitHub repositories via the GraphQL API, stores them in a local SQLite database, and provides an interactive interface for browsing issues and their comments. All 14 user stories from the Product Requirements Document have been implemented, providing a complete MVP for offline GitHub issue review.

## Features

- **Offline Issue Review**: Download and store GitHub issues locally for offline access
- **Interactive TUI**: Keyboard-driven terminal interface using Charmbracelet's BubbleTea framework
- **Multiple Authentication Methods**: Supports GITHUB_TOKEN environment variable, config file tokens, and GitHub CLI authentication
- **Configurable Display**: Customizable columns, sorting, and color themes
- **Markdown Rendering**: Issue and comment bodies rendered with markdown support
- **Incremental Sync**: Only fetches updated data since last sync
- **Error Handling**: Clear error messages with actionable guidance
- **Multi-Repository Support**: Configure and switch between multiple GitHub repositories

## Installation

### Prerequisites

- Go 1.25.5+ (for building from source)
- GitHub authentication (token or GitHub CLI)
- Terminal with UTF-8 support (Linux, macOS, Windows Terminal recommended)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/shepbook/github-issues-tui.git
cd github-issues-tui

# Build the application
go build -o ghissues ./cmd/ghissues

# Make it available in your PATH (optional)
sudo cp ghissues /usr/local/bin/
```

### Using the Test Script

```bash
./test_e2e.sh
```

## Quick Start

1. **Build the application**:
   ```bash
   git clone <repository-url>
   cd github-issues-tui
   go build -o ghissues ./cmd/ghissues
   ```

2. **Run the application** (first-time setup will prompt for configuration):
   ```bash
   ./ghissues
   ```

3. **Configure repository and authentication**:
   - Enter GitHub repository in `owner/repo` format
   - Choose authentication method (GitHub CLI token recommended)
   - Accept default database location or specify custom path

4. **Start browsing issues**:
   - Use `j`/`k` or arrow keys to navigate
   - Press `Enter` to view comments
   - Press `?` for help overlay

## Usage

### First-Time Setup

Run `ghissues` without existing configuration to start the interactive setup wizard:

```bash
./ghissues
```

The setup wizard will prompt for:
1. GitHub repository (owner/repo format)
2. Authentication method preference
3. Database location (default: `.ghissues.db` in current directory)

Configuration is saved to `~/.config/ghissues/config.toml`.

### Running the Application

```bash
# Launch the TUI with default configuration
ghissues

# Specify a custom database location
ghissues --db /path/to/database.db

# View available themes
ghissues themes

# Manually sync issues
ghissues sync
```

### Authentication Methods

The application supports three authentication methods in order of precedence:

1. **Environment Variable**: `GITHUB_TOKEN` environment variable
2. **Config File Token**: Token stored in config file (`~/.config/ghissues/config.toml`)
3. **GitHub CLI**: Uses `gh` CLI authentication (if installed and configured)

## Configuration

Configuration is stored in `~/.config/ghissues/config.toml`. Example configuration:

```toml
[auth]
# Token is optional if using gh CLI auth
token = "ghp_..."

[default]
repository = "owner/repo"

[[repositories]]
owner = "anthropics"
name = "claude-code"
database = ".ghissues-claude.db"

[[repositories]]
owner = "charmbracelet"
name = "bubbletea"
database = ".ghissues-bubbletea.db"

[display]
theme = "dracula"
columns = ["number", "title", "author", "updated", "comments"]

[sort]
field = "updated"
descending = true

[database]
path = ".ghissues.db"  # default, can be overridden per-repo
```

## Keybindings

### Global Keybindings
- `?` - Toggle help overlay
- `Ctrl+C` - Force quit

### Issue List View
- `j`/`↓` - Move down
- `k`/`↑` - Move up
- `g`/`Home` - Go to first issue
- `G`/`End` - Go to last issue
- `Enter` - Open comments view for selected issue
- `s` - Cycle sort field
- `S` - Toggle sort direction
- `r` - Refresh from GitHub
- `q` - Quit application

### Issue Detail View
- `j`/`↓` - Scroll down
- `k`/`↑` - Scroll up
- `g` - Go to top
- `G` - Go to bottom
- `m` - Toggle markdown rendering
- `Enter` - Open comments view
- `Esc` - Return to list view

### Comments View
- `j`/`↓` - Next comment
- `k`/`↑` - Previous comment
- `m` - Toggle markdown rendering
- `Esc`/`q` - Return to issue list view

### Error Modal
- `Enter`/`Space`/`Esc` - Dismiss error

## Color Themes

The application includes 6 built-in color themes:

1. **default** - Default terminal colors
2. **dracula** - Dark purple/cyan theme
3. **gruvbox** - Warm earthy colors
4. **nord** - Arctic blue/gray theme
5. **solarized-dark** - Low contrast dark theme
6. **solarized-light** - Low contrast light theme

Preview and select themes with:

```bash
ghissues themes
```

## Database

The application uses SQLite for local storage. The database schema includes:

- **issues**: Stores issue details (number, title, body, author, dates, labels, assignees)
- **comments**: Stores issue comments with foreign key to issues
- **metadata**: Stores sync timestamps and other metadata

Database location precedence:
1. Command line `--db` flag
2. `database.path` in config file
3. Default: `.ghissues.db` in current directory

## Project Structure

```
github-issues-tui/
├── cmd/
│   └── ghissues/
│       └── main.go              # CLI entry point
├── internal/
│   ├── cli/                     # CLI command definitions (Cobra)
│   ├── config/                  # Configuration management
│   ├── database/                # SQLite database operations
│   ├── github/                  # GitHub API integration
│   ├── setup/                   # First-time setup wizard
│   ├── sync/                    # Issue synchronization logic
│   └── tui/                     # Terminal user interface
├── go.mod                       # Go module dependencies
└── test_e2e.sh                  # End-to-end test script
```

## Development

### Dependencies

- **TUI Framework**: `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`
- **Markdown Rendering**: `github.com/charmbracelet/glamour`
- **GitHub API**: `github.com/google/go-github/v62`
- **CLI Framework**: `github.com/spf13/cobra`
- **Interactive Prompts**: `github.com/manifoldco/promptui`
- **Database**: `github.com/mattn/go-sqlite3`
- **Configuration**: `github.com/pelletier/go-toml/v2`
- **OAuth**: `golang.org/x/oauth2`

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

The project includes comprehensive tests:
- **Unit Tests**: Business logic in config, database, auth modules
- **Integration Tests**: Database operations with in-memory SQLite
- **Behavioral Tests**: TUI interaction flows using teatest
- **Golden File Tests**: TUI component rendering snapshots

### Code Patterns

The codebase follows consistent patterns documented in `.ralph-tui/progress.md`:

1. **Manager Pattern**: Dependency injection for testability
2. **TOML Configuration**: Secure file storage with 0600 permissions
3. **Interactive CLI Setup**: Validated prompts using promptui
4. **Component Composition**: Standalone Bubble Tea components
5. **Database Path Resolution**: Precedence rules with safety checks
6. **Column Configuration**: Configurable column rendering with fallback defaults
7. **Relative Time Formatting**: Human-readable relative time strings
8. **Theme Management**: Centralized color theme definitions

## Future Enhancements

Potential future features (post-MVP):
- Issue filtering and search
- Label/assignee modification
- Comment creation
- Issue state changes
- Offline change queuing
- Custom keybinding configuration
- Export to markdown/JSON

## Troubleshooting

### Authentication Issues
- **No valid authentication found**: Ensure you have either:
  - `GITHUB_TOKEN` environment variable set
  - GitHub CLI installed and authenticated (`gh auth login`)
  - Valid token in config file (`~/.config/ghissues/config.toml`)
- **Invalid token error**: Regenerate your GitHub token with appropriate permissions (repo scope)

### Database Issues
- **Database path not writable**: Check permissions on parent directories
- **Database corruption**: Delete the database file and run `ghissues sync` to re-fetch data

### Sync Issues
- **Network errors**: Check internet connectivity and GitHub API status
- **Rate limiting**: GitHub API has rate limits; wait before retrying
- **Repository not found**: Verify repository name is in `owner/repo` format

### TUI Issues
- **Application crashes on launch**: Ensure terminal supports UTF-8 and has sufficient dimensions
- **Keybindings not working**: Check terminal compatibility; some terminals may intercept keys

## License

[Add appropriate license]

## Contributing

[Add contribution guidelines]