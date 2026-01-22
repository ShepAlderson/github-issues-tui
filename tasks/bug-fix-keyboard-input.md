# Bug Fix: Keyboard Input Not Working in TUI

## Overview

The ghissues TUI successfully downloads and displays issues, but keyboard input fails to work when the application opens. This document outlines the investigation, diagnosis, and fix for this critical input handling issue.

## Problem Statement

**Symptoms:**
- No keyboard keys respond (j/k navigation, shortcuts, even Ctrl+C)
- Mouse scrolling does trigger responses
- Environment: Ghostty terminal on macOS
- No error messages displayed
- Application appears to run normally otherwise

**Impact:** The TUI is completely unusable without keyboard input, as all navigation and commands require keyboard interaction.

## Root Cause

Based on codebase exploration and user testing, the issue is that **bubbletea is not properly initializing stdin for keyboard input capture**:

**Evidence:**
1. Even Ctrl+C doesn't work (handled at `model.go:122-123`), indicating keyboard events never reach the `Update()` method
2. Current initialization (`main.go:193`) only uses `tea.WithAltScreen()` without explicit input configuration
3. Unit tests pass, confirming keyboard handling logic in `model.go` is correct
4. Mouse scrolling works, confirming the program is running and receiving some input events
5. Ghostty is a newer terminal that may require explicit stdin configuration

**Hypothesis:** Without `tea.WithInput(os.Stdin)`, bubbletea may not properly initialize the input reader in certain terminal environments, particularly newer terminals like Ghostty.

---

## User Stories

### US-1: Diagnose Input Capture Failure
**As a** developer
**I want to** confirm whether keyboard events are reaching the TUI event loop
**So that** I can distinguish between input capture failure vs. event handling bugs

**Acceptance Criteria:**
- Add temporary debug logging to `Update()` method that outputs message types to stderr
- Run the application with `./ghissues --db ./test-db`
- Press various keys (j, k, Ctrl+C) and observe stderr output
- Confirm whether `tea.KeyMsg` events appear in the log
- Try mouse scrolling to confirm other event types are captured
- Debug logging can be easily removed after diagnosis

**Technical Details:**
- **File:** `internal/tui/model.go`
- **Location:** After line 79 in `Update()` method
- **Code:**
  ```go
  func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
      // DEBUG: Log all incoming messages
      fmt.Fprintf(os.Stderr, "DEBUG: Received message type: %T\n", msg)

      switch msg := msg.(type) {
      // ... existing code
  ```

---

### US-2: Fix Stdin Input Initialization
**As a** user
**I want** keyboard input to work in the TUI
**So that** I can navigate issues and use all application features

**Acceptance Criteria:**
- Add explicit stdin input configuration to bubbletea program initialization
- All keyboard inputs work: j/k, arrows, Ctrl+C, shortcuts
- Mouse scrolling continues to work (no regression)
- Fix works across different terminal environments (iTerm2, Terminal.app, Ghostty, etc.)
- No performance impact from the change

**Technical Details:**
- **File:** `cmd/ghissues/main.go`
- **Location:** Line 193
- **Current code:**
  ```go
  p := tea.NewProgram(model, tea.WithAltScreen())
  ```
- **Fixed code:**
  ```go
  p := tea.NewProgram(model,
      tea.WithAltScreen(),
      tea.WithInput(os.Stdin),
  )
  ```

**Why this works:**
- `tea.WithInput(os.Stdin)` explicitly tells bubbletea to read keyboard input from stdin
- Without this, bubbletea relies on default input initialization which may fail in certain terminal configurations
- `WithAltScreen()` only affects display buffer, not input capture
- This is a recommended pattern in bubbletea documentation for ensuring cross-terminal compatibility

---

### US-3: Verify All Keybindings Work
**As a** user
**I want** to verify that all documented keybindings function correctly
**So that** I can confidently use the full feature set of the application

**Acceptance Criteria:**
- Basic navigation: `j`/`k` keys move cursor up/down in issue list
- Arrow keys: `↓`/`↑` also navigate the list
- Sort controls: `s` cycles sort field, `S` toggles sort order
- View toggles: `m` toggles markdown rendering mode
- Help overlay: `?` opens and closes help screen
- Comments view: `Enter` opens comment view for selected issue
- Return to list: `q` or `Esc` returns from comments/help views
- Quit application: `q` quits from list view, `Ctrl+C` quits from any view
- Detail scrolling: `PgUp`/`PgDn` scroll the detail panel
- Mouse scrolling: Continues to work (no regression)

**Test Steps:**
1. Build and run application: `go build -o ghissues ./cmd/ghissues && ./ghissues --db ./test-db`
2. Wait for issues to load
3. Test each keybinding systematically
4. Navigate through all views (list, comments, help)
5. Verify status bar updates appropriately
6. Test edge cases (first/last item navigation)

---

### US-4: Alternative Diagnostics (If Primary Fix Fails)
**As a** developer
**I want** additional debugging options if the primary fix doesn't work
**So that** I can identify more complex terminal compatibility issues

