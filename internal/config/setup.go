package config

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SetupStep represents the current step in the setup process
type SetupStep int

const (
	StepRepository SetupStep = iota
	StepAuthMethod
	StepToken
	StepConfirm
	StepComplete
)

// AuthMethod represents the chosen authentication method
type AuthMethod int

const (
	AuthEnvVar AuthMethod = iota
	AuthConfigFile
)

// SetupModel represents the state of the setup wizard
type SetupModel struct {
	step          SetupStep
	repoInput     textinput.Model
	authSelection int
	tokenInput    textinput.Model
	config        *Config
	repoError     string
	tokenError    string
	width         int
	height        int
	done          bool
}

// SetupResult represents the result of the setup process
type SetupResult struct {
	Config *Config
	Error  error
}

var (
	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#5E81AC")).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88C0D0")).
			MarginLeft(4)

	promptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#E5E9F0")).
			MarginLeft(4)

	inputStyle = lipgloss.NewStyle().
			MarginLeft(4)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BF616A")).
			MarginLeft(4)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A3BE8C")).
			MarginLeft(4)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4C566A")).
			MarginLeft(4)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#88C0D0")).
			Bold(true)

	unselectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4C566A"))
)

// NewSetupModel creates a new setup model
func NewSetupModel() SetupModel {
	// Initialize repository input
	repoInput := textinput.New()
	repoInput.Placeholder = "owner/repo"
	repoInput.Focus()
	repoInput.CharLimit = 100
	repoInput.Width = 40

	// Initialize token input
	tokenInput := textinput.New()
	tokenInput.Placeholder = "ghp_..."
	tokenInput.CharLimit = 100
	tokenInput.Width = 50

	return SetupModel{
		step:          StepRepository,
		repoInput:     repoInput,
		authSelection: 0,
		tokenInput:    tokenInput,
		config: &Config{
			Display: DisplayConfig{
				Theme:   "default",
				Columns: []string{"number", "title", "author", "updated", "comments"},
			},
			Sort: SortConfig{
				Field:      "updated",
				Descending: true,
			},
			Database: DatabaseConfig{
				Path: ".ghissues.db",
			},
		},
	}
}

// Init initializes the program
func (m SetupModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages and updates the model
func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEsc:
			if m.step == StepRepository {
				return m, tea.Quit
			}
			// Go back one step
			switch m.step {
			case StepAuthMethod:
				m.step = StepRepository
				m.repoInput.Focus()
			case StepToken:
				if m.authSelection == 0 {
					m.step = StepAuthMethod
				} else {
					m.step = StepAuthMethod
				}
				m.tokenInput.Blur()
			case StepConfirm:
				m.step = StepAuthMethod
			}
			return m, nil

		case tea.KeyEnter:
			switch m.step {
			case StepRepository:
				repo := strings.TrimSpace(m.repoInput.Value())
				if err := ValidateRepository(repo); err != nil {
					m.repoError = err.Error()
					return m, nil
				}
				m.repoError = ""
				m.config.Default.Repository = repo
				owner, name, _ := ParseRepository(repo)
				m.config.Repositories = append(m.config.Repositories, RepositoryConfig{
					Owner:    owner,
					Name:     name,
					Database: ".ghissues.db",
				})
				m.step = StepAuthMethod
				m.repoInput.Blur()
				return m, nil

			case StepAuthMethod:
				if m.authSelection == 1 {
					m.step = StepToken
					m.tokenInput.Focus()
					return m, textinput.Blink
				}
				m.step = StepConfirm
				return m, nil

			case StepToken:
				token := strings.TrimSpace(m.tokenInput.Value())
				if token == "" {
					m.tokenError = "Token cannot be empty"
					return m, nil
				}
				// Basic token format validation
				if !strings.HasPrefix(token, "ghp_") && !strings.HasPrefix(token, "github_pat_") {
					m.tokenError = "Token should start with 'ghp_' or 'github_pat_'"
					return m, nil
				}
				m.tokenError = ""
				m.config.Auth.Token = token
				m.step = StepConfirm
				m.tokenInput.Blur()
				return m, nil

			case StepConfirm:
				// Save the config
				if err := m.config.Save(); err != nil {
					m.tokenError = fmt.Sprintf("Failed to save config: %v", err)
					return m, tea.Quit
				}
				m.done = true
				return m, tea.Quit
			}

		case tea.KeyDown, tea.KeyTab:
			if m.step == StepAuthMethod {
				m.authSelection = (m.authSelection + 1) % 2
			}

		case tea.KeyUp:
			if m.step == StepAuthMethod {
				m.authSelection = (m.authSelection - 1 + 2) % 2
			}
		}
	}

	// Update text inputs
	var cmd tea.Cmd
	switch m.step {
	case StepRepository:
		m.repoInput, cmd = m.repoInput.Update(msg)
	case StepToken:
		m.tokenInput, cmd = m.tokenInput.Update(msg)
	}

	return m, cmd
}

