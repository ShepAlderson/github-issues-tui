# ghissues - GitHub Issues TUI

A terminal-based user interface for reviewing GitHub issues offline. Fetch issues from GitHub, store them locally in a SQLite database, and browse them with an interactive, keyboard-driven UI.

![TUI Screenshot Placeholder]

## Features

- **Offline Access**: Issues are stored locally after syncing
- **Keyboard Navigation**: Fully keyboard-driven interface
- **Markdown Rendering**: Issues and comments rendered with GitHub-flavored markdown
- **Multiple Repositories**: Configure and switch between multiple repos
- **Color Themes**: 6 built-in themes (default, dracula, gruvbox, nord, solarized-dark, solarized-light)
- **Incremental Sync**: Only fetches updated issues on refresh

## Installation

### Prerequisites

- Go 1.25.5 or later
- GitHub Personal Access Token (with `repo` scope for private repositories)

### Building

```bash
git clone https://github.com/yourusername/github-issues-tui.git
cd github-issues-tui
go build -o ghissues ./cmd/ghissues
```

## Configuration

### First-Time Setup

Run the interactive configuration wizard:

```bash
./ghissues config
```

This will prompt you for:
- GitHub Personal Access Token (or use `GITHUB_TOKEN` environment variable)
- Repository owner/name (e.g., `owner/repo`)
- Display preferences (theme, columns, sorting)

### Manual Configuration

The config file is stored at `~/.config/ghissues/config.toml`:

```toml
repositories = ["owner/repo1", "owner/repo2"]
default_repository = "owner/repo1"

[auth]
token = "ghp_xxxxxxxxxxxx"  # Optional; also reads GITHUB_TOKEN env

[display]
theme = "dracula"
columns = ["number", "title", "author", "date", "comments"]
sort = "updated"
sort_order = "desc"

[database]
path = ".ghissues.db"
```

### Authentication

Authentication is checked in this order:
1. `GITHUB_TOKEN` environment variable
2. `token` in config file
3. GitHub CLI token (`gh auth token`)

Create a token at: https://github.com/settings/tokens (needs `repo` scope)

## Usage

### Commands

| Command | Description |
|---------|-------------|
| `./ghissues config` | Run interactive configuration setup |
| `./ghissues sync` | Manually sync issues from GitHub |
| `./ghissues tui` | Launch the TUI interface |
| `./ghissues themes` | Preview available color themes |
| `./ghissues repos` | List configured repositories |
| `./ghissues` | Run with defaults (TUI + auto-refresh) |

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--db PATH` | Database file path | `.ghissues.db` |
| `--repo OWNER/REPO` | Repository to use | From config |

Example:
```bash
./ghissues --db /path/to/db.sqlite --repo owner/repo
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `j` / `Down` | Move down in issue list |
| `k` / `Up` | Move up in issue list |
| `Enter` | Open issue comments |
| `r` | Refresh issues from GitHub |
| `m` | Toggle markdown rendering |
| `?` | Toggle help overlay |
| `q` | Go back / Quit |
| `Ctrl+C` | Force quit |

### Sorting and Filtering

Press `s` to cycle through sorting options:
- By updated date
- By created date
- By issue number
- By comment count

Toggle sort order with `o`.

## Project Structure

```
├── cmd/ghissues/          # CLI commands and TUI
│   ├── main.go            # Entry point
│   ├── config.go          # Setup wizard
│   ├── sync.go            # GitHub sync logic
│   ├── tui.go             # Main TUI application
│   └── themes.go          # Theme preview
├── internal/
│   ├── config/            # Configuration management
│   ├── auth/              # GitHub authentication
│   ├── github/            # GitHub API client
│   ├── db/                # SQLite database layer
│   ├── themes/            # Color theme definitions
│   └── errors/            # Error handling
```

## Development

### Running Tests

```bash
go test ./...
```

### Adding a New Theme

Themes are defined in `internal/themes/themes.go`:

```go
var themes = []Theme{
    {Name: "mytheme", ...},
}
```

## License

MIT