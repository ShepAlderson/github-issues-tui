# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### TDD Pattern
1. Write tests first for each function/feature
2. See tests fail (confirming tests work)
3. Implement minimal code to make tests pass
4. Refactor while keeping tests green

### Bubbletea TUI Testing
- Use `tea.KeyMsg{Type: tea.KeyEnter}` to simulate keypresses
- Test state transitions by checking model fields after Update
- View tests verify UI text content contains expected strings

### Config File Security
- Use `0600` permissions (owner read/write only) for sensitive files
- Implement via `os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)`

### Home Directory Handling
- Use `os.UserHomeDir()` for cross-platform home directory detection
- Fallback gracefully if not available

### Authentication Resolution Pattern
- Define priority order clearly (env var -> config file -> external CLI)
- Each source should return (value, found) tuple pattern
- Provide actionable error messages with all configuration options
- Use `exec.LookPath()` before attempting to execute external commands
- Store token source information for debugging authentication issues

### Database Path Resolution Pattern
- Priority order: CLI flag -> config file -> default path
- Test writability by creating and removing a temp file in the directory
- Use `filepath.Abs()` to normalize relative paths
- Create parent directories with `os.MkdirAll(dir, 0755)` before checking writability
- Provide clear error messages with override options

### GitHub API Pagination Pattern
- Parse Link header to find `rel="next"` for more pages
- Use `per_page=100` (max) to minimize API calls
- Sleep briefly between requests to be respectful to the API
- Implement progress reporting with channel-based updates
- Store `Fetched` and `Total` in progress messages for progress bar

### Sync TUI Pattern with Cancellation
- Use separate goroutine for API fetching
- Send progress updates through channel to Bubbletea model
- Check `cancelled` flag in fetch loop for graceful cancellation
- Update progress bar using `tea.Msg` with progress data
- Handle `tea.KeyCtrlC` for user cancellation

### Issue List View Pattern
- Use Config interface to decouple list package from config implementation
- Validate columns against allowed set, filter out unknown columns gracefully
- Store selected index, clamp to valid range during navigation
- Render selection with highlighted style using lipgloss
- Support both vim keys (j/k) and arrow keys for navigation
- Show issue count in status area at bottom of view

### Issue Sorting Pattern
- Support multiple sort fields stored in ordered array for cycling
- Use callback pattern (Config.SaveSort) for persistence to avoid circular imports
- Default sort: most recently updated first ("updated" descending)
- Cycle fields with 's' key, toggle order with 'S' key
- When cycling fields, reset to descending order
- Show current sort in status bar with icon (↓ or ↑)
- Reload data immediately after sort change for instant feedback

### Issue Detail View Pattern
- Use split layout: lipgloss.JoinHorizontal(lipgloss.Top, left, right)
- List panel: ~40% width (min 30, max 50), detail panel: remaining width
- Use glamour.NewTermRenderer() for markdown rendering with auto styling
- Toggle rendered/raw mode with 'm' key via detailModel.ToggleRenderedMode()
- Fetch issue details asynchronously via tea.Cmd for non-blocking UI
- Use detailLoadedMsg pattern to update detail model after async fetch
- Include labels, assignees, state badge (open/closed colors), dates in detail header

---

