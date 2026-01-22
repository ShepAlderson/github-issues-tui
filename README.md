# ghissues

A terminal user interface (TUI) for browsing GitHub issues offline. Sync issues from any GitHub repository to a local database and browse them with a fast, keyboard-driven interface.

## Features

- **Offline browsing** - Sync issues once, browse anytime without internet
- **Interactive setup** - First-run wizard configures repository and authentication
- **Multiple auth methods** - Environment variable, config file, or GitHub CLI
- **Split-panel interface** - Issue list with detail view side-by-side
- **Markdown rendering** - View issue bodies and comments with full markdown support
- **Comments drill-down** - Browse all comments in a full-screen view
- **Configurable display** - Choose which columns to show and how to sort
- **6 color themes** - Default, Dracula, Gruvbox, Nord, Solarized Dark/Light
- **Multi-repository support** - Configure and switch between multiple repos
- **Auto-refresh** - Syncs on startup with manual refresh available

## Prerequisites

- Go 1.21 or later
- A GitHub account with access to the repository you want to browse
- One of the following for authentication:
  - `GITHUB_TOKEN` environment variable
  - GitHub CLI (`gh`) installed and authenticated
  - A personal access token

## Installation

### From Source

```bash
git clone https://github.com/your-username/github-issues-tui.git
cd github-issues-tui
go build -o ghissues ./cmd/ghissues
```

### Install to GOBIN

```bash
go install ./cmd/ghissues
```

This installs the binary to your `$GOBIN` directory (usually `~/go/bin`).

## Quick Start

1. **Run the application:**
   ```bash
   ghissues
   ```

2. **Complete the setup wizard:**
   - Enter the repository in `owner/repo` format
   - Choose your authentication method
   - Optionally specify a custom database path

3. **Browse issues:**
   - Use `j/k` or arrow keys to navigate
   - Press `Enter` to view comments
   - Press `?` for help

## Usage

### Commands

```bash
ghissues              # Launch TUI (runs setup on first use)
ghissues sync         # Sync issues without launching TUI
ghissues config       # Re-run the setup wizard
ghissues themes       # List available themes
ghissues repos        # List configured repositories
```

### Command Flags

```bash
ghissues --repo owner/repo    # Use a specific repository
ghissues --db /path/to.db     # Use a specific database file
```

### Theme Management

```bash
ghissues themes                    # List all available themes
ghissues themes --preview dracula  # Preview a theme
ghissues themes --set dracula      # Set the active theme
```

### Repository Management

```bash
ghissues repos                           # List configured repos
ghissues repos --add owner/repo          # Add a new repository
ghissues repos --set-default owner/repo  # Set the default repository
```

## Keybindings

### Issue List View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | View comments |
| `r` / `R` | Refresh issues |
| `s` | Cycle sort field |
| `S` | Toggle sort order |
| `m` | Toggle markdown rendering |
| `h` / `l` | Scroll detail panel |
| `?` | Show help |
| `q` / `Ctrl+C` | Quit |

### Comments View

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `Esc` / `q` | Return to list |

## Configuration

Configuration is stored in `~/.config/ghissues/config.toml`.

### Full Configuration Example

```toml
# Default repository to use
default_repository = "owner/repo"

# Multiple repositories can be configured
[[repositories]]
  name = "owner/repo"
  database_path = ".owner-repo.db"

[[repositories]]
  name = "another/repo"
  database_path = ".another-repo.db"

# Authentication settings
[auth]
  method = "env"    # "env", "token", or "gh"
  # token = "ghp_xxx..."  # Only needed if method = "token"

# Database settings
[database]
  path = ".ghissues.db"  # Default database path

# Display preferences
[display]
  columns = ["number", "title", "author", "date", "comments"]
  sort_field = "updated"  # "updated", "created", "number", "comments"
  sort_order = "desc"     # "desc" or "asc"
  theme = "default"
```

### Authentication Methods

Authentication is resolved in priority order:

1. **Environment variable** (`method = "env"`)
   - Uses the `GITHUB_TOKEN` environment variable
   - Recommended for security

2. **Config file token** (`method = "token"`)
   - Stores token directly in config file
   - Set `token = "ghp_..."` in the `[auth]` section

3. **GitHub CLI** (`method = "gh"`)
   - Uses `gh auth token` to retrieve the token
   - Requires GitHub CLI to be installed and authenticated

### Display Columns

Available columns for the issue list:
- `number` - Issue number
- `title` - Issue title
- `author` - Issue author username
- `date` - Created date
- `comments` - Comment count

### Themes

Available themes:
- `default` - Clean terminal colors
- `dracula` - Dark purple-based theme
- `gruvbox` - Warm retro colors
- `nord` - Arctic blue colors
- `solarized-dark` - Solarized dark theme
- `solarized-light` - Solarized light theme

## Database

Issues are stored in a local SQLite-compatible database (LibSQL). By default, the database is created at `.ghissues.db` in the current directory, but this can be configured.

### Database Location Priority

1. `--db` command-line flag
2. Repository-specific `database_path` in config
3. `[database].path` in config
4. Default: `.ghissues.db`

### Data Storage

The database contains:
- **issues** - Issue metadata (number, title, author, body, dates, etc.)
- **comments** - Issue comments with author and body
- **metadata** - Sync timestamps and other metadata

Only open issues are synced. When issues are closed or deleted on GitHub, they are removed from the local database on the next sync.

## Development

### Building

```bash
go build -o ghissues ./cmd/ghissues
```

### Running Tests

```bash
go test ./...
```

### Project Structure

```
├── cmd/ghissues/         # Application entry point
├── internal/
│   ├── auth/             # GitHub authentication
│   ├── cmd/              # CLI commands (cobra)
│   ├── config/           # Configuration loading/saving
│   ├── db/               # Database operations
│   ├── github/           # GitHub API client
│   ├── setup/            # Interactive setup wizard
│   ├── sync/             # Issue synchronization
│   ├── themes/           # Color theme definitions
│   └── tui/              # Terminal UI (bubbletea)
├── go.mod
└── go.sum
```

### Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [huh](https://github.com/charmbracelet/huh) - Interactive forms
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [go-libsql](https://github.com/tursodatabase/go-libsql) - SQLite-compatible database

## License

MIT
