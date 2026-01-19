# Product Requirements Document: ghissues

## Overview
A terminal-based user interface (TUI) application for reviewing GitHub issues offline. The app fetches open issues from a GitHub repository via the GraphQL API, stores them in a local libsql database, and provides an interactive interface for browsing issues and their comments.

## Goals
- Enable offline review of GitHub issues with a responsive, keyboard-driven interface
- Provide a clean, configurable TUI experience using charmbracelet tooling
- Support flexible configuration for different workflows and preferences
- Build with strict test-driven development practices

## Non-Goals (MVP)
- Modifying issues (labels, assignees, status)
- Creating new issues or comments
- Filtering/searching issues
- Closed issues view
- Conflict handling for offline edits

---

## User Stories

### US-1: First-Time Setup
**As a** user running ghissues for the first time  
**I want to** be prompted to configure the repository and authentication  
**So that** I can quickly get started without manually editing config files

**Acceptance Criteria:**
- Interactive prompt asks for GitHub repository (owner/repo format)
- Interactive prompt asks for authentication method preference
- Configuration is saved to `~/.config/ghissues/config.toml`
- User can skip interactive setup if config file already exists
- User can re-run setup with `ghissues config` command

---

### US-2: GitHub Authentication
**As a** user  
**I want to** authenticate with GitHub using multiple methods  
**So that** I can use whichever method fits my workflow

**Acceptance Criteria:**
- Authentication attempts in order: environment variable (`GITHUB_TOKEN`) → config file → `gh` CLI token
- Clear error message if no valid authentication found
- Token is validated on first API call with helpful error if invalid
- Config file token is stored securely (file permissions 0600)

---

### US-3: Initial Issue Sync
**As a** user  
**I want to** fetch all open issues from my configured repository  
**So that** I can review them offline

**Acceptance Criteria:**
- Progress bar displays during fetch showing issues fetched / total
- All open issues are fetched (handles pagination automatically)
- Issues stored in local libsql database at configured path
- Issue data includes: number, title, body, author, created date, updated date, comment count, labels, assignees
- Comments for each issue are fetched and stored
- Sync can be cancelled with Ctrl+C gracefully

---

### US-4: Database Storage Location
**As a** user  
**I want to** control where the local database is stored  
**So that** I can organize my data as I prefer

**Acceptance Criteria:**
- Default location is `.ghissues.db` in current working directory
- Override via `--db` flag or `database.path` in config file
- Flag takes precedence over config file
- Parent directories are created if they don't exist
- Clear error if path is not writable

---

### US-5: Issue List View
**As a** user  
**I want to** see a list of all synced issues  
**So that** I can browse and select issues to review

**Acceptance Criteria:**
- Issues displayed in left panel (vertical split layout)
- Configurable columns with defaults: number, title, author, date, comment count
- Column configuration stored in config file under `display.columns`
- Currently selected issue is highlighted
- Vim keys (`j`/`k`) and arrow keys for navigation
- Issue count shown in status area

---

### US-6: Issue Sorting
**As a** user  
**I want to** sort issues by different criteria  
**So that** I can prioritize my review

**Acceptance Criteria:**
- Default sort: most recently updated first
- Available sort options: updated date, created date, issue number, comment count
- Sort order toggled with keybinding (e.g., `s` to cycle, `S` to reverse)
- Current sort shown in status bar
- Sort preference persisted to config file

---

### US-7: Issue Detail View
**As a** user  
**I want to** view the full details of a selected issue  
**So that** I can understand the issue context

**Acceptance Criteria:**
- Right panel shows selected issue details
- Header shows: issue number, title, author, status, dates
- Body rendered with glamour (charmbracelet markdown renderer)
- Toggle between raw markdown and rendered with keybinding (e.g., `m`)
- Labels and assignees displayed if present
- Scrollable if content exceeds panel height
- `Enter` on issue list opens dedicated comments view

---

### US-8: Comments View
**As a** user  
**I want to** view all comments on an issue  
**So that** I can follow the discussion

**Acceptance Criteria:**
- Drill-down view replaces main interface when activated
- Shows issue title/number as header
- Comments displayed chronologically
- Each comment shows: author, date, body (markdown rendered)
- Toggle markdown rendering with `m` key
- Scrollable comment list
- `Esc` or `q` returns to issue list view

---

### US-9: Data Refresh
**As a** user  
**I want to** update my local issue cache  
**So that** I have the latest issue data

**Acceptance Criteria:**
- Auto-refresh triggered on app launch
- Manual refresh with keybinding (e.g., `r` or `R`)
- Progress bar shown during refresh
- Only fetches issues updated since last sync (incremental)
- Handles deleted issues (removes from local db)
- Handles new comments on existing issues

---

### US-10: Last Synced Indicator
**As a** user  
**I want to** know when my data was last updated  
**So that** I understand how current my offline data is

**Acceptance Criteria:**
- Status bar shows "Last synced: <relative time>" (e.g., "5 minutes ago")
- Timestamp stored in database metadata table
- Updates after each successful sync

---

