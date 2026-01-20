# Ralph Progress Log

This file tracks progress across iterations. It's automatically updated
after each iteration and included in agent prompts for context.

## Codebase Patterns (Study These First)

### TDD Approach
- Write failing tests first, then implement code to make them pass
- Test interactive prompts with buffer-based I/O for deterministic tests
- Use table-driven tests for comprehensive coverage of edge cases

### Package Structure
```
/
├── main.go                 # CLI entry point
├── main_test.go           # Integration tests
└── internal/
    ├── cmd/               # Command implementations
    │   ├── config.go      # Config command
    │   └── config_test.go
    ├── config/            # Configuration logic
    │   ├── config.go      # Load/save config
    │   └── config_test.go
    └── prompt/            # Interactive prompts
        ├── prompt.go      # User input handling
        └── prompt_test.go
```

### Input Handling Pattern
When working with interactive prompts and `bufio.Reader`:
- Create a SINGLE `bufio.NewReader(input)` in the entry function
- Pass the `*bufio.Reader` to all prompt functions
- This ensures the reader state is maintained across multiple reads
- Multiple `bufio.NewReader()` on the same input will cause EOF errors after the first read

---

## [2026-01-20] - US-001 - First-Time Setup

**What was implemented:**
- Interactive configuration setup for first-time users
- Config command (`ghissues config`) to re-run setup
- Configuration file management (`~/.config/ghissues/config.toml`)
- Automatic setup prompting when no config exists
- Support for multiple authentication methods

**Files changed:**
- Created `main.go` - CLI entry point with command routing
- Created `main_test.go` - Integration tests for main flow
- Created `internal/cmd/config.go` - Config command implementation
- Created `internal/cmd/config_test.go` - Config command tests
- Created `internal/config/config.go` - Config file I/O operations
- Created `internal/config/config_test.go` - Config tests
- Created `internal/prompt/prompt.go` - Interactive prompt handling
- Created `internal/prompt/prompt_test.go` - Prompt tests
- Created `internal/prompt/integration_test.go` - Integration tests
- Created `go.mod` - Go module initialization with go-toml dependency
- Updated `.ralph-tui/progress.md` - Documented patterns and learnings

**Learnings:**
- **Pattern Discovered:** When using `bufio.NewReader` with the same `io.Reader`, create it once and reuse it. Creating multiple `bufio.NewReader` instances on the same reader causes state issues and EOF errors.
- **Gotcha Encountered:** Test input for interactive prompts must include newlines (`\n`) to simulate user pressing Enter. Without them, `ReadString('\n')` will hang or return EOF.
- **TDD Benefit:** Writing tests first forced cleaner API design and revealed edge cases early (e.g., invalid repository format, empty tokens).
- **Package Layout:** The `internal/` package structure keeps implementation details private while allowing focused unit tests for each component.

**Testing Notes:**
- All tests use buffer-based I/O (bytes.Buffer, strings.Reader) for deterministic behavior
- Integration tests verify end-to-end flows
- Test coverage includes success paths, error paths, and edge cases
- Verified with: `go test ./...`, `golangci-lint run`, `go vet`, `gofmt`

---