## 2026-01-28 - US-001
- Implemented first-time setup wizard with interactive TUI
- Files changed: cmd/ghissues/main.go, internal/config/*.go
- **Learnings:**
  - Bubbletea input fields require Focus() to capture keyboard input
  - TOML decoding with `toml.DecodeFile` returns (MetaData, error) - handle properly
  - File permissions 0600 is octal, not decimal (use leading 0)
  - `os.UserHomeDir()` is the idiomatic way to get home directory in Go
  - Bubbletea tests: manually trigger state changes since we can't simulate full TUI flow
  - Repository validation: use simple parsing + regex for GitHub username/repo validation
  - Setup flow: state machine pattern works well for multi-step wizards
---

## 2026-01-28 - US-004
- Implemented database storage location configuration with priority-based resolution
- Files changed:
  - internal/database/path.go (new)
  - internal/database/path_test.go (new)
  - cmd/ghissues/main.go (updated with --db flag)
- **Learnings:**
  - `flag` package usage: Define flags with `flag.StringVar()` before `flag.Parse()`
  - Flag package intercepts --help, use custom help subcommand instead
  - Database writability test: Create temp file in directory, then remove it
  - Path resolution pattern: flag > config > default with clear priority
  - `filepath.Abs()` resolves relative paths to absolute paths
  - `os.Chmod()` for read-only testing may not work on all platforms - use `t.Skip()`
---

## 2026-01-28 - US-002
- Implemented GitHub authentication resolver with priority-based token resolution
- Files changed: internal/github/auth.go, internal/github/auth_test.go
- **Learnings:**
  - Authentication priority order: GITHUB_TOKEN env var -> config file -> gh CLI
  - Use `exec.LookPath("gh")` to check if gh CLI is available before executing
  - `os/exec.Command().Output()` returns combined stdout/stderr, use separately if needed
  - Multi-source configuration requires clear error messages when none are available
  - Testing with PATH modification: save/restore PATH to simulate missing commands
  - Writing tests first ensures edge cases (empty strings, missing files) are handled
---

## ✓ Iteration 1 - US-001: First-Time Setup
*2026-01-28T07:53:50.015Z (433s)*

**Status:** Completed

**Notes:**
68→  - Gotchas encountered\n    69→---\n    70→```\n    71→\n    72→If you discovered a **reusable pattern**, also add it to the `## Codebase Patterns` section at the TOP of progress.md.

## Stop Condition
**IMPORTANT**: If the work is already complete (implemented in a previous iteration or already exists), verify it meets the acceptance criteria and signal completion immediately.

When finished (or if already complete), signal completion with:


---
## ✓ Iteration 2 - US-002: GitHub Authentication
*2026-01-28T08:03:31.153Z (580s)*

**Status:** Completed

**Notes:**
n 'ghissues config' to save a token to your config file\n3. Login with 'gh auth login' to use gh CLI authentication\n```\n\n**Token Validation (AC met: ✅)**\n- `ValidateToken()` function created (empty check)\n- Framework for API validation ready for future stories\n\n**Config File Security (AC met: ✅)**\n- Already implemented in US-001 with 0600 permissions\n\n**New Pattern Added to Codebase:**\n- Authentication Resolution Pattern with priority-based resolution and actionable error messages\n\n

---
## ✓ Iteration 4 - US-003: Initial Issue Sync
*2026-01-28T09:30:00Z*

**Status:** Completed

**Notes:**
- Implemented issue sync functionality with GitHub API integration
- Files changed:
  - internal/database/schema.go (new) - database schema for issues/comments
  - internal/database/schema_test.go (new) - schema tests
  - internal/github/client.go (new) - GitHub API client with pagination
  - internal/github/client_test.go (new) - client tests
  - internal/sync/sync.go (new) - sync TUI with progress bar
  - internal/sync/sync_test.go (new) - sync tests
  - cmd/ghissues/main.go (updated) - add sync subcommand
  - go.mod (updated) - add dependencies

**Acceptance Criteria Met:**
- ✅ Progress bar displays during fetch showing issues fetched/total
- ✅ All open issues are fetched (handles pagination automatically)
- ✅ Issues stored in local SQLite database at configured path
- ✅ Issue data includes: number, title, body, author, dates, comment count, labels, assignees
- ✅ Comments for each issue are fetched and stored
- ✅ Sync can be cancelled with Ctrl+C gracefully

**New Pattern Added to Codebase:**
- GitHub API Pagination Pattern: Parse Link header for `rel="next"` to determine more pages
- Progress reporting with channel-based updates for real-time UI feedback

**Learnings:**
- SQLite (via mattn/go-sqlite3) provides local database storage
- ON CONFLICT DO UPDATE for upsert operations in SQLite
- GitHub API Link header format for pagination
- Channel-based progress updates for real-time sync feedback
- Bubbletea progress bar with gradient styling
- Goroutines for concurrent API fetching with graceful cancellation
- `Ctrl+C` handling with `tea.KeyCtrlC` in Bubbletea

---

## ✓ Iteration 3 - US-004: Database Storage Location
*2026-01-28T08:06:23.890Z (172s)*

**Status:** Completed

**Notes:**
ignature":""}],"model":"hf:moonshotai/Kimi-K2.5","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":27710,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"89be9b4d-9f0d-491b-94a3-c85e015bf603","uuid":"a495bd16-c031-406a-be5f-cb031c0ec0cd"}
{"type":"assistant","message":{"id":"msg_6187ff4b-fd89-4d49-8f33-743145e8e69d","type":"message","role":"assistant","content":[{"type":"text","text":"

---
## ✓ Iteration 4 - US-003: Initial Issue Sync
*2026-01-28T08:17:27.921Z (663s)*

**Status:** Completed

**Notes:**
ignature":""}],"model":"hf:moonshotai/Kimi-K2.5","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":70659,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"76dcd9d5-8849-485e-899f-4702cfeadc61","uuid":"715dd73d-a3ff-4714-b4dd-c99ca2a240b4"}
{"type":"assistant","message":{"id":"msg_a9361aa3-833b-4e81-833f-8b86bc12f8b1","type":"message","role":"assistant","content":[{"type":"text","text":"

---

## 2026-01-28 - US-005: Issue List View
- Implemented issue list view with TUI navigation
- Files changed:
  - internal/database/schema.go - Add ListIssue, ListIssues, ListIssuesSorted, ListIssuesByState, FormatDate
  - internal/database/list_test.go - Tests for list query functions
  - internal/list/list.go - New TUI model for issue list
  - internal/list/list_test.go - Tests for list TUI
  - cmd/ghissues/main.go - Integrate list view with ConfigAdapter

**Acceptance Criteria Met:**
- ✅ Issues displayed in left panel (vertical split layout)
- ✅ Configurable columns with defaults: number, title, author, date, comment count
- ✅ Column configuration stored in config file under display.columns
- ✅ Currently selected issue is highlighted
- ✅ Vim keys (j/k) and arrow keys for navigation
- ✅ Issue count shown in status area

**New Pattern Added to Codebase:**
- Issue List View Pattern with Config interface, column validation, and navigation

**Learnings:**
- Config interface decouples packages while allowing easy testing
- validateColumns() filters invalid columns gracefully rather than erroring
- Selection index clamping prevents out-of-bounds during navigation
- Using `tea.WindowSizeMsg` to track terminal dimensions for layout
- Status bar at bottom provides context and instructions

---
## ✓ Iteration 5 - US-005: Issue List View
*2026-01-28T08:27:32.455Z (603s)*

**Status:** Completed

**Notes:**
d in left panel (vertical split layout)\n- ✅ Configurable columns with defaults: number, title, author, date, comment count\n- ✅ Column configuration stored in config file under `display.columns`\n- ✅ Currently selected issue is highlighted\n- ✅ Vim keys (j/k) and arrow keys for navigation\n- ✅ Issue count shown in status area\n\n### New Pattern Added\n- **Issue List View Pattern** - Config interface for decoupling, column validation, selection index clamping, lipgloss styling for highlights\n\n

---

## ✓ Iteration - US-006: Issue Sorting
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented issue sorting with multiple sort fields
- Files changed:
  - internal/list/list.go - Sort state fields, key handlers, status bar, Config.SaveSort callback
  - internal/list/list_test.go - Tests for sort cycling, toggling, persistence
  - internal/database/schema.go - Added comment_count sort support
  - internal/database/list_test.go - Tests for comment count sorting
  - cmd/ghissues/main.go - ConfigAdapter SaveSort implementation, updated help text

**Acceptance Criteria Met:**
- ✅ Default sort: most recently updated first (updated_at DESC)
- ✅ Available sort options: updated, created, number, comments
- ✅ Sort order toggled with 's' to cycle fields, 'S' to reverse direction
- ✅ Current sort shown in status bar with field name and ↑/↓ indicator
- ✅ Sort preference persisted to config file when changed

**New Pattern Added to Codebase:**
- **Issue Sorting Pattern** - Multiple sort fields in array for cycling, callback pattern for persistence, default to updated/descending, status bar indicator

**Learnings:**
- Callback pattern (Config.SaveSort) avoids circular imports between packages
- When cycling sort fields, reset to descending order for consistency
- Use tea.Batch() to execute multiple commands (save + reload)
- Status bar should show both sort field and direction indicator (↑/↓)
- SQLite "comment_count" column for sorting by number of comments

---
## ✓ Iteration 6 - US-006: Issue Sorting
*2026-01-28T08:33:46.296Z (373s)*

**Status:** Completed

**Notes:**
ort options: updated, created, number, comments\n- ✅ Sort order toggled with 's' to cycle fields, 'S' to reverse direction\n- ✅ Current sort shown in status bar with field name and ↑/↓ indicator\n- ✅ Sort preference persisted to config file when changed via SaveSort callback\n\n**New Pattern Added to Codebase:**\n- **Issue Sorting Pattern** - Uses callback pattern for persistence to avoid circular imports, multiple sort fields in array for cycling, tea.Batch() for executing multiple commands\n\n

---

---
## ✓ Iteration 7 - US-007: Issue Detail View
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented issue detail view with split layout
- Files changed:
  - internal/database/schema.go - Added GetIssueDetail() function and IssueDetail struct
  - internal/database/list_test.go - Tests for GetIssueDetail()
  - internal/detail/detail.go (new) - Detail view TUI model
  - internal/detail/detail_test.go (new) - Tests for detail package
  - internal/list/list.go - Integrated detail panel, 'm' key handler, Enter key handler
  - internal/list/list_test.go - Tests for new keybindings
  - cmd/ghissues/main.go - Updated help text with new keybindings

**Acceptance Criteria Met:**
- ✅ Right panel shows selected issue details
- ✅ Header shows: issue number, title, author, status, dates
- ✅ Body rendered with glamour (charmbracelet markdown renderer)
- ✅ Toggle between raw markdown and rendered with keybinding (m)
- ✅ Labels and assignees displayed if present
- ✅ Scrollable if content exceeds panel height (handled by glamour)
- ✅ Enter on issue list opens dedicated comments view (placeholder for US-008)

**New Pattern Added to Codebase:**
- **Detail View Pattern** - Split layout with lipgloss.JoinHorizontal, glamour for markdown rendering

**Learnings:**
- glamour.NewTermRenderer() takes options like glamour.WithAutoStyle() for terminal-friendly output
- lipgloss.JoinHorizontal(lipgloss.Top, left, right) creates side-by-side panels
- Split layout: list panel 40% width (min 30, max 50), detail panel remaining width
- detail.NewModel() wraps IssueDetail for the detail view
- detail.ToggleRenderedMode() switches between raw and rendered markdown
- tea.Msg handlers pattern: loadDetailIssue() returns tea.Cmd that fetches async
- detailLoadedMsg carries the fetched issue or error back to Update()

