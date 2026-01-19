# ghissues - GitHub Issues TUI

A terminal user interface (TUI) for reviewing GitHub issues offline.

## Features

- 🚀 Interactive first-time setup
- 🔐 Multiple authentication methods (env var, token, gh CLI)
- 💾 Secure config storage with proper file permissions
- 📝 Easy reconfiguration with `ghissues config`

## Installation

```bash
go install github.com/shepbook/github-issues-tui/cmd/ghissues@latest
```

Or build from source:

```bash
git clone https://github.com/shepbook/github-issues-tui.git
cd github-issues-tui
go build -o ghissues ./cmd/ghissues
```

## Quick Start

1. Run `ghissues` - you'll be prompted to configure on first run
2. Enter your GitHub repository in `owner/repo` format
3. Choose an authentication method
4. Start reviewing issues!

## Usage

```bash
# Start the TUI (runs setup if needed)
ghissues

# Run/re-run interactive configuration
ghissues config

# Show help
ghissues --help

# Show version
ghissues --version
```

## Configuration

Configuration is stored in `~/.config/ghissues/config.toml`.

### Authentication Methods

- **env**: Use `GITHUB_TOKEN` environment variable
- **token**: Store token in config file (secure 0600 permissions)
- **gh**: Use GitHub CLI (gh) authentication

### Manual Configuration

You can also manually edit the config file:

```toml
[github]
repository = "owner/repo"
auth_method = "env"  # or "token" or "gh"
token = "ghp_..."    # only if auth_method is "token"
```

## Development

### Running Tests

```bash
go test -v ./...
```

### Building

```bash
go build -o ghissues ./cmd/ghissues
```

## License

MIT License - see LICENSE file for details