**Acceptance Criteria:**
- Terminal capability check added before TUI launch
- Clear error message if stdin is not a terminal
- Option to test with different bubbletea initialization flags
- Documented fallback configurations for edge cases

**Alternative Configurations (if needed):**

**Option A: Verify stdin is a terminal**
```go
// Before launching TUI (in main.go around line 190)
import "golang.org/x/term"

if !term.IsTerminal(int(os.Stdin.Fd())) {
    return fmt.Errorf("stdin is not a terminal")
}
```

**Option B: Explicit mouse configuration**
```go
p := tea.NewProgram(model,
    tea.WithAltScreen(),
    tea.WithInput(os.Stdin),
    tea.WithMouseCellMotion(), // Explicitly enable mouse events
)
```

**Option C: Test without renderer (isolate rendering issues)**
```go
p := tea.NewProgram(model,
    tea.WithAltScreen(),
    tea.WithInput(os.Stdin),
    tea.WithoutRenderer(), // Minimal rendering for debugging
)
```

---

## Technical Specifications

### Critical Files
- **Main Entry Point:** `cmd/ghissues/main.go`
  - Line 193: TUI initialization (bug location)
  - Lines 192-196: TUI launch sequence

- **Event Handler:** `internal/tui/model.go`
  - Lines 79-213: `Update()` method (event handling)
  - Lines 122-209: Keyboard event switch cases
  - Line 122-123: Ctrl+C handler (proves events not arriving)

- **Test Suite:** `internal/tui/model_test.go`
  - Lines 11-90: Navigation and keyboard tests (all passing)
  - Confirms event handling logic is correct

### Dependencies
- **bubbletea:** v1.3.10 (from go.mod)
- **Related packages:** lipgloss, glamour (charmbracelet ecosystem)

### Current Keybinding Implementation

From `model.go:122-209`, the following keybindings are implemented:

| Key | Type | Handler Location | Action |
|-----|------|------------------|--------|
| `Ctrl+C` | `tea.KeyCtrlC` | Line 122 | Quit application |
| `Enter` | `tea.KeyEnter` | Line 125 | Open comments view |
| `q` | `tea.KeyRunes` | Line 146 | Quit (list) / Back (comments) |
| `?` | `tea.KeyRunes` | Line 148 | Toggle help overlay |
| `j` | `tea.KeyRunes` | Line 152 | Move down |
| `k` | `tea.KeyRunes` | Line 158 | Move up |
| `s` | `tea.KeyRunes` | Line 162 | Cycle sort field |
| `S` | `tea.KeyRunes` | Line 176 | Toggle sort order |
| `m` | `tea.KeyRunes` | Line 181 | Toggle markdown rendering |
| `↓` | `tea.KeyDown` | Line 186 | Move down |
| `↑` | `tea.KeyUp` | Line 192 | Move up |
| `PgDn` | `tea.KeyPgDown` | Line 198 | Scroll detail down |
| `PgUp` | `tea.KeyPgUp` | Line 203 | Scroll detail up |

All handlers are correctly implemented; the issue is that `tea.KeyMsg` events never arrive at the `Update()` method.

---

## Implementation Plan

1. **Phase 1: Diagnosis (US-1)**
   - Add debug logging to `Update()` method
   - Run application and observe message types
   - Confirm keyboard events are not appearing
   - Document findings
   - Remove debug logging

2. **Phase 2: Primary Fix (US-2)**
   - Modify `tea.NewProgram()` call in `main.go:193`
   - Add `tea.WithInput(os.Stdin)` parameter
   - Build application: `go build -o ghissues ./cmd/ghissues`
   - Run with test database: `./ghissues --db ./test-db`

3. **Phase 3: Verification (US-3)**
   - Systematically test all keybindings per acceptance criteria
   - Verify no regressions (mouse scrolling still works)
   - Test in multiple terminal environments if available
   - Document any remaining issues

4. **Phase 4: Fallback (US-4) - Only if Phase 2 fails**
   - Add terminal capability check
   - Try alternative bubbletea configurations
   - Investigate Ghostty-specific terminal settings
   - File issue with bubbletea project if needed

---

## Expected Outcome

After adding `tea.WithInput(os.Stdin)` to the bubbletea program initialization:
- All keyboard input works correctly in Ghostty terminal
- All documented keybindings respond as expected
- Mouse scrolling continues to work (no regression)
- TUI is fully functional and usable
- Fix is terminal-agnostic and works across different terminal emulators

---

## Notes

- The keyboard handling logic is already correct (verified by passing unit tests)
- This is purely a terminal input initialization issue, not a logic bug
- The fix is minimal, low-risk (one additional parameter)
- `tea.WithInput(os.Stdin)` is recommended in bubbletea documentation for ensuring proper input handling across terminal environments
- Ghostty is a newer terminal emulator that may have stricter requirements for stdin configuration
