# ghissues

A terminal-based user interface (TUI) for browsing GitHub issues locally. `ghissues` syncs issues from GitHub to a local SQLite database and provides a fast, keyboard-driven interface for browsing and searching issues offline.

![Version](https://img.shields.io/badge/version-0.1.0-blue)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)

## Features

- **Offline Browsing** - View issues and comments without an internet connection
- **Multi-Repository Support** - Configure and switch between multiple repositories
- **Vim-like Navigation** - Intuitive keyboard shortcuts for efficient navigation
- **Color Themes** - Six built-in themes (default, dracula, gruvbox, nord, solarized-dark, solarized-light)
- **Incremental Sync** - Only fetches updated issues to minimize API usage
- **Flexible Sorting** - Sort by number, created date, updated date, or comment count
- **Markdown Rendering** - Toggle between raw and rendered markdown views
- **Full-Text Search** - Search through issue comments

## Installation

### Prerequisites

- Go 1.25 or later
- A GitHub personal access token (create one at https://github.com/settings/tokens)

### From Source

```bash
git clone https://github.com/shepbook/ghissues.git
cd ghissues
go build -o ghissues ./cmd/ghissues
sudo install ghissues /usr/local/bin/
```

## Quick Start

1. Run `ghissues` to start the interactive setup wizard
2. Enter your repository in `owner/repo` format (e.g., `golang/go`)
3. Choose your authentication method (GitHub CLI token, environment variable, or stored token)
4. Issues will sync automatically on first run
5. Browse issues using the keyboard shortcuts below

## Usage

```
ghissues [flags] [command]
```

### Flags

| Flag | Description |
|------|-------------|
| `--config string` | Path to config file (default: `~/.config/ghissues/config.toml`) |
| `--db string` | Path to database file (default: `.ghissues.db`) |
| `--repo string` | Repository to use (owner/repo format) |
| `--help` | Show help information |
| `--version` | Show version information |

### Commands

| Command | Description |
|---------|-------------|
| `config` | Run interactive configuration wizard |
| `themes` | List available color themes and show current theme |
| `repos` | List configured repositories |
| `sync` | Sync issues from GitHub (runs automatically on first use) |

## Keyboard Shortcuts

### Main View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Open comments view |
| `Space` | Select current issue |
| `s` | Cycle sort field |
| `S` | Toggle sort order (ascending/descending) |
| `m` | Toggle markdown rendering |
| `r` | Incremental refresh (fetch updated issues) |
| `R` | Full refresh (re-fetch all issues) |
| `?` | Show help overlay |
| `q` / `Ctrl+C` | Quit |

### Comments View

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `m` | Toggle markdown rendering |
| `Esc` / `q` | Back to issue list |

## Configuration

The configuration file is stored at `~/.config/ghissues/config.toml`. Run `ghissues config` to reconfigure at any time.

### Example Configuration

```toml
[default]
repository = "owner/repository"

[[repositories]]
owner = "golang"
name = "go"
database = ".ghissues-golang-go.db"

[[repositories]]
owner = "rust-lang"
name = "rust"
database = ".ghissues-rust-lang-rust.db"

[display]
theme = "dracula"
columns = ["number", "title", "author", "updated", "comments"]

[sort]
field = "updated"
descending = true

[database]
path = ".ghissues.db"
```

### Configuration Options

- **`[default]`** - Default repository settings
  - `repository` - Default repository in `owner/repo` format

- **`[[repositories]]`** - Additional repositories (array)
  - `owner` - Repository owner/organization
  - `name` - Repository name
  - `database` - Database filename for this repository

- **`[display]`** - Display settings
  - `theme` - Color theme (default, dracula, gruvbox, nord, solarized-dark, solarized-light)
  - `columns` - Columns to display (number, title, author, updated, comments)

- **`[sort]`** - Sorting options
  - `field` - Sort field (number, created, updated, comments)
  - `descending` - Sort order (true for descending)

- **`[database]`** - Database settings
  - `path` - Default database path

## Authentication

`ghissues` supports multiple authentication methods. The token is used in the following order:

1. **Config file token** - Stored in `~/.config/ghissues/auth.toml` (set up via `ghissues config`)
2. **GitHub CLI** - Automatically uses `gh` auth token if available
3. **Environment variable** - `GITHUB_TOKEN` environment variable

The auth file is protected with file permissions `0600` for security.

### Creating a GitHub Token

1. Go to https://github.com/settings/tokens
2. Click "Generate new token" (classic)
3. Select scopes: `public_repo` (for public repositories) or `repo` (for private)
4. Generate and copy the token

## Themes

The following color themes are available:

| Theme | Description |
|-------|-------------|
| `default` | Clean blue and cyan theme |
| `dracula` | Dark purple-based theme |
| `gruvbox` | Warm orange and brown theme |
| `nord` | Cool blue and gray theme |
| `solarized-dark` | Solarized dark variant |
| `solarized-light` | Solarized light variant |

Run `ghissues themes` to see all available themes and your current theme. Change themes by editing `~/.config/ghissues/config.toml`.

## Multi-Repository Setup

To work with multiple repositories:

1. Run `ghissues config` to add repositories
2. List configured repositories with `ghissues repos`
3. Select a specific repository when running:
   ```bash
   ghissues --repo owner/repo
   ```

Each repository uses its own SQLite database file, keeping issues separate.

## Development

### Project Structure

```
github-issues-tui/
├── cmd/ghissues/          # Main entry point
├── internal/
│   ├── auth/              # Authentication and token management
│   ├── config/            # Configuration file handling
│   ├── database/          # Database path management
│   ├── github/            # GitHub API client
│   ├── setup/             # Interactive setup wizard
│   ├── sort/              # Issue sorting logic
│   ├── storage/           # SQLite database operations
│   ├── sync/              # GitHub sync operations
│   ├── theme/             # Color theme definitions
│   ├── timefmt/           # Time formatting utilities
│   └── tui/               # Terminal UI (Bubble Tea)
├── tasks/                 # Product requirements
├── go.mod
└── go.sum
```

### Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [modernc.org/sqlite](https://gitlab.com/cznic/sqlite) - Embedded SQLite

### Building

```bash
go build -o ghissues ./cmd/ghissues
```

### Running Tests

```bash
go test ./...
```

## License

MIT License - see LICENSE file for details

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Author

Created by [@shepbook](https://github.com/shepbook)
