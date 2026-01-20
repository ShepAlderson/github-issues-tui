# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### Testing Pattern
- Use dependency injection for testability (e.g., `RunInteractiveSetupWithReader` allows passing custom io.Reader for tests)
- Write tests first (TDD), see them fail, then implement
- Use table-driven tests for validation logic with multiple cases
- Use httptest.NewServer for testing HTTP clients (GitHub API)

### Configuration Pattern
- Config stored in `~/.config/ghissues/config.toml` (follows XDG conventions)
- Use 0600 permissions for security when storing sensitive data (tokens)
- Validate config on load with clear error messages
- Support multiple auth methods: env variable, token in config, gh CLI

### CLI Pattern
- Subcommands implemented via simple argument parsing
- Provide --help and --version flags
- Interactive setup on first run with validation loops for invalid input
- Clear error messages with actionable guidance

### Authentication Pattern
- Multiple auth methods with fallback order: env variable -> config file -> external CLI
- Switch-based method dispatch for clean authentication selection
- Token validation via API call with 10-second timeout
- Error messages include actionable next steps (e.g., "run 'ghissues config'")
- Use `exec.Command` for external CLI integration (e.g., `gh auth token`)

### Database Pattern
- Three-tier precedence: CLI flag > config file > default value
- Always resolve relative paths to absolute paths for consistency
- Test writability by attempting to create a test file (more reliable than permission checks)
- Initialize idempotently (succeed whether target exists or not)
- Create parent directories automatically with `os.MkdirAll`
- Use transactions for multi-step database operations
- Use `INSERT OR REPLACE` for upsert operations in SQLite
- **SQLite Driver**: Use `modernc.org/sqlite` (pure Go) instead of libsql for better test compatibility
  - Driver name: "sqlite" (not "libsql")
  - Connection string: direct file path (not "file:" prefix)
  - Import driver in all files that open database connections: `_ "modernc.org/sqlite"`

### API Client Pattern
- Use context.Context for cancellation support
- Implement pagination automatically using Link headers
- Parse response headers for pagination metadata
- Use httptest.NewServer for testing API clients
- Set appropriate headers (Authorization, Accept) for API requests
- Respect per_page limits (GitHub API: max 100)

### Progress Tracking Pattern
- Use progressbar library for visual feedback
- Set clear descriptions and show counts
- Clear progress bar on finish for clean output
- Update progress after each unit of work completes

### Signal Handling Pattern
- Use signal.Notify for Ctrl+C handling
- Create cancellable context for graceful shutdown
- Handle cancellation in goroutine to avoid blocking main logic
- Clean up resources (close database, etc.) with defer