// View renders the setup UI
func (m SetupModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString(titleStyle.Render("✨ ghissues - First-Time Setup"))
	b.WriteString("\n\n")

	switch m.step {
	case StepRepository:
		b.WriteString(subtitleStyle.Render("Step 1/3: Configure Repository"))
		b.WriteString("\n\n")
		b.WriteString(promptStyle.Render("Enter the GitHub repository you want to view issues from:"))
		b.WriteString("\n")
		b.WriteString(inputStyle.Render(m.repoInput.View()))
		b.WriteString("\n")
		if m.repoError != "" {
			b.WriteString(errorStyle.Render("✗ " + m.repoError))
			b.WriteString("\n")
		}
		b.WriteString(infoStyle.Render("Format: owner/repo (e.g., charmbracelet/bubbletea)"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Enter: Continue  •  Esc: Cancel  •  Ctrl+C: Quit"))

	case StepAuthMethod:
		b.WriteString(subtitleStyle.Render("Step 2/3: Authentication Method"))
		b.WriteString("\n\n")
		b.WriteString(promptStyle.Render("How would you like to authenticate with GitHub?"))
		b.WriteString("\n\n")

		// Option 1: Environment variable
		if m.authSelection == 0 {
			b.WriteString(selectedStyle.Render("  ● Use GITHUB_TOKEN environment variable"))
		} else {
			b.WriteString(unselectedStyle.Render("  ○ Use GITHUB_TOKEN environment variable"))
		}
		b.WriteString("\n")
		b.WriteString(unselectedStyle.Render("    Store your token in the GITHUB_TOKEN environment variable"))
		b.WriteString("\n\n")

		// Option 2: Config file
		if m.authSelection == 1 {
			b.WriteString(selectedStyle.Render("  ● Store token in config file"))
		} else {
			b.WriteString(unselectedStyle.Render("  ○ Store token in config file"))
		}
		b.WriteString("\n")
		b.WriteString(unselectedStyle.Render("    The token will be stored securely in ~/.config/ghissues/config.toml"))
		b.WriteString("\n\n")

		b.WriteString(helpStyle.Render("↑/↓: Select  •  Enter: Continue  •  Esc: Back  •  Ctrl+C: Quit"))

	case StepToken:
		b.WriteString(subtitleStyle.Render("Step 2/3: Enter GitHub Token"))
		b.WriteString("\n\n")
		b.WriteString(promptStyle.Render("Enter your GitHub Personal Access Token:"))
		b.WriteString("\n")
		b.WriteString(inputStyle.Render(m.tokenInput.View()))
		b.WriteString("\n")
		if m.tokenError != "" {
			b.WriteString(errorStyle.Render("✗ " + m.tokenError))
			b.WriteString("\n")
		}
		b.WriteString(infoStyle.Render("You can create a token at https://github.com/settings/tokens"))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("Required scopes: repo (for private repos) or public_repo (for public only)"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("Enter: Continue  •  Esc: Back  •  Ctrl+C: Quit"))

	case StepConfirm:
		b.WriteString(subtitleStyle.Render("Step 3/3: Confirm Configuration"))
		b.WriteString("\n\n")

		b.WriteString(promptStyle.Render("Repository:"))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("  " + m.config.Default.Repository))
		b.WriteString("\n\n")

		b.WriteString(promptStyle.Render("Authentication:"))
		b.WriteString("\n")
		if m.authSelection == 0 {
			b.WriteString(infoStyle.Render("  Environment variable (GITHUB_TOKEN)"))
		} else {
			b.WriteString(infoStyle.Render("  Stored in config file"))
			if m.config.Auth.Token != "" {
				b.WriteString("\n")
				maskedToken := m.config.Auth.Token[:4] + "****"
				b.WriteString(unselectedStyle.Render("  Token: " + maskedToken))
			}
		}
		b.WriteString("\n\n")

		b.WriteString(promptStyle.Render("Configuration will be saved to:"))
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("  " + ConfigPath()))
		b.WriteString("\n\n")

		b.WriteString(helpStyle.Render("Enter: Save & Finish  •  Esc: Back  •  Ctrl+C: Quit"))
	}

	return b.String()
}

// RunSetup runs the interactive setup wizard
func RunSetup() (*Config, error) {
	p := tea.NewProgram(NewSetupModel())
	m, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run setup: %w", err)
	}

	model := m.(SetupModel)
	if !model.done {
		return nil, fmt.Errorf("setup cancelled")
	}

	return model.config, nil
}
