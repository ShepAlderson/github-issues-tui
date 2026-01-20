package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RefreshProgressMsg is sent when refresh progress updates
type RefreshProgressMsg struct {
	Current int
	Total   int
	Message string
}

// RefreshCompleteMsg is sent when refresh is complete
type RefreshCompleteMsg struct {
	Success bool
	Error   error
}

// RefreshModel represents the refresh progress state
type RefreshModel struct {
	Active      bool
	Current     int
	Total       int
	Message     string
	StartTime   time.Time
	Success     bool
	Complete    bool
	Width       int
	Height      int
}

// NewRefreshModel creates a new refresh model
func NewRefreshModel() RefreshModel {
	return RefreshModel{
		Active:    false,
		StartTime: time.Now(),
	}
}

// Init initializes the refresh model
func (m RefreshModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the refresh model
func (m RefreshModel) Update(msg tea.Msg) (RefreshModel, tea.Cmd) {
	switch msg := msg.(type) {
	case RefreshProgressMsg:
		m.Active = true
		m.Current = msg.Current
		m.Total = msg.Total
		if msg.Message != "" {
			m.Message = msg.Message
		} else {
			m.Message = fmt.Sprintf("Fetching issues... %d/%d", m.Current, m.Total)
		}
		return m, nil

	case RefreshCompleteMsg:
		m.Active = false
		m.Complete = true
		m.Success = msg.Success
		if msg.Error != nil {
			m.Message = "Refresh failed: " + msg.Error.Error()
		} else if m.Total > 0 {
			m.Message = fmt.Sprintf("✓ Refreshed %d issues", m.Total)
		} else {
			m.Message = "✓ No new issues"
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	}

	return m, nil
}

// View renders the refresh progress
func (m RefreshModel) View() string {
	if !m.Active && !m.Complete {
		return ""
	}

	// Calculate progress bar width
	barWidth := 40
	if m.Width > 0 && m.Width < barWidth+10 {
		barWidth = m.Width - 10
	}

	// Calculate progress
	var progress float64
	if m.Total > 0 {
		progress = float64(m.Current) / float64(m.Total)
	} else if m.Active {
		progress = 0.5 // Indeterminate progress
	}

	// Build progress bar
	filledWidth := int(progress * float64(barWidth))
	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	// Calculate elapsed time
	elapsed := time.Since(m.StartTime)
	elapsedStr := elapsed.Round(time.Second).String()

	// Build the view
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true)

	content := fmt.Sprintf("\n%s\n%s %s (%s)",
		style.Render(m.Message),
		bar,
		fmt.Sprintf("%.0f%%", progress*100),
		elapsedStr,
	)

	return content
}

// IsActive returns whether the refresh is currently active
func (m RefreshModel) IsActive() bool {
	return m.Active
}

// IsComplete returns whether the refresh has completed
func (m RefreshModel) IsComplete() bool {
	return m.Complete
}

// Reset resets the refresh model for a new refresh
func (m *RefreshModel) Reset() {
	m.Active = false
	m.Complete = false
	m.Current = 0
	m.Total = 0
	m.Message = ""
	m.Success = false
	m.StartTime = time.Now()
}
