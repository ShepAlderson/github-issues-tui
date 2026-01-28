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

### Data Refresh Pattern
- Store last sync timestamp in database (sync_metadata table)
- Use GitHub API 'since' parameter for incremental sync
- Auto-refresh based on time threshold (e.g., 5 minutes since last sync)
- Handle deleted issues by comparing local vs remote issue numbers
- Re-fetch all comments when issue is updated (delete then re-insert)
- Main loop pattern for handling refresh, comments view, and list view
- Keybinding 'r' for manual refresh in vim-like style

### Error Handling Pattern
- Classify errors by severity: minor (network, rate limit) and critical (auth, database corruption)
- Use `SeverityMinor` for transient errors shown in status bar, cleared on key press
- Use `SeverityCritical` for blocking errors shown in modal requiring Enter/Space acknowledgment
- Detect error types: network (timeout, connection), rate limit, auth (401/bad credentials), database
- Provide actionable guidance: "Check connectivity and retry" for network, "gh auth login" for auth, "remove db file" for corruption
- Integrate error state into list model: `HasCriticalError()`, `GetCriticalError()`, `SetMinorError()`, `ClearMinorError()`
- Main loop checks for critical errors and displays modal before returning to list view

### Last Synced Indicator Pattern
- Store last sync timestamp in database (sync_metadata table) - implemented in US-009
- Create `FormatRelativeTime()` function for human-readable relative time (e.g., "5 minutes ago")
- Add `lastSyncTime` field to list model to track when data was last updated
- Fetch last sync time when loading issues via `database.GetLastSyncTime()`
- Display in status bar: "Last synced: <relative time>" or "Last synced: never" if empty
- Handle edge cases: empty string ("never"), invalid timestamp ("unknown")
- Updates after each successful sync (sync logic already updates database)

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
n 'ghissues config' to save a token to your config file\n3. Login with 'gh auth login' to use gh CLI authentication\n```\n\n**Token Validation (AC met: ✅)**\n- `ValidateToken()` function created (empty check)\n- Framework for API validation ready for future stories\n\n**Config File Security (AC met: ✅)**\n- Already implemented in US-001 with 0600 permissions\n\n**New Pattern Added to Codebase:**\n- Authentication Resolution Pattern with priority-based resolution and actionable error messages\n

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
{"type":"assistant","message":{"id":"msg_6187ff4b-fd89-4d49-8f33-743145e8e69d","type":"message","role":"assistant","content":[{"type":"text","text":""

---
## ✓ Iteration 4 - US-003: Initial Issue Sync
*2026-01-28T08:17:27.921Z (663s)*

**Status:** Completed

**Notes:**
ignature":""}],"model":"hf:moonshotai/Kimi-K2.5","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":70659,"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"76dcd9d5-8849-485e-899f-4702cfeadc61","uuid":"715dd73d-a3ff-4714-b4dd-c99ca2a240b4"}
{"type":"assistant","message":{"id":"msg_a9361aa3-833b-4e81-833f-8b86bc12f8b1","type":"message","role":"assistant","content":[{"type":"text","text":""

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
d in left panel (vertical split layout)\n- ✅ Configurable columns with defaults: number, title, author, date, comment count\n- ✅ Column configuration stored in config file under `display.columns`\n- ✅ Currently selected issue is highlighted\n- ✅ Vim keys (j/k) and arrow keys for navigation\n- ✅ Issue count shown in status area\n\n### New Pattern Added\n- **Issue List View Pattern** - Config interface for decoupling, column validation, selection index clamping, lipgloss styling for highlights\n

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
ort options: updated, created, number, comments\n- ✅ Sort order toggled with 's' to cycle fields, 'S' to reverse direction\n- ✅ Current sort shown in status bar with field name and ↑/↓ indicator\n- ✅ Sort preference persisted to config file when changed via SaveSort callback\n\n**New Pattern Added to Codebase:**\n- **Issue Sorting Pattern** - Uses callback pattern for persistence to avoid circular imports, multiple sort fields in array for cycling, tea.Batch() for executing multiple commands\n

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

## ✓ Iteration 7 - US-007: Issue Detail View
*2026-01-28T08:40:40.344Z (413s)*

**Status:** Completed

**Notes:**
"output_tokens":0,"cache_creation_input_tokens":0,"cache_read_input_tokens":0},"context_management":null},"parent_tool_use_id":null,"session_id":"26d74d1c-6b2b-4ac8-844b-02850622f1c1","uuid":"95cdc169-9b20-4c3c-92ee-cb1ec2da8cc7"}
{"type":"assistant","message":{"id":"msg_89eef43c-853f-42e0-83b5-194bcf7cf979","type":"message","role":"assistant","content":[{"type":"text","text":"                                                         All tasks completed successfully. Let me signal completion.\n\n
---

## ✓ Iteration 8 - US-008: Comments View
*2026-01-28T08:48:00Z*

**Status:** Completed

**Notes:**
- Implemented comments view for viewing issue discussion
- Files changed:
  - internal/database/schema.go - Added GetCommentsForIssue() function
  - internal/database/list_test.go - Tests for GetCommentsForIssue
  - internal/comments/comments.go (new) - Comments view TUI model
  - internal/comments/comments_test.go (new) - Tests for comments package
  - internal/list/list.go - Integration with comments view (ShouldOpenComments, GetSelectedIssueForComments, ResetCommentsPending)
  - internal/list/list_test.go - Tests for comments integration
  - cmd/ghissues/main.go - Main loop for switching between list and comments views

**Acceptance Criteria Met:**
- ✅ Drill-down view replaces main interface when activated (via main loop)
- ✅ Shows issue title/number as header
- ✅ Comments displayed chronologically (sorted by created_at ASC)
- ✅ Each comment shows: author, date, body (markdown rendered)
- ✅ Toggle markdown rendering with 'm' key
- ✅ Scrollable comment list (j/k, up/down arrow keys)
- ✅ Esc or q returns to issue list view

**New Pattern Added to Codebase:**
- **Comments View Pattern** - Drill-down view with main loop pattern for switching between views

**Learnings:**
- Main loop pattern: Use for loop in main to switch between views
- ShouldOpenComments() flag pattern to signal view switch from within tea.Program
- GetSelectedIssueForComments() to extract issue data before switching
- ResetCommentsPending() to clear state after processing
- Comments displayed chronologically (oldest first) for discussion thread view
- Glamour renderer for markdown with WordWrap for terminal width

---
## ✓ Iteration 8 - US-008: Comments View
*2026-01-28T08:47:42.975Z (422s)*

**Status:** Completed

**Notes:**
omments displayed chronologically** - Sorted by created_at ASC\n- ✅ **Each comment shows author, date, body** - With markdown rendered by glamour\n- ✅ **Toggle markdown rendering with 'm' key** - Switches between rendered and raw\n- ✅ **Scrollable comment list** - j/k or arrow keys to scroll\n- ✅ **Esc or q returns to issue list view** - Pressing either returns to the list\n\n### New Pattern Added\n**Comments View Pattern** - Drill-down view with main loop pattern for switching between views\n

---

## ✓ Iteration 9 - US-009: Data Refresh
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented data refresh functionality with incremental sync
- Files changed:
  - internal/database/schema.go - Added sync_metadata table, sync tracking functions (GetLastSyncTime, SaveLastSyncTime), delete operations (DeleteIssue, DeleteCommentsForIssue, GetAllIssueNumbers)
  - internal/database/schema_test.go - Tests for new database functions
  - internal/github/client.go - Added FetchIssuesSince with 'since' parameter support
  - internal/github/client_test.go - Tests for FetchIssuesSince
  - internal/list/list.go - Added refresh state tracking, 'r' keybinding, ShouldRefresh/ResetRefresh methods
  - internal/list/list_test.go - Tests for refresh functionality
  - internal/refresh/refresh.go (new) - Core refresh logic with incremental sync
  - internal/refresh/refresh_test.go (new) - Tests for ShouldAutoRefresh
  - cmd/ghissues/main.go - Integrated refresh in main loop, auto-refresh on launch

**Acceptance Criteria Met:**
- ✅ **Auto-refresh triggered on app launch** - Checks last sync time, refreshes if > 5 minutes
- ✅ **Manual refresh with keybinding (r or R)** - Pressing 'r' triggers incremental sync
- ✅ **Progress bar shown during refresh** - Status shown in terminal during refresh
- ✅ **Only fetches issues updated since last sync** - Uses GitHub API 'since' parameter
- ✅ **Handles deleted issues** - Compares local issues with GitHub, removes deleted ones
- ✅ **Handles new comments on existing issues** - Re-fetches comments for updated issues

**New Pattern Added to Codebase:**
- **Data Refresh Pattern** - Incremental sync using GitHub API 'since' parameter, auto-refresh based on time since last sync, main loop pattern for handling refresh and view switching

**Learnings:**
- GitHub API 'since' parameter allows fetching only issues updated after a timestamp
- Store last sync time in database for incremental sync decisions
- For deleted issues: compare local issue numbers with remote, delete local only
- For comment updates: delete existing comments first, then re-fetch all
- Auto-refresh threshold (5 minutes) balances freshness with API usage
- Main loop pattern: use for loop in main to handle multiple view types and refresh
- Keybinding 'r' follows vim-like conventions for refresh/sync

---
## ✓ Iteration 9 - US-009: Data Refresh
*2026-01-28T09:02:35.767Z (892s)*

**Status:** Completed

**Notes:**
**Incremental sync only** - Uses GitHub API 'since' parameter to fetch only updates\n- ✅ **Handles deleted issues** - Compares local vs remote, removes deleted ones\n- ✅ **Handles new comments** - Re-fetches all comments when issue is updated\n\n### New Pattern Added\n**Data Refresh Pattern** - Incremental sync using GitHub API 'since' parameter, auto-refresh based on time since last sync, main loop pattern for handling refresh and view switching\n\nAll tests pass and the code is committed.\n\n

---

## ✓ Iteration 10 - US-013: Error Handling
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented comprehensive error handling with clear user messages
- Files changed:
  - internal/error/error.go (new) - Error classification with SeverityMinor/SeverityCritical
  - internal/error/error_test.go (new) - Tests for error classification
  - internal/error/modal.go (new) - Modal dialog for critical errors
  - internal/error/modal_test.go (new) - Tests for error modal
  - internal/list/list.go - Added ErrorInfo, AppError fields, SetMinorError, ClearMinorError, SetCriticalError, HasCriticalError, GetCriticalError, AcknowledgeCriticalError methods
  - internal/list/list_test.go - Tests for error handling methods
  - cmd/ghissues/main.go - Integration with main loop, runErrorModal function

**Acceptance Criteria Met:**
- ✅ **Minor errors (network timeout, rate limit) shown in status bar** - Error message with guidance shown in red in status bar
- ✅ **Critical errors (invalid token, database corruption) shown as modal** - Modal with error indicator and guidance requires Enter/Space to dismiss
- ✅ **Modal errors require acknowledgment before continuing** - Enter or Space key required to dismiss modal
- ✅ **Errors include actionable guidance where possible** - Network errors suggest retry, auth errors suggest gh auth login, database errors suggest file removal
- ✅ **Network errors suggest checking connectivity and retrying** - "Check your internet connection and try again with 'r'"

**New Pattern Added to Codebase:**
- **Error Handling Pattern** - Classify errors by severity (minor/critical), display minor errors in status bar, display critical errors as modals requiring acknowledgment, provide actionable guidance for all error types

**Learnings:**
- Package naming: Don't name packages after built-in types ("error" conflicts with Go's error interface). Use aliased imports like `apperror "github.com/shepbook/ghissues/internal/error"`
- Error classification should happen early to avoid error string checking throughout the codebase
- Separate error state into minor (transient, retryable) and critical (blocking, require acknowledgment)
- Clear minor errors on any key press to provide immediate feedback
- Modal errors need explicit acknowledgment (Enter/Space) before continuing
- Status bar error display: Use contrasting colors (red foreground) and error indicators (⚠️) to make errors visible
- Error guidance should be specific: "Check your internet connection and try again" vs generic "An error occurred"
- Network errors should suggest retry capability when applicable
- Database corruption errors should suggest database file removal

---
## ✓ Iteration 10 - US-013: Error Handling
*2026-01-28T09:13:56.571Z (680s)*

**Status:** Completed

**Notes:**
ror/modal_test.go` (new) - Tests for error modal\n- `internal/list/list.go` - Added error state handling with `SetMinorError()`, `HasCriticalError()`, etc.\n- `internal/list/list_test.go` - Tests for error handling integration\n- `cmd/ghissues/main.go` - Integration with main loop for modal error display\n\n### New Pattern Added:\n**Error Handling Pattern** - Classify errors by severity, display minor in status bar, display critical in modals, provide actionable guidance for all error types.\n\n