### TUI Pattern (Bubbletea)
- Use Bubbletea's Model-Update-View architecture for interactive UIs
- Model holds state (issues, cursor position, dimensions)
- Update handles messages (key presses, window resizes) and returns updated model
- View renders the current state to a string
- Use lipgloss for styling (colors, bold, etc.)
- Handle both vim keys (j/k) and arrow keys for navigation
- Use tea.WithAltScreen() to take over terminal and restore on exit
- Implement boundary checks for navigation (don't go past first/last item)
- Defer database Close() to ensure cleanup even if TUI exits early
- **Split-pane layout**: Calculate panel widths as percentages, render each panel independently, combine line-by-line with separator
- **Conditional rendering**: Render simple view before dimensions arrive (WindowSizeMsg), switch to rich layout after
- **Glamour for markdown**: Use glamour.NewTermRenderer with WithAutoStyle() and WithWordWrap() for terminal markdown rendering
- **Scrollable panels**: Track scroll offset, split content into lines, slice visible range, clamp offsets to prevent bounds errors
- **View mode with lazy loading**: Use viewMode field (integer constants) to switch between views. Add view-specific state fields. Dispatch in Update() and View() based on mode. Load data only when entering view, clear when leaving.

### Incremental Sync Pattern
- Use metadata table (key-value pairs) to track sync state like last_sync_time
- GitHub API supports `since` parameter (RFC3339 format) to fetch only updated issues
- Use `state=all` in incremental sync to detect both open and closed issues
- Remove closed issues from local database after incremental sync
- Fall back to full sync if no previous sync timestamp exists
- Store timestamp AFTER successful sync, not before (prevents missed data on failure)
- Auto-refresh on app launch keeps data current without user intervention
- Provide both full sync and incremental refresh as separate commands for user control

### Error Handling Pattern (TUI)
- **Two-tier error system**: Status bar errors for minor issues (network timeout, rate limit) that don't block interaction. Modal errors for critical issues (auth failure, database corruption) that require acknowledgment.
- **Error messages as Bubbletea messages**: Define custom message types (StatusErrorMsg, ClearStatusErrorMsg, ModalErrorMsg) for error reporting
- **Modal blocking**: When modal error is active, check at start of Update() and only allow Ctrl+C (quit) and Enter (acknowledge)
- **Error overlay rendering**: Render base view first, then overlay modal on top to maintain context
- **Actionable error messages**: Include operation context in error messages (e.g., "Failed to load comments: database closed")
- **Status bar error styling**: Use red color and bold styling for errors, appended with bullet separator
- **Test error scenarios**: Use closed database or mock failures to test error handling paths reliably

### Help Overlay Pattern (TUI)
- **Help as view mode**: Create dedicated viewModeHelp constant for help overlay, accessible from any view
- **previousViewMode tracking**: Store which view help was opened from, allowing return to correct location on dismiss
- **Toggle key pattern**: Use same key ('?') to both open and dismiss help overlay for intuitive interaction
- **Context-sensitive footer**: Create renderFooter() method that shows relevant keybindings based on current viewMode
- **Organized help structure**: Group keybindings by category (Navigation, Sorting, View Options, Application) with consistent formatting
- **Help key in all views**: Add '?' handler to each view mode's key handling function to enable help from anywhere
- **Block keys in help view**: When help is active, only allow dismiss keys (?, Esc) and global quit (Ctrl+C)
- **Help discoverability**: Show "?: help" in footer of all views as persistent reminder

### Theme System Pattern (TUI)
- **Separate theme package**: Create internal/theme package with Theme struct containing all lipgloss.Style fields used by TUI
- **Complete theme definitions**: Each theme function returns complete Theme with all styles defined (no partial themes or inheritance)
- **Theme access functions**: Provide GetTheme(name) that returns theme by name with fallback to default, and ListThemes() to get all themes
- **Theme integration**: Add theme field to Model struct, pass theme name (string) to constructor, load theme inside constructor with theme.GetTheme()
- **Replace global styles**: Replace all global style variables with m.theme.StyleName throughout rendering code (use sed for bulk replacement)
- **Config validation**: Add config.GetTheme() helper that validates theme name and returns "default" for unknown themes
- **Theme preview command**: Provide CLI command to preview themes with sample colored output using actual theme styles
- **Case-insensitive matching**: Use lowercase and trimmed theme names for matching to be user-friendly

### Multi-Resource Configuration Pattern
- **Resource array in config**: Use TOML `[[resource]]` syntax for array of resources (e.g., [[repositories]])
- **Default resource field**: Add default_resource field to select which resource to use by default
- **Precedence resolution**: Implement GetResource(cfg, flag) with 4-tier precedence: CLI flag > default > first in list > legacy field
- **Backward compatibility**: Keep legacy single-value resource field marked `omitempty` for existing configs
- **Per-resource data isolation**: Store each resource's data separately using standardized paths (~/.local/share/app/<resource_name>)
- **List command**: Provide command to list configured resources with default marked (e.g., "repos" command)
- **Default validation**: Validate default_resource exists in resource list during config validation
- **Resource name normalization**: Convert resource names to filesystem-safe format (e.g., owner/repo → owner_repo)

---


[...older entries truncated...]

## ✓ Iteration 5 - US-008: Comments View
*2026-01-19T22:57:15.245Z (435s)*

**Status:** Completed

**Notes:**
aving to free memory\n\n### All Acceptance Criteria Met ✅\n\n- ✅ Drill-down view replaces main interface when activated\n- ✅ Shows issue title/number as header\n- ✅ Comments displayed chronologically\n- ✅ Each comment shows: author, date, body (markdown rendered)\n- ✅ Toggle markdown rendering with m key\n- ✅ Scrollable comment list\n- ✅ Esc or q returns to issue list view\n\n### Commits\n\n- `3e29382` - feat: US-008 - Comments View\n- `e081594` - docs: document US-008 learnings and patterns\n\n

---

## [2026-01-19] - US-009 - Data Refresh

### What was implemented
- Metadata table to track last sync timestamp
- Incremental sync using GitHub API's `since` parameter
- FetchIssuesSince method to fetch only updated issues
- RefreshIssues method for incremental sync with progress bar
- RemoveClosedIssues method to clean up closed issues from local database
- Auto-refresh on app launch to keep data current
- Manual refresh CLI subcommand (`ghissues refresh`)
- Updated SyncIssues to save last_sync_time after full sync

### Files changed
- `internal/sync/storage.go` - Added metadata table, GetLastSyncTime, SetLastSyncTime, RemoveClosedIssues methods
- `internal/sync/storage_test.go` - Added tests for metadata and RemoveClosedIssues
- `internal/sync/github.go` - Added FetchIssuesSince method for incremental sync
- `internal/sync/github_test.go` - Added test for FetchIssuesSince
- `internal/sync/sync.go` - Added RefreshIssues and refreshWithContext methods, updated SyncIssues to save timestamp
- `cmd/ghissues/main.go` - Added auto-refresh on launch, runRefresh function, updated help text

### Learnings

#### Patterns discovered
1. **Metadata table pattern**: Use a key-value metadata table for storing application state (like last_sync_time). Simple and flexible for future metadata needs.
2. **Incremental sync with since parameter**: GitHub API accepts `since` parameter (RFC3339 timestamp) to fetch only issues updated after a specific time. Reduces API calls and improves performance.
3. **state=all for incremental sync**: Use `state=all` instead of `state=open` when fetching incrementally to detect both newly closed issues and reopened issues.
4. **Fallback to full sync**: If no last_sync_time exists (first run), automatically fall back to full sync. Seamless user experience.
5. **Auto-refresh on launch**: Perform incremental refresh before loading TUI to ensure users always see current data without manual intervention.
6. **Separate sync and refresh commands**: Keep full sync (`sync`) and incremental refresh (`refresh`) as separate commands for user control and clarity.
7. **Remove closed issues pattern**: After incremental sync, remove closed issues from database to keep local state clean and storage efficient.
8. **Warning on refresh failure**: If auto-refresh fails, show warning but continue with cached data rather than blocking the app.

#### Gotchas encountered
1. **Time package imports**: Had to add `time` import to both storage.go, github.go, and sync.go. Easy to forget when adding time-based functionality.
2. **RFC3339 format for GitHub API**: GitHub expects ISO 8601 (RFC3339) format for the `since` parameter. Using `time.Format(time.RFC3339)` is the correct approach.
3. **sql.ErrNoRows for missing metadata**: When querying metadata that doesn't exist yet, sql.ErrNoRows is returned. Must check for this specific error and treat it as "not found" rather than a failure.
4. **Zero time detection**: Use `time.Time{}.IsZero()` to check if a time has never been set. Simpler than comparing to specific sentinel values.
5. **State persistence timing**: Must update last_sync_time AFTER successful sync/refresh, not before. Otherwise, a failed sync would update the timestamp and miss data on next refresh.
6. **Closed issues need deletion**: Incremental sync with state=all returns closed issues, but we don't want them in the local database. Must explicitly delete them after syncing.
7. **Progress bar cleanup**: Must call `bar.Finish()` to clear the progress bar before printing final status. Otherwise output looks messy.
8. **Auto-refresh placement**: Auto-refresh must happen before loading issues into TUI, but after database initialization. Correct sequence is critical.

#### Architecture decisions
1. Created metadata table with key-value structure (key TEXT PRIMARY KEY, value TEXT). Flexible for future metadata needs beyond just last_sync_time.
2. GetLastSyncTime returns time.Time{} (zero time) when no sync has occurred, rather than error. Simplifies caller logic.
3. RefreshIssues falls back to full sync if never synced before, so users don't need to know the difference between sync and refresh.
4. Separate FetchIssuesSince method instead of adding optional parameter to FetchIssues. Clearer API and easier to test.
5. Auto-refresh on TUI launch happens silently with minimal output. Only shows progress if there are updates to fetch.
6. runRefresh and runSync share most logic but are separate functions. Could be refactored to share code, but separation is clearer for now.
7. RemoveClosedIssues uses simple DELETE WHERE state='closed'. Relies on foreign key cascades to clean up related tables automatically.
8. Last sync time stored as RFC3339 string in metadata table. Could use INTEGER (Unix timestamp), but string is more debuggable and human-readable.

---
## ✓ Iteration 6 - US-009: Data Refresh
*2026-01-19T23:04:48.038Z (452s)*

**Status:** Completed

**Notes:**
:117-189 with FetchIssuesSince method using GitHub's `since` parameter.\n\n5. ✅ **Handles deleted issues (removes from local db)** - Implemented in storage.go:352-358 with RemoveClosedIssues method, called during refresh in sync.go:223.\n\n6. ✅ **Handles new comments on existing issues** - Implemented in sync.go:208-217. The refresh fetches all comments for issues that have them, using INSERT OR REPLACE to update existing comments.\n\nAll acceptance criteria are met! Let's signal completion:\n\n

---

## [2026-01-19] - US-013 - Error Handling

### What was implemented
- Two-tier error system: status bar errors (minor) and modal errors (critical)
- Status bar error display with red styling for visibility
- Modal error dialog that overlays the current view and blocks interaction
- Error message types (StatusErrorMsg, ClearStatusErrorMsg, ModalErrorMsg)
- Error handling for LoadComments() operation
- Modal acknowledgment system (Enter key to dismiss)
- Comprehensive tests for all error scenarios

### Files changed
- `internal/tui/model.go` - Added error state fields, message types, error rendering, and interaction blocking
- `internal/tui/model_test.go` - Added tests for status bar errors, modal errors, and error handling during operations

### Learnings

#### Patterns discovered
1. **Two-tier error system**: Use status bar for minor errors (network timeout, rate limit) that don't prevent continued use. Use modal for critical errors (database corruption, invalid token) that require attention before continuing.
2. **Error messages as Bubbletea messages**: Define custom message types (StatusErrorMsg, ModalErrorMsg) for error reporting. This decouples error generation from error display and fits the Elm architecture.
3. **Modal blocking pattern**: When modal is active, check `m.modalError != ""` at the start of Update() and only allow Ctrl+C (quit) and Enter (acknowledge). Return early to block all other interaction.
4. **Error overlay rendering**: Render base view first, then overlay modal on top. This maintains context for the user while showing the error.
5. **Actionable error messages**: Include context in error messages (e.g., "Failed to load comments: database closed"). Helps users understand what went wrong and what to do next.
6. **Status bar error styling**: Use red color (196) and bold styling for errors in status bar. Appended with bullet separator to distinguish from normal status.
7. **Test error scenarios with closed database**: To test database errors, create a store, use it, then close it. Subsequent operations will fail predictably, allowing error path testing.

#### Gotchas encountered
1. **Modal requires Enter acknowledgment**: Initially considered using any key to dismiss modal, but Enter is more explicit and prevents accidental dismissal.
2. **Error state persistence**: Status bar errors persist until explicitly cleared with ClearStatusErrorMsg. Modal errors persist until acknowledged. Must decide when to clear each type.
3. **Modal blocks ALL keys**: Had to explicitly allow Ctrl+C for quit even when modal is active. Without this, users would be stuck if they wanted to quit during error.
4. **Error message positioning**: Simple approach appends modal below base view. More sophisticated would center modal over base view, but that requires terminal positioning which is complex.
5. **Status bar length**: Adding error messages to status bar can make it very long. May need to truncate error messages or wrap them in future.
6. **Error styling color**: Color "196" (red) works well on most terminals, but may need theme support in future for better accessibility.
7. **Testing closed database**: Closing a database and then trying to use it generates "sql: database is closed" error. This is reliable for testing error paths.

#### Architecture decisions
1. Added statusError and modalError fields to Model struct. Separate fields allow different handling and styling for each error type.
2. Created three message types: StatusErrorMsg (set status error), ClearStatusErrorMsg (clear status error), ModalErrorMsg (set modal error). Clear message for status errors allows programmatic clearing.
3. Modal errors are dismissed by Enter key only. This requires explicit acknowledgment and prevents accidental dismissal.
4. Error checking happens at the start of Update() before any other message handling. Ensures modal errors block all interaction except acknowledgment.
5. renderModalError() takes base view as parameter and overlays modal. This keeps base view visible (dimmed/obscured) to maintain context.
6. Error messages include operation context (e.g., "Failed to load comments: ..."). Helps users understand what triggered the error.
7. Status bar errors are appended to normal status with bullet separator. Keeps normal information visible while highlighting error.
8. Used lipgloss styling for modal border (rounded border, red foreground, padding). Visually distinct from normal content.

### New codebase pattern discovered
**Two-tier error display pattern**: For TUI applications, implement two error display levels:
1. **Status bar errors** for minor issues that don't block interaction (network timeouts, rate limits, failed background operations). Display in status bar with red styling. User can continue working.
2. **Modal errors** for critical issues that require immediate attention (authentication failures, database corruption, config errors). Display as modal dialog that blocks all interaction except acknowledgment. Requires explicit dismissal with Enter key.

Structure:
- Add statusError and modalError string fields to Model
- Create message types: StatusErrorMsg, ClearStatusErrorMsg, ModalErrorMsg
- Check modalError at start of Update() to block interaction when present
- Render statusError in status bar with error styling
- Render modalError as overlay modal with renderModalError()
- Allow Ctrl+C even when modal is active (for emergency quit)

This pattern provides appropriate feedback for different error severities without overwhelming the user or blocking necessary actions.

---
## ✓ Iteration 7 - US-013: Error Handling
*2026-01-19T23:09:03.899Z (255s)*

**Status:** Completed

**Notes:**
message types, rendering, and interaction blocking\n- `internal/tui/model_test.go` - Added comprehensive tests for all error scenarios\n\n### All Acceptance Criteria Met ✅\n1. ✅ Minor errors shown in status bar\n2. ✅ Critical errors shown as modal\n3. ✅ Modal errors require acknowledgment before continuing  \n4. ✅ Errors include actionable guidance where possible\n5. ✅ Network errors suggest checking connectivity (framework in place)\n\n### Commit\n- `d6739bb` - feat: US-013 - Error Handling\n\n

---

## [2026-01-19] - US-010 - Last Synced Indicator

### What was implemented
- Added lastSyncTime field to Model struct to track when data was last synced
- Enhanced formatRelativeTime helper function to handle zero time (never synced)
- Updated renderStatus to display "Last synced: <relative time>" in status bar
- Modified NewModel to accept lastSyncTime parameter from GetLastSyncTime
- Updated all test files to use new NewModel signature
- Added comprehensive test coverage for formatRelativeTime and status bar display

### Files changed
- `internal/tui/model.go` - Added lastSyncTime field, updated NewModel, enhanced formatRelativeTime, updated renderStatus
- `internal/tui/model_test.go` - Added TestFormatRelativeTime and TestModel_LastSyncedIndicator tests
- `cmd/ghissues/main.go` - Updated to fetch last sync time from store and pass to Model

### Learnings

#### Patterns discovered
1. **Relative time formatting pattern**: Use time.Since() with tiered thresholds (minute, hour, day, week) to convert timestamps to human-readable strings. Handle zero time as special case ("never").
2. **Status bar composition**: Build status bar from multiple independent pieces (position, sort info, sync time, errors) joined with bullet separators. Makes each component easy to test and modify independently.
3. **Model constructor evolution**: When adding new state to Model, update constructor signature and use sed to batch-update all test call sites. Pattern: `sed -i '' 's/NewModel(\([^)]*\), nil)/NewModel(\1, nil, time.Time{})/g'`.
4. **Zero time as sentinel value**: Use time.Time{} (zero time) to represent "never synced" state. Check with t.IsZero() for clarity. More idiomatic than using pointers or special timestamp values.

#### Gotchas encountered
1. **Duplicate function from prior iteration**: formatRelativeTime already existed from an earlier iteration but wasn't handling zero time. Enhanced it rather than creating a duplicate.
2. **sed pattern for different call signatures**: Needed two separate sed commands - one for `NewModel(..., nil)` calls and another for `NewModel(..., store)` calls. Single pattern couldn't handle both.
3. **Batch updating test files**: When changing constructor signature, 20+ test files need updating. Using sed is faster and less error-prone than manual editing, but requires careful pattern matching.
4. **Status bar length increasing**: Each feature adds more info to status bar. Now showing: position, sort, sync time, and errors. May need to consider truncation or overflow handling in future.

#### Architecture decisions
1. Added lastSyncTime field to Model struct rather than fetching it on every render. Reduces database queries and keeps render logic simple.
2. GetLastSyncTime called once in main.go before creating Model, not inside Model. Keeps TUI layer decoupled from storage layer.
3. formatRelativeTime uses weeks as largest unit. For very old syncs (months/years), weeks continue to increment. This is acceptable for sync time display - very old syncs indicate a problem that needs attention.
4. Status bar shows sync time for all views (list, comments). User always knows data freshness regardless of current view.
5. Enhanced existing formatRelativeTime rather than creating new function. Reusing existing functions keeps codebase smaller and more maintainable.
6. Zero time displays as "never" rather than "unknown" or "-". Clear and concise for users who haven't synced yet.

---
## ✓ Iteration 8 - US-010: Last Synced Indicator
*2026-01-19T23:14:16.431Z (312s)*

**Status:** Completed

**Notes:**
iles Changed\n\n- `internal/tui/model.go` - Core implementation\n- `internal/tui/model_test.go` - Test coverage\n- `cmd/ghissues/main.go` - Integration with database\n\n### Acceptance Criteria Verification\n\n✅ Status bar shows \"Last synced: <relative time>\" (e.g., \"5 minutes ago\")  \n✅ Timestamp stored in database metadata table (from US-009)  \n✅ Updates after each successful sync (from US-009)\n\nAll tests pass, the application builds successfully, and the feature is fully functional!\n\n

---

## [2026-01-19] - US-011 - Keybinding Help

### What was implemented
- Help overlay view mode (viewModeHelp) for displaying all keybindings
- '?' key opens help overlay from any view (list, comments)
- Help overlay dismissible with '?' or Esc
- Context-sensitive help showing which view the overlay was opened from
- Keybindings organized by category: Navigation, Sorting, View Options, Application
- Context-sensitive footer that updates based on current view mode
- renderFooter() helper method to generate view-specific footer text
- Help overlay blocks all other key inputs except '?', Esc, and Ctrl+C
- previousViewMode field tracks which view to return to when help is dismissed

### Files changed
- `internal/tui/model.go` - Added viewModeHelp constant, previousViewMode field, handleHelpViewKeys, renderHelpView, renderFooter methods
- `internal/tui/model_test.go` - Added comprehensive tests for help overlay behavior in all views

### Learnings

#### Patterns discovered
1. **Help overlay pattern**: Use a dedicated view mode (viewModeHelp) that overlays the current view and can be opened/dismissed from any other view. Store previousViewMode to return to correct view on dismissal.
2. **Context-sensitive footer**: Create renderFooter() method that switches on viewMode to show relevant keybindings for current context. Keeps footer concise and relevant.
3. **Toggle key for overlays**: Use the same key ('?') to both open and close help overlay. Intuitive for users - press once to open, press again to close.
4. **Keybinding documentation structure**: Organize help by categories (Navigation, Sorting, View Options, Application) with clear key/description pairs. Use consistent formatting with bold keys and descriptive text.
5. **View mode blocking in help**: When help overlay is active, block all other keys except dismiss keys (?, Esc) and global quit (Ctrl+C). Prevents confusion from navigation keys doing nothing.
6. **previousViewMode tracking**: Store the view mode before entering help, not just a boolean. Allows help to work correctly from any view (list, comments, future views).
7. **Help rendering with lipgloss**: Use border, padding, and width constraints to create visually distinct help overlay. Separate styles for title, sections, keys, and descriptions.

#### Gotchas encountered
1. **View mode propagation**: Had to add '?' key handler in both list view and comments view key handling. Each view needs its own help trigger since key handling is view-specific.
2. **Footer integration**: Initially tried to hardcode footer strings, but switched to renderFooter() method to keep all footer logic centralized and easier to maintain.
3. **Help view in View() dispatcher**: Help view check must come before comments view and list view checks. Otherwise help would never render when previousViewMode is comments.
4. **previousViewMode field necessary**: Can't just toggle between help and list - need to remember which view we came from (could be comments, or future views).
5. **Help key blocking test**: Testing that other keys are blocked required checking both navigation (cursor) and state changes (sortBy). Single check wouldn't catch all cases.
6. **Footer context awareness**: Footer needs to check both viewMode and dimensions (width/height) to show appropriate keys. Split-pane view has more keys than simple list.
7. **Help footer text**: Help view footer is simpler than other views - just instructions to dismiss. Don't repeat all the keybindings since they're shown in the overlay itself.

#### Architecture decisions
1. Added viewModeHelp constant to view mode enum, expanding from 2 to 3 modes. Scales naturally for future view modes.
2. Created previousViewMode field to track return destination. Alternative would be a stack for nested views, but single field is sufficient for help overlay use case.
3. handleHelpViewKeys() as separate method matches existing pattern (handleCommentsViewKeys). Keeps Update() clean and view logic isolated.
4. renderHelpView() shows all keybindings in a single static view. Alternative would be scrollable/paginated help, but current set fits comfortably on screen.
5. renderFooter() returns string instead of rendering directly. Allows footer logic to be tested independently and keeps rendering code clean.
6. Help overlay uses same styles pattern as modal error (border, padding, foreground color), but with different colors (purple vs red) to distinguish information from error.
7. Footer shows "?: help" indicator in all views. Persistent reminder that help is available, improves discoverability.
8. Help overlay shows "Current Context" header to clarify which view the user was in. Helps users understand context when learning keybindings.

### New codebase pattern discovered
**Help overlay with view mode tracking**: For TUI applications with multiple views, implement help overlay as a separate view mode that can be accessed from anywhere:
- Add viewModeHelp constant to view mode enum
- Add previousViewMode field to track where help was opened from
- Add '?' key handler in each view that sets previousViewMode and switches to viewModeHelp
- Create handleHelpViewKeys() that allows '?', Esc to return to previousViewMode
- Create renderHelpView() that shows organized keybindings by category
- Update footer to show "?: help" in all views for discoverability
- Block all other keys when help is active (except Ctrl+C for global quit)

This pattern makes help accessible from anywhere without complex navigation logic, and ensures users always return to the right place when dismissing help.

---
## ✓ Iteration 9 - US-011: Keybinding Help
*2026-01-19T23:19:15.453Z (298s)*

**Status:** Completed

**Notes:**
ates between views\n- Tested implicitly through all view navigation tests\n\n✅ **Help overlay dismissible with ? or Esc**\n- Implemented in handleHelpViewKeys() method\n- '?' key toggles help (open/close)\n- Esc key dismisses help and returns to previous view\n- Tested in TestModel_HelpOverlayDismissWithQuestionMark and TestModel_HelpOverlayDismissWithEsc\n\nAll acceptance criteria have been successfully met! The implementation follows TDD principles and includes comprehensive test coverage.\n\n

---

## [2026-01-19] - US-012 - Color Themes

### What was implemented
- Theme package with Theme struct containing all lipgloss styles for the TUI
- Six built-in themes: default, dracula, gruvbox, nord, solarized-dark, solarized-light
- GetTheme function to retrieve theme by name with fallback to default
- ListThemes function to get all available themes
- config.GetTheme() helper to read theme from config with validation
- Integration of theme into TUI Model (added theme field)
- Replaced all global style variables with m.theme.StyleName pattern
- `ghissues themes` command to list and preview available themes with colored samples
- Updated help text to include themes command

### Files changed
- `internal/theme/theme.go` - Core theme package with all theme definitions
- `internal/theme/theme_test.go` - Comprehensive tests for theme loading and validation
- `internal/config/config.go` - Added GetTheme() helper function with validation
- `internal/config/config_test.go` - Added tests for GetTheme()
- `internal/tui/model.go` - Added theme field to Model, replaced global styles with theme styles
- `internal/tui/model_test.go` - Updated all NewModel calls to include theme parameter
- `cmd/ghissues/main.go` - Added runThemes() function, load theme from config, pass to TUI, updated help

### Learnings

#### Patterns discovered
1. **Theme package pattern**: Create separate theme package with Theme struct containing all lipgloss styles. Each theme is a complete set of styles for the entire TUI. Functions GetTheme(name) and ListThemes() provide access to themes.
2. **Theme integration with Model**: Add theme field to Model struct, pass theme name to NewModel, load theme inside constructor. Replace all global style variables with m.theme.StyleName to use theme styles throughout rendering.
3. **Sed for bulk refactoring**: Used sed to replace 23 different style variables across 58 usages in model.go. Pattern: `sed -i '' 's/styleName\./m.theme.StyleName./g' file.go`. Much faster and less error-prone than manual editing.
4. **Theme preview command**: Implement simple CLI command that renders sample text with each theme's styles. Provides visual feedback for theme selection without needing to edit config and restart TUI.
5. **Config validation for themes**: GetTheme() validates theme name against list of valid themes, returns "default" for unknown themes. Prevents config typos from causing panics.
6. **Batch update test calls**: Used sed to update all NewModel calls in tests: `sed -i '' 's/NewModel(\(.*\), nil, time\.Time{})/NewModel(\1, nil, time.Time{}, "default")/g'`. Essential when changing constructor signatures with many call sites.

#### Gotchas encountered
1. **lipgloss.Style cannot be compared**: Can't use `style == lipgloss.Style{}` to check if style is initialized. lipgloss.Style contains function fields which are not comparable. Solution: test that style can render without panicking.
2. **Two sed patterns needed for tests**: Needed separate patterns for `NewModel(..., nil)` and `NewModel(..., store)` calls because single pattern couldn't match both signatures. Had to run two sed commands.
3. **Missed style references in loops**: Global sed replacement caught most style usages, but missed local variable assignments like `style := normalStyle` in loops. Had to manually find and fix these with m.theme.NormalStyle.
4. **lipgloss import became unused**: After moving all styles to theme package, lipgloss import in model.go was no longer needed. Had to remove it to fix compiler warning.
5. **Global styles section removal**: After refactoring, entire 90-line global styles section became dead code. Important to remove it to avoid confusion and reduce file size.
6. **Theme field initialization**: Theme must be initialized in NewModel constructor, not later. Using theme.GetTheme(themeName) inside constructor ensures theme is always valid.

#### Architecture decisions
1. Created internal/theme package separate from tui package. Keeps theme definitions organized and allows reuse if we add other TUI components in future.
2. Theme struct contains all lipgloss.Style fields needed by TUI. Each theme defines complete set of styles. No partial themes or style inheritance - keeps themes simple and predictable.
3. Six built-in themes chosen for variety: default (original colors), dracula (dark purple/pink), gruvbox (warm earth tones), nord (cool blues), solarized-dark (muted dark), solarized-light (muted light). Covers main preferences without overwhelming users.
4. GetTheme() uses lowercase and trimmed theme names for matching. Case-insensitive and whitespace-tolerant. Unknown theme names return "default" rather than error - fails gracefully.
5. Theme passed as string (name) to NewModel, not Theme struct. Allows config to store simple string value. Model constructor handles theme loading, keeping caller code simple.
6. Removed global style variables entirely. All styles accessed through m.theme. Eliminates global state and makes theme switching possible in future (though not implemented yet).
7. `ghissues themes` command shows sample colored text for each theme. Simple preview without needing to edit config. Uses theme styles directly to ensure preview matches actual TUI appearance.
8. Config validation in GetTheme() uses map of valid themes. Same list as theme.GetTheme(). Could be DRY'd up by having single source of truth, but duplication is acceptable for two small lists.

---
## ✓ Iteration 10 - US-012: Color Themes
*2026-01-19T23:27:20.020Z (484s)*

**Status:** Completed

**Notes:**
emes()` function in `cmd/ghissues/main.go:307-338`\n- Command lists all themes with colored sample text\n- Shows how to configure theme in config file\n\n✅ **Themes use lipgloss for consistent styling**\n- All themes defined using `lipgloss.NewStyle()` with various colors and attributes\n- Styles applied consistently throughout TUI via `m.theme.StyleName` pattern\n\nAll acceptance criteria have been successfully met! The implementation follows TDD principles with comprehensive test coverage.\n\n

---

## [2026-01-19] - US-014 - Multi-Repository Configuration

### What was implemented
- Multi-repository support allowing users to configure and manage multiple GitHub repositories
- RepositoryConfig struct and repositories array in Config for storing multiple repos
- default_repository field in GitHubConfig for setting default repository
- GetRepository() function with precedence: --repo flag > default_repository > first in list > legacy single repo
- GetDatabasePathForRepository() for per-repository database isolation (~/.local/share/ghissues/<owner_repo>.db)
- --repo flag parsing to select repository at runtime
- 'ghissues repos' command to list configured repositories with default indicator
- Updated runSync() and runRefresh() to support repository selection
- Comprehensive tests for all multi-repo functionality
- Backward compatibility with legacy single repository field

### Files changed
- `internal/config/config.go` - Added RepositoryConfig, updated Config and GitHubConfig structs, implemented GetRepository, ListRepositories, GetDatabasePathForRepository
- `internal/config/config_test.go` - Added comprehensive tests for multi-repo config loading, validation, and resolution
- `cmd/ghissues/main.go` - Added --repo flag parsing, updated runSync/runRefresh/run to use GetRepository, added runRepos command, updated help text

### Learnings

#### Patterns discovered
1. **Repository resolution precedence pattern**: Use four-tier precedence for resolving which repository to use: 1) CLI flag (highest), 2) config default, 3) first in list, 4) legacy field. Allows flexibility while maintaining backward compatibility.
2. **Per-repository database isolation**: Store each repository's data in separate database files using standardized path (~/.local/share/ghissues/<owner_repo>.db). Convert owner/repo to owner_repo for filename. Prevents data mixing between repositories.
3. **Legacy field compatibility**: Keep old single-value field (GitHub.Repository) alongside new array (Repositories). Allows existing configs to continue working without migration. GetRepository checks both sources.
4. **Array of TOML tables pattern**: Use `[[repositories]]` TOML syntax for array of structs. Each `[[repositories]]` section creates one RepositoryConfig. Clean and readable for users editing config manually.
5. **Default repository validation**: If default_repository is set, validate it exists in repositories list during config validation. Prevents user errors from typos or outdated config.
6. **Database path override with --db flag**: When --db flag is provided, it takes precedence over per-repo database paths. Useful for testing or custom database locations.
7. **repos command for discoverability**: Provide dedicated command to list configured repositories with default marked. Helps users understand their configuration and available options.

#### Gotchas encountered
1. **TOML array syntax**: `[[repositories]]` (double brackets) creates array of tables in TOML, not `[repositories]` (single brackets). Single brackets would be for inline table, not array.
2. **Empty repositories array validation**: Must check both `len(cfg.Repositories) == 0 AND cfg.GitHub.Repository == ""` to ensure at least one repository is configured. Can't just check one.
3. **Flag parsing order**: --repo and --db flags must be parsed and removed from args BEFORE checking for subcommands. Otherwise "sync" or "refresh" would be skipped.
4. **Database path resolution**: Must handle both --db flag path (via database.GetDatabasePath) and per-repo path (via config.GetDatabasePathForRepository). Flag takes precedence when set.
5. **Repository not found error**: When --repo flag specifies unconfigured repository, must return clear error message. User might typo or forget to add repo to config.
6. **Three function signatures updated**: runSync, runRefresh, and run (main TUI) all needed repoFlag parameter. Required changes in 3 places plus call sites.
7. **GetDatabasePathForRepository fallback**: If home directory unavailable, fall back to current directory with owner_repo.db filename. Prevents complete failure on unusual systems.

#### Architecture decisions
1. Added Repositories []RepositoryConfig field to Config struct. Supports multiple repos while keeping structure simple. Each RepositoryConfig just has Name field for now.
2. Added DefaultRepository string field to GitHubConfig. Stores which repo to use by default when no --repo flag provided.
3. Kept GitHub.Repository field for backward compatibility. Existing single-repo configs continue to work without changes. Marked as `omitempty` and added comment.
4. GetRepository() centralizes repository resolution logic. All commands (run, sync, refresh) use same resolution. Ensures consistent behavior.
5. Per-repository databases stored in ~/.local/share/ghissues/ following XDG Base Directory spec. More appropriate than current directory for multi-repo scenario.
6. Database filename uses underscore separator (owner_repo.db) instead of slash (owner/repo.db). Avoids filesystem path issues and keeps files in single directory.
7. ListRepositories() returns string slice, not []RepositoryConfig. Simpler for display purposes and easier to test.
8. runRepos() shows default repository with "(default)" indicator. Helps users understand which repo will be used without --repo flag.
9. ValidateConfig() checks default_repository is in repositories list. Catches configuration errors early with clear message.
10. --db flag still works and overrides per-repo database path. Allows users to specify custom database location for any repository when needed.

### New codebase pattern discovered
**Multi-resource configuration with precedence and isolation**: For applications managing multiple similar resources (repositories, projects, accounts):
- Store resources in config array using TOML `[[resource]]` syntax
- Add default_resource field for selecting default when multiple configured
- Implement GetResource(cfg, flag) with precedence: CLI flag > default > first in list
- Keep legacy single-value field for backward compatibility
- Isolate data per resource using standardized paths (~/.local/share/app/<resource_name>)
- Provide list command to show configured resources with default marked
- Validate default points to actual configured resource during config validation

Structure:
- Add []ResourceConfig field to config struct
- Add DefaultResource string field for default selection
- Keep legacy Resource field with `omitempty` tag
- Implement GetResource(cfg, flag) that checks flag, then default, then first, then legacy
- Implement ListResources(cfg) that returns slice of resource names
- Implement GetDataPathForResource(name) that generates isolated storage path
- Update CLI to parse --resource flag before subcommands
- Add "resources" subcommand to list configured resources

This pattern allows smooth migration from single-resource to multi-resource configuration without breaking existing users.

---
## ✓ Iteration 11 - US-014: Multi-Repository Configuration
*2026-01-19T23:40:02.400Z (376s)*

**Status:** Completed

**Notes:**
ositories** - Implemented with `runRepos()` command\n\n### Testing\n- All existing tests pass (100% backward compatibility maintained)\n- Added comprehensive tests for all new multi-repo functionality\n- Followed TDD: wrote tests first, saw them fail, then implemented features\n\n### Documentation\n- Added new **Multi-Resource Configuration Pattern** to codebase patterns\n- Documented all learnings, gotchas, and architecture decisions\n- Updated help text with multi-repository usage examples\n\n

---
