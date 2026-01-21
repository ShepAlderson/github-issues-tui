package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config represents the application configuration
type Config struct {
	Repository string         `toml:"repository"`
	Auth       AuthConfig     `toml:"auth"`
	Database   DatabaseConfig `toml:"database"`
	Display    DisplayConfig  `toml:"display"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Path string `toml:"path,omitempty"`
}

// DisplayConfig represents display configuration
type DisplayConfig struct {
	Columns   []string  `toml:"columns,omitempty"`
	SortField SortField `toml:"sort_field,omitempty"`
	SortOrder SortOrder `toml:"sort_order,omitempty"`
	Theme     Theme     `toml:"theme,omitempty"`
}

// ValidColumns contains all valid column names for issue display
var ValidColumns = []string{"number", "title", "author", "date", "comments"}

// SortField represents the field to sort issues by
type SortField string

// Sort field constants
const (
	SortByUpdated  SortField = "updated"
	SortByCreated  SortField = "created"
	SortByNumber   SortField = "number"
	SortByComments SortField = "comments"
)

// ValidSortFields contains all valid sort fields
var ValidSortFields = []SortField{SortByUpdated, SortByCreated, SortByNumber, SortByComments}

// SortOrder represents the sort order (ascending or descending)
type SortOrder string

// Sort order constants
const (
	SortDesc SortOrder = "desc"
	SortAsc  SortOrder = "asc"
)

// Theme represents a color theme for the TUI
type Theme string

// Theme constants for built-in themes
const (
	ThemeDefault        Theme = "default"
	ThemeDracula        Theme = "dracula"
	ThemeGruvbox        Theme = "gruvbox"
	ThemeNord           Theme = "nord"
	ThemeSolarizedDark  Theme = "solarized-dark"
	ThemeSolarizedLight Theme = "solarized-light"
)

// ValidThemes contains all valid theme names
var ValidThemes = []Theme{ThemeDefault, ThemeDracula, ThemeGruvbox, ThemeNord, ThemeSolarizedDark, ThemeSolarizedLight}

// DefaultDisplayColumns returns the default columns to display
func DefaultDisplayColumns() []string {
	return []string{"number", "title", "author", "date", "comments"}
}

// ValidateDisplayColumn validates that a column name is valid
func ValidateDisplayColumn(column string) error {
	if slices.Contains(ValidColumns, column) {
		return nil
	}
	return fmt.Errorf("invalid display column: %q, must be one of: %v", column, ValidColumns)
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Method string `toml:"method"` // "env", "token", or "gh"
	Token  string `toml:"token,omitempty"`
}

// DefaultConfigPath returns the default path for the config file
// (~/.config/ghissues/config.toml)
func DefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home can't be determined
		return filepath.Join(".config", "ghissues", "config.toml")
	}
	return filepath.Join(homeDir, ".config", "ghissues", "config.toml")
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		Auth: AuthConfig{
			Method: "env", // Default to environment variable method
		},
	}
}

// Exists checks if a config file exists at the given path
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Save writes the config to the specified path with secure permissions (0600)
func Save(cfg *Config, path string) error {
	// Create parent directories if they don't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create or truncate the file with secure permissions
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Encode the config as TOML
	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Load reads and parses a config file from the specified path
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

// ValidateRepository validates that a repository string is in owner/repo format
func ValidateRepository(repo string) error {
	if repo == "" {
		return errors.New("repository cannot be empty")
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return errors.New("repository must be in owner/repo format")
	}

	if parts[0] == "" {
		return errors.New("owner cannot be empty")
	}

	if parts[1] == "" {
		return errors.New("repository name cannot be empty")
	}

	return nil
}

// ValidateAuthMethod validates that an auth method is one of the supported methods
func ValidateAuthMethod(method string) error {
	switch method {
	case "env", "token", "gh":
		return nil
	default:
		return fmt.Errorf("invalid auth method: %q, must be one of: env, token, gh", method)
	}
}

// ValidateSortField validates that a sort field is valid
func ValidateSortField(field SortField) error {
	if slices.Contains(ValidSortFields, field) {
		return nil
	}
	return fmt.Errorf("invalid sort field: %q, must be one of: %v", field, ValidSortFields)
}

// ValidateSortOrder validates that a sort order is valid
func ValidateSortOrder(order SortOrder) error {
	if order == SortDesc || order == SortAsc {
		return nil
	}
	return fmt.Errorf("invalid sort order: %q, must be one of: desc, asc", order)
}

// DefaultSortConfig returns the default sort field and order
// Default: most recently updated first
func DefaultSortConfig() (SortField, SortOrder) {
	return SortByUpdated, SortDesc
}

// AllSortFields returns all available sort fields
func AllSortFields() []SortField {
	return []SortField{SortByUpdated, SortByCreated, SortByNumber, SortByComments}
}

// NextSortField returns the next sort field in the cycle
func NextSortField(current SortField) SortField {
	fields := AllSortFields()
	if current == "" {
		current = SortByUpdated
	}
	for i, f := range fields {
		if f == current {
			return fields[(i+1)%len(fields)]
		}
	}
	return SortByUpdated
}

// ToggleSortOrder toggles the sort order between ascending and descending
func ToggleSortOrder(current SortOrder) SortOrder {
	if current == "" {
		current = SortDesc
	}
	if current == SortDesc {
		return SortAsc
	}
	return SortDesc
}

// DisplayName returns a human-readable name for the sort field
func (f SortField) DisplayName() string {
	switch f {
	case SortByUpdated:
		return "Updated"
	case SortByCreated:
		return "Created"
	case SortByNumber:
		return "Number"
	case SortByComments:
		return "Comments"
	default:
		return "Updated"
	}
}

// ValidateTheme validates that a theme name is valid
func ValidateTheme(theme Theme) error {
	if slices.Contains(ValidThemes, theme) {
		return nil
	}
	return fmt.Errorf("invalid theme: %q, must be one of: %v", theme, ValidThemes)
}

// DefaultTheme returns the default theme
func DefaultTheme() Theme {
	return ThemeDefault
}

// AllThemes returns all available themes
func AllThemes() []Theme {
	return []Theme{ThemeDefault, ThemeDracula, ThemeGruvbox, ThemeNord, ThemeSolarizedDark, ThemeSolarizedLight}
}

// DisplayName returns a human-readable name for the theme
func (t Theme) DisplayName() string {
	switch t {
	case ThemeDefault:
		return "Default"
	case ThemeDracula:
		return "Dracula"
	case ThemeGruvbox:
		return "Gruvbox"
	case ThemeNord:
		return "Nord"
	case ThemeSolarizedDark:
		return "Solarized Dark"
	case ThemeSolarizedLight:
		return "Solarized Light"
	default:
		return "Default"
	}
}