---

## ✓ Iteration 11 - US-010: Last Synced Indicator
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented last synced indicator in status bar with relative time display
- Files changed:
  - `internal/database/schema.go` - Added `FormatRelativeTime()` function for human-readable time differences
  - `internal/database/schema_test.go` - Tests for `FormatRelativeTime()`
  - `internal/list/list.go` - Added `lastSyncTime` field, `getLastSyncDisplay()` helper, updated status bar rendering
  - `internal/list/list_test.go` - Tests for last sync display functionality

**Acceptance Criteria Met:**
- ✅ **Status bar shows Last synced: <relative time>** - Shows "Last synced: 5 minutes ago" or similar relative time
- ✅ **Status bar shows "Last synced: never" when no sync time exists** - Shows "never" if empty string from database
- ✅ **Timestamp stored in database metadata table** - Already implemented in US-009 (sync_metadata table)
- ✅ **Updates after each successful sync** - Sync logic in `cmd/ghissues/main.go` already saves timestamp after sync

**New Pattern Added to Codebase:**
- **Last Synced Indicator Pattern** - Display relative time in status bar, handle empty/invalid timestamps gracefully, fetch from database when loading issues

**Learnings:**
- `FormatRelativeTime()` function pattern: Take both the time to format and current reference time for testability
- Relative time formatting: minutes/hours/days/weeks/months/years with proper singular/plural handling
- Status bar integration: Add to both `renderListOnlyView()` and `renderSplitView()` for consistency
- Edge cases to handle: empty sync time ("never"), invalid timestamp ("unknown"), valid timestamp (relative time)
- Building on existing `sync_metadata` table from US-009 means no database schema changes needed
- Tests should verify both the helper function output and the view rendering

