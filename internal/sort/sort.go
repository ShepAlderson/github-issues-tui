package sort

import (
	"fmt"
	"sort"

	"github.com/shepbook/ghissues/internal/storage"
)

// Valid sort fields
const (
	FieldUpdated  = "updated"
	FieldCreated  = "created"
	FieldNumber   = "number"
	FieldComments = "comments"
)

// ValidSortFields contains all valid sort field names
var ValidSortFields = map[string]bool{
	FieldUpdated:  true,
	FieldCreated:  true,
	FieldNumber:   true,
	FieldComments: true,
}

// SortFieldCycle defines the cycle order for sort fields
var SortFieldCycle = []string{
	FieldUpdated,
	FieldCreated,
	FieldNumber,
	FieldComments,
}

// ValidateSortField checks if a sort field is valid
func ValidateSortField(field string) error {
	if field == "" {
		return fmt.Errorf("sort field cannot be empty")
	}

	if !ValidSortFields[field] {
		return fmt.Errorf("invalid sort field: %s", field)
	}

	return nil
}

// GetDefaultSortField returns the default sort field
func GetDefaultSortField() string {
	return FieldUpdated
}

// GetDefaultSortDescending returns the default sort order
func GetDefaultSortDescending() bool {
	return true // Most recently updated first
}

// CycleSortField returns the next sort field in the cycle
func CycleSortField(current string) string {
	// If current is invalid or empty, start with default
	if current == "" || !ValidSortFields[current] {
		return GetDefaultSortField()
	}

	// Find current field in cycle and return next
	for i, field := range SortFieldCycle {
		if field == current {
			nextIndex := (i + 1) % len(SortFieldCycle)
			return SortFieldCycle[nextIndex]
		}
	}

	// Fallback to default if not found in cycle
	return GetDefaultSortField()
}

// GetSortFieldLabel returns a human-readable label for a sort field
func GetSortFieldLabel(field string) string {
	switch field {
	case FieldUpdated:
		return "Updated"
	case FieldCreated:
		return "Created"
	case FieldNumber:
		return "Number"
	case FieldComments:
		return "Comments"
	default:
		return "Unknown"
	}
}

// SortIssues sorts a slice of issues by the specified field and order
func SortIssues(issues []storage.Issue, field string, descending bool) []storage.Issue {
	// Create a copy to avoid modifying the original
	sorted := make([]storage.Issue, len(issues))
	copy(sorted, issues)

	// Use the appropriate sort function
	sort.Slice(sorted, func(i, j int) bool {
		var less bool

		switch field {
		case FieldUpdated:
			less = sorted[i].UpdatedAt.Before(sorted[j].UpdatedAt)
		case FieldCreated:
			less = sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
		case FieldNumber:
			less = sorted[i].Number < sorted[j].Number
		case FieldComments:
			less = sorted[i].Comments < sorted[j].Comments
		default:
			// Fallback to updated date if field is invalid
			less = sorted[i].UpdatedAt.Before(sorted[j].UpdatedAt)
		}

		// Reverse if descending
		if descending {
			return !less
		}
		return less
	})

	return sorted
}
