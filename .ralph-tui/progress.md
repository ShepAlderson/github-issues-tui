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