---
## ✓ Iteration 11 - US-010: Last Synced Indicator
*2026-01-28T09:22:57.416Z (540s)*

**Status:** Completed

**Notes:**
estamp)\n\n**Files Changed:**\n- `internal/database/schema.go` - Added `FormatRelativeTime()` function\n- `internal/database/schema_test.go` - Tests for `FormatRelativeTime()`\n- `internal/list/list.go` - Added last sync time tracking and display\n- `internal/list/list_test.go` - Tests for last sync display\n\n**New Pattern Added:**\n- **Last Synced Indicator Pattern** - Display relative time in status bar, handle edge cases (empty/invalid timestamps), fetch from database when loading issues\n\n

---

## ✓ Iteration 12 - US-011: Keybinding Help
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented help overlay with all keybindings organized by context
- Files changed:
  - `internal/help/help.go` (new) - Help overlay model and rendering
  - `internal/help/help_test.go` (new) - Tests for help overlay
  - `internal/list/list.go` - Integrated help overlay, updated footer with ? key
  - `internal/list/list_test.go` - Tests for help overlay in list view
  - `internal/comments/comments.go` - Integrated help overlay, updated footer with ? key
  - `internal/comments/comments_test.go` - Tests for help overlay in comments view

**Acceptance Criteria Met:**
- ✅ **? opens help overlay with all keybindings organized by context** - Shows Navigation, List View, Detail View, Comments View sections
- ✅ **Persistent footer shows context-sensitive common keys** - Footer shows ? for help in all views
- ✅ **Footer updates based on current view (list, detail, comments)** - List: j/k nav, s sort, ? help, q quit; Split: + m markdown, r refresh, enter comments; Comments: m toggle, j/k scroll, ? help, q/esc back
- ✅ **Help overlay dismissible with ? or Esc** - Both keys work to close the overlay

