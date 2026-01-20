# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### Debug Logging Pattern
- For TUI event debugging, add `fmt.Fprintf(os.Stderr, "DEBUG: Received message type: %T\n", msg)` at the start of the `Update()` method
- This outputs to stderr which can be captured separately from the TUI display
- Remember to import both "fmt" and "os" packages

---

## 2026-01-19 - US-1: Diagnose Input Capture Failure
- **What was implemented:** Added debug logging to the TUI Update() method to diagnose keyboard input capture issues
- **Files changed:**
  - `internal/tui/model.go`: Added debug logging at line 81 and imported "os" package
  - `cmd/test-tui/main.go`: Created minimal test harness for TUI testing (bypasses authentication)
- **Learnings:**
  - The Update() method in bubbletea is the central event handler that receives all messages (keyboard, mouse, window resize, etc.)
  - Debug output to stderr allows monitoring events without interfering with the TUI display
  - Authentication is required for the main app, so a test harness is useful for isolated testing
  - The application expects GITHUB_TOKEN to be set and validates it before launching the TUI
  - Existing test-db has 149 issues cached, allowing offline testing once auth is bypassed
  - **Pattern discovered:** When debugging TUI applications, creating a minimal test harness (cmd/test-tui) that bypasses setup/auth is valuable for quick iteration
  - **Gotcha:** Cannot run interactive TUI tests in non-interactive shells; manual testing required for keyboard input verification

---

