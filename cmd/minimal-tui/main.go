package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Minimal TUI for testing keyboard input
// This is the simplest possible bubbletea program to isolate input issues

type model struct {
	lastKey     string
	keyCount    int
	messages    []string
	width       int
	height      int
}

func initialModel() model {
	return model{
		lastKey:  "none",
		keyCount: 0,
		messages: []string{"Press any key to test keyboard input", "Press 'q' or Ctrl+C to quit"},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Log all messages to stderr for debugging
	fmt.Fprintf(os.Stderr, "DEBUG: msg type=%T value=%+v\n", msg, msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.messages = append(m.messages, fmt.Sprintf("WindowSizeMsg: %dx%d", msg.Width, msg.Height))
		return m, nil

	case tea.KeyMsg:
		m.keyCount++
		m.lastKey = msg.String()

		// Quit on q or Ctrl+C
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyRunes:
			if string(msg.Runes) == "q" {
				return m, tea.Quit
			}
		}

		m.messages = append(m.messages, fmt.Sprintf("KeyMsg #%d: type=%v string=%q runes=%v", m.keyCount, msg.Type, msg.String(), msg.Runes))

		// Keep only last 20 messages
		if len(m.messages) > 20 {
			m.messages = m.messages[len(m.messages)-20:]
		}

		return m, nil

	case tea.MouseMsg:
		m.messages = append(m.messages, fmt.Sprintf("MouseMsg: %+v", msg))
		if len(m.messages) > 20 {
			m.messages = m.messages[len(m.messages)-20:]
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString("=== Minimal TUI Keyboard Test ===\n\n")
	b.WriteString(fmt.Sprintf("Last key: %s\n", m.lastKey))
	b.WriteString(fmt.Sprintf("Key count: %d\n", m.keyCount))
	b.WriteString(fmt.Sprintf("Terminal size: %dx%d\n", m.width, m.height))
	b.WriteString("\n--- Event Log ---\n")

	for _, msg := range m.messages {
		b.WriteString(msg + "\n")
	}

	b.WriteString("\n[Press 'q' or Ctrl+C to quit]\n")

	return b.String()
}

func main() {
	// Parse environment variable for options
	tuiOpts := os.Getenv("GHISSUES_TUI_OPTIONS")

	options := []tea.ProgramOption{
		tea.WithInput(os.Stdin),
	}

	useAltScreen := true
	if tuiOpts != "" {
		flags := strings.Split(tuiOpts, ",")
		for _, flag := range flags {
			flag = strings.TrimSpace(flag)
			switch flag {
			case "mouse":
				options = append(options, tea.WithMouseCellMotion())
			case "noaltscreen":
				useAltScreen = false
			}
		}
	}

	if useAltScreen {
		options = append(options, tea.WithAltScreen())
	}

	fmt.Fprintln(os.Stderr, "DEBUG: Starting minimal TUI...")
	fmt.Fprintln(os.Stderr, "DEBUG: Options:", tuiOpts)

	p := tea.NewProgram(initialModel(), options...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
