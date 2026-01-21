# GitHub Issues TUI

A beautiful terminal user interface for managing GitHub issues, built with Go and Charmbracelet's Bubbletea framework.

## Features

- **Interactive TUI**: Navigate issues with keyboard shortcuts in a rich terminal interface
- **Multi-Repository Support**: Configure and switch between multiple GitHub repositories
- **Local Caching**: Sync issues to a local SQLite database for offline access and fast performance
- **Rich Issue Viewing**: View issue details, comments, labels, and metadata
- **Color Themes**: Built-in dark and light themes with customizable colors
- **Incremental Sync**: Automatically syncs new changes when launching the TUI
- **Progress Tracking**: Visual progress bars during sync operations
- **Error Handling**: User-friendly error messages with modal dialogs
- **Search & Navigation**: Quick filtering and navigation through issues

## Installation

### Prerequisites

- Go 1.25.5 or later
- GitHub Personal Access Token (PAT) with `repo` scope

### Build from Source

```bash
# Clone the repository
git clone https://github.com/shepbook/git/github-issues-tui.git
cd github-issues-tui

# Build the binary
go build -o ghissues

# Install to $GOPATH/bin
go install
```

### Pre-built Binary

You can also download pre-built binaries from the releases page.

## Quick Start

1. **Initial Setup**: Run without arguments for first-time configuration
   ```bash
   ghissues
   ```

2. **Configure Repository**: Interactive setup for GitHub repository and authentication
   ```bash
   ghissues config
   ```

3. **Sync Issues**: Download issues from GitHub to local database
   ```bash
   ghissues sync
   ```

4. **Launch TUI**: Open the interactive terminal interface (default command)
   ```bash
   ghissues list
   # or simply:
   ghissues
   ```

## Usage

### Commands

- `ghissues` or `ghissues list` - Launch the TUI (default)
- `ghissues config` - Configure repository and authentication
- `ghissues sync` - Sync issues from GitHub to local database
- `ghissues repos` - List configured repositories
- `ghissues --help` - Show help information

### Flags

- `--db <path>` - Override database path
- `--repo <owner/repo>` - Specify repository for multi-repo setups

### Environment Variables

- `GHISSUES_CONFIG` - Override config file path
- `GHISSUES_GITHUB_URL` - Override GitHub API URL (for testing)

## Configuration

### Config File Location

Configuration is stored in TOML format at:
- Linux/macOS: `~/.config/ghissues/config.toml`
- Windows: `%APPDATA%\ghissues\config.toml`

### Multi-Repository Setup

```toml
default_repo = "owner/repo1"

[repositories]
  [repositories.owner_repo1]
    database_path = "/path/to/repo1.db"

  [repositories.owner_repo2]
    database_path = "/path/to/repo2.db"

[display]
  theme = "dark"  # or "light"
  columns = ["number", "title", "state", "created_at", "updated_at"]

  [display.sort]
    field = "updated_at"
    descending = true
```

### Single Repository (Backward Compatible)

```toml
repository = "owner/repo"
token = "ghp_your_token_here"

[database]
  path = ".ghissues.db"

[display]
  theme = "dark"
  columns = ["number", "title", "state", "created_at"]

  [display.sort]
    field = "updated_at"
    descending = true
```

## TUI Keybindings

### Global
- `Tab` / `Shift+Tab` - Navigate between panels
- `q` / `Ctrl+c` - Quit application
- `?` - Show help dialog

### Issue List
- `↑` / `↓` or `k` / `j` - Navigate issues
- `Enter` - View issue details
- `r` - Refresh/sync issues
- `/` - Search/filter issues
- `1-9` - Quick filter by number

### Issue Detail View
- `←` / `→` or `h` / `l` - Switch between tabs (Details/Comments)
- `c` - View comments
- `b` - Open issue in browser
- `Esc` - Return to list view

### Modal Dialogs
- `Enter` or `Esc` - Close modal/acknowledge error

## Development

### Project Structure

```
.
├── main.go                 # CLI entry point
├── main_test.go           # Integration tests
├── go.mod                 # Go module definition
├── internal/
│   ├── auth/              # Authentication logic
│   ├── cmd/               # Command implementations
│   ├── config/            # Configuration management
│   ├── db/                # Database operations (SQLite)
│   ├── github/            # GitHub API client
│   ├── prompt/            # Interactive prompts
│   └── tui/               # Terminal UI components
│       ├── app.go         # Main TUI application
│       ├── list.go        # Issue list view
│       ├── detail.go      # Issue detail view
│       ├── comments.go    # Comments view
│       ├── help.go        # Help dialog
│       ├── modal.go       # Modal dialogs
│       └── errors.go      # Error handling
├── .ralph-tui/            # Ralph orchestration files
└── tasks/                 # Task definitions
```

### Building

```bash
# Build binary
go build -o ghissues

# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Build with race detection
go build -race -o ghissues
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/tui/...

# Run with verbose output
go test -v ./...

# Run integration tests with test mode
GHISSIES_TEST=1 go test
```

### Test Mode

Set `GHISSUES_TEST=1` to bypass TTY requirements during testing.

### Database Schema

Issues are cached in SQLite with the following fields:
- `number` - Issue number
- `title` - Issue title
- `state` - Issue state (open/closed)
- `user` - Issue author
- `assignees` - Comma-separated assignees
- `labels` - Comma-separated labels
- `milestone` - Milestone name
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp
- `closed_at` - Closure timestamp (if closed)
- `body` - Issue body/description
- `comments` - Comment count
- `url` - GitHub URL
- `repository` - Repository name

## Authentication

### GitHub Token Setup

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Generate a new token (classic) with `repo` scope
3. Use the token during `ghissues config` setup

### Token Security

- Tokens are stored in the config file - ensure proper file permissions
- Use environment variables for CI/CD environments
- Never commit tokens to version control
- Consider using GitHub CLI's credential storage as an alternative

## Troubleshooting

### Common Issues

**"Configuration not found"**
- Run `ghissues config` to set up initial configuration

**"Token validation failed"**
- Check token has `repo` scope
- Verify token hasn't expired
- Ensure network connectivity to GitHub

**"Database path is not writable"**
- Check directory permissions
- Verify sufficient disk space
- Try specifying a different path with `--db` flag

**Sync issues**
- Check GitHub API rate limits
- Verify repository exists and you have access
- Try manual sync with `ghissues sync`

### Debug Mode

Set `GHISSUES_DEBUG=1` for verbose logging (coming soon).

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests following TDD approach
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with:
- [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Charmbracelet Bubbles](https://github.com/charmbracelet/bubbles) - UI components
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) - SQLite driver
- [pelletier/go-toml](https://github.com/pelletier/go-toml) - TOML configuration

## Changelog

### v1.0.0
- Initial release
- Multi-repository support
- Interactive TUI with list/detail views
- Color themes (dark/light)
- Local caching with SQLite
- Incremental sync
- Error handling with modals
- Progress tracking
- Keybinding help system