### US-11: Keybinding Help
**As a** user  
**I want to** discover available keybindings  
**So that** I can use the app effectively

**Acceptance Criteria:**
- `?` opens help overlay with all keybindings organized by context
- Persistent footer shows context-sensitive common keys
- Footer updates based on current view (list, detail, comments)
- Help overlay dismissible with `?` or `Esc`

---

### US-12: Color Themes
**As a** user  
**I want to** choose a color theme  
**So that** the app matches my terminal aesthetic

**Acceptance Criteria:**
- Multiple built-in themes: default, dracula, gruvbox, nord, solarized-dark, solarized-light
- Theme selected via config file `display.theme`
- Theme can be previewed/changed with command `ghissues themes`
- Themes use lipgloss for consistent styling

---

### US-13: Error Handling
**As a** user  
**I want to** see clear error messages  
**So that** I can understand and resolve issues

**Acceptance Criteria:**
- Minor errors (network timeout, rate limit) shown in status bar
- Critical errors (invalid token, database corruption) shown as modal
- Modal errors require acknowledgment before continuing
- Errors include actionable guidance where possible
- Network errors suggest checking connectivity and retrying

---

### US-14: Multi-Repository Configuration
**As a** user  
**I want to** configure multiple repositories  
**So that** I can review issues from different projects

**Acceptance Criteria:**
- Config file supports multiple repository entries
- Each repository has its own database file
- `ghissues --repo owner/repo` selects which repo to view
- Default repository can be set in config
- `ghissues repos` lists configured repositories

---

## Technical Specifications

### Technology Stack
- **Language:** Go 1.22+
- **TUI Framework:** bubbletea, lipgloss, bubbles (charmbracelet)
- **Markdown Rendering:** glamour (charmbracelet)
- **Database:** libsql (github.com/tursodatabase/libsql)
- **GraphQL Client:** github.com/shurcooL/githubv4
- **Config Format:** TOML

### Project Structure
```
ghissues/
├── cmd/
│   └── ghissues/
│       └── main.go
├── internal/
│   ├── config/        # Configuration loading and validation
│   ├── github/        # GraphQL API client and queries
│   ├── database/      # libsql storage layer
│   ├── issues/        # Issue domain logic
│   ├── comments/      # Comment domain logic
│   ├── sync/          # Sync orchestration
│   └── tui/
│       ├── app.go     # Main bubbletea application
│       ├── list/      # Issue list component
│       ├── detail/    # Issue detail component
│       ├── comments/  # Comments view component
│       ├── help/      # Help overlay component
│       ├── status/    # Status bar component
│       └── theme/     # Theme definitions
├── go.mod
├── go.sum
└── config.example.toml
```

### Database Schema
```sql
CREATE TABLE repositories (
    id INTEGER PRIMARY KEY,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    last_synced_at DATETIME,
    UNIQUE(owner, name)
);

CREATE TABLE issues (
    id INTEGER PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    number INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    author TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    comment_count INTEGER DEFAULT 0,
    labels TEXT, -- JSON array
    assignees TEXT, -- JSON array
    FOREIGN KEY (repository_id) REFERENCES repositories(id),
    UNIQUE(repository_id, number)
);

CREATE TABLE comments (
    id INTEGER PRIMARY KEY,
    issue_id INTEGER NOT NULL,
    github_id TEXT NOT NULL,
    author TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (issue_id) REFERENCES issues(id),
    UNIQUE(github_id)
);

CREATE INDEX idx_issues_repository ON issues(repository_id);
CREATE INDEX idx_issues_updated ON issues(updated_at);
CREATE INDEX idx_comments_issue ON comments(issue_id);
```

### Configuration File Structure
```toml
# ~/.config/ghissues/config.toml

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

### Default Keybindings
| Key | Context | Action |
|-----|---------|--------|
| `j` / `↓` | List | Move down |
| `k` / `↑` | List | Move up |
| `g` / `Home` | List | Go to first |
| `G` / `End` | List | Go to last |
| `Enter` | List | Open comments view |
| `s` | List | Cycle sort field |
| `S` | List | Toggle sort direction |
| `r` | Any | Refresh from GitHub |
| `m` | Detail/Comments | Toggle markdown rendering |
| `?` | Any | Toggle help overlay |
| `q` | Comments | Back to list |
| `q` | List | Quit application |
| `Ctrl+C` | Any | Force quit |

### Testing Strategy
- **Unit Tests:** All business logic (config parsing, database operations, sync logic)
- **Golden File Tests:** TUI component rendering snapshots
- **Behavioral Tests:** Using `teatest` for TUI interaction flows
- **Integration Tests:** Database operations with in-memory libsql
- **Test Coverage Target:** 80%+ for non-TUI code

---

## Future Considerations (Post-MVP)
- Issue filtering and search
- Label/assignee modification
- Comment creation
- Issue state changes (open/close)
- Offline change queuing with sync
- Multiple view layouts (horizontal split, single panel)
- Custom keybinding configuration
- Export issues to markdown/JSON
- Notifications for issue updates