# Troubleshooting Guide

This guide covers common issues and edge cases when running ghissues TUI.

## Terminal Compatibility Issues

### Error: "stdin is not a terminal"

**Symptom:** The application exits with an error message stating that stdin is not a terminal.

**Cause:** The TUI application requires an interactive terminal environment to capture keyboard input and display the interface. This error occurs when:
- Input is redirected from a file: `ghissues < input.txt`
- Input is piped from another command: `cat data | ghissues`
- Running in a non-interactive environment (some CI/CD systems)

**Solution:**
1. Run ghissues directly in a terminal without input redirection
2. Ensure you're running in an interactive terminal session
3. If running in a script, ensure the script is executed in an interactive shell

**Example of correct usage:**
```bash
# Correct - run directly in terminal
ghissues

# Correct - with flags
ghissues --db ./test-db --repo owner/repo
```

**Example of incorrect usage:**
```bash
# Incorrect - stdin redirected
echo "" | ghissues

# Incorrect - input from file
ghissues < /dev/null
```

## Advanced Debugging Options

### Environment Variable: GHISSUES_TUI_OPTIONS

The `GHISSUES_TUI_OPTIONS` environment variable allows you to customize the TUI initialization for debugging purposes.

#### Available Options

| Option | Description | Use Case |
|--------|-------------|----------|
| `nomouse` | Disable mouse support | Debug mouse-related issues or test keyboard-only mode |
| `noaltscreen` | Disable alternate screen buffer | Keep TUI output in terminal history for debugging |

#### Usage Examples

**Disable alternate screen buffer:**
```bash
# Output stays in terminal after exit (useful for debugging)
GHISSUES_TUI_OPTIONS=noaltscreen ghissues
```

**Disable mouse support:**
```bash
# Keyboard-only mode
GHISSUES_TUI_OPTIONS=nomouse ghissues
```

**Combine multiple options:**
```bash
# Both disabled (comma-separated)
GHISSUES_TUI_OPTIONS=nomouse,noaltscreen ghissues
```

### When to Use Debug Options

- **`noaltscreen`**: Use when you need to inspect the TUI output after the program exits. This is helpful for debugging rendering issues or capturing screenshots of error states.

- **`nomouse`**: Use when testing keyboard-only workflows or when mouse events interfere with debugging.

## Terminal Environment Detection

The application performs the following checks before launching the TUI:

1. **Stdin Terminal Check**: Verifies that standard input is connected to a terminal device
2. **Terminal Capabilities**: Ensures the terminal supports the required features for the TUI

If any check fails, the application will exit with a clear error message explaining the issue and suggesting corrective actions.

## Edge Cases

### Running in Different Terminal Emulators

The application has been tested and works correctly in:
- iTerm2 (macOS)
- Terminal.app (macOS)
- Ghostty (cross-platform)
- Alacritty (cross-platform)
- GNOME Terminal (Linux)
- Windows Terminal (Windows)

**Key Implementation Detail:** The application uses `tea.WithInput(os.Stdin)` in all TUI initializations to ensure keyboard input works reliably across different terminal environments.

### Running in SSH Sessions

Remote SSH sessions are fully supported as long as:
- The SSH connection is interactive (allocated a pseudo-TTY)
- The remote terminal supports ANSI escape sequences

**Correct SSH usage:**
```bash
# Force TTY allocation if needed
ssh -t user@host ghissues
```

### Running in tmux or screen

The application works correctly within tmux or screen multiplexers. No special configuration is required.

### Running in CI/CD Environments

**Not Supported:** TUI applications cannot run in non-interactive CI/CD environments. If you need to automate ghissues operations in CI/CD:

1. Use the command-line interface for syncing:
   ```bash
   ghissues sync --repo owner/repo
   ```

2. Access the database directly (SQLite) for querying issues programmatically

3. Consider building a headless mode or API wrapper if automation is required

## Getting Help

If you encounter issues not covered in this guide:

1. Check the GitHub issues: https://github.com/shepbook/github-issues-tui/issues
2. Enable debug logging by setting `GHISSUES_TUI_OPTIONS=noaltscreen` to capture output
3. Report the issue with:
   - Your terminal emulator and version
   - Operating system
   - Complete error message
   - Steps to reproduce

## Related Documentation

- [README.md](../README.md) - Main documentation
- [Codebase Patterns](.ralph-tui/progress.md) - Implementation patterns and technical details