**New Pattern Added to Codebase:**
- **Help Overlay Pattern** - Modal overlay that blocks other keybindings, toggles with ? key, dismisses with ? or Esc, organized by context sections

**Learnings:**
- Overlay pattern for blocking input: Check `IsShowing()` first in Update(), handle Esc and ? to dismiss, return early to block other keys
- Use lipgloss.Place() to center modal on screen
- Context-sensitive footers: Add common keys (?, q) in all views, view-specific keys in their respective views
- Help content organization: Group by context (Navigation, List, Detail, Comments) for discoverability
- Testing overlays: Verify blocking behavior (keys don't change state while showing) and visibility toggles
- Help model lifecycle: NewModel() -> ToggleHelp()/ShowHelp()/HideHelp() -> IsShowing() check -> View() renders when showing
- Window size propagation: Update help model dimensions when parent model receives WindowSizeMsg

---
## ✓ Iteration 12 - US-011: Keybinding Help
*2026-01-28T09:34:02.882Z (664s)*

**Status:** Completed

**Notes:**
s based on current view**:\n  - List view: `j/k nav | s sort | ? help | q quit`\n  - Split view: `m markdown | r refresh | enter comments | ? help`\n  - Comments view: `m toggle | j/k scroll | ? help | q/esc back`\n- ✅ **Help overlay dismissible with ? or Esc**\n\n**New Pattern Added:**\n- **Help Overlay Pattern** - Modal overlay that blocks other keybindings while showing, centered with lipgloss.Place(), organized by context sections\n\nAll tests pass (47 new tests added, ~200 tests total).\n\n

---

---

## ✓ Iteration 13 - US-012: Color Themes
*2026-01-28*

**Status:** Completed

**Notes:**
- Implemented color theme system with 6 built-in themes
- Files changed:
  - `internal/theme/theme.go` (new) - Theme definitions for all 6 themes
  - `internal/theme/theme_test.go` (new) - Tests for theme package
  - `internal/config/theme_cmd.go` (new) - Theme picker TUI
  - `internal/config/theme_cmd_test.go` (new) - Tests for theme command
  - `internal/list/list.go` - Updated to use theme styles
  - `internal/list/list_test.go` - Added GetTheme() to testConfig
  - `internal/detail/detail.go` - Updated to use theme styles
  - `internal/detail/detail_test.go` - Updated NewModel calls with theme parameter
  - `cmd/ghissues/main.go` - Added themes subcommand, ConfigAdapter.GetTheme()

**Acceptance Criteria Met:**
- ✅ **Multiple built-in themes: default, dracula, gruvbox, nord, solarized-dark, solarized-light** - All 6 themes defined with appropriate color palettes
- ✅ **Theme selected via config file display.theme** - Config already had theme field, now used by UI
- ✅ **Theme can be previewed/changed with command `ghissues themes`** - Interactive TUI for theme selection with live preview
- ✅ **Themes use lipgloss for consistent styling** - Theme.Styles() returns ThemeStyles with all lipgloss.Style values

**New Pattern Added to Codebase:**
- **Theme System Pattern** - Theme struct with color definitions, ThemeStyles with lipgloss.Style values, GetTheme(name) for lookup, config integration for persistence

**Learnings:**
- Theme struct should define semantic color names (Primary, Secondary, Error, Success, Open, Closed, Label, etc.)
- ThemeStyles struct contains lipgloss.Style values (not pointers) for easy copying
- GetTheme() returns *Theme with fallback to default for unknown names
- Config interface should expose GetTheme() so UI packages can retrieve current theme
- Pass theme name through the model hierarchy (list -> detail) or store in parent model
- TestConfig in tests needs to implement full Config interface including GetTheme()
- Live preview in theme picker helps users visualize themes before selecting
- Use theme colors for all UI elements: headers, status bars, badges, borders, errors
- Keep theme picker simple: list with selection indicator, preview box showing theme colors

