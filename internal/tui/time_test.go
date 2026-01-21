package tui

import (
	"testing"
	"time"
)

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "just now",
			input:    now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			input:    now.Add(-1 * time.Minute),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			input:    now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			input:    now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "2 hours ago",
			input:    now.Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "1 day ago",
			input:    now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "2 days ago",
			input:    now.Add(-48 * time.Hour),
			expected: "2 days ago",
		},
		{
			name:     "1 week ago",
			input:    now.Add(-7 * 24 * time.Hour),
			expected: "1 week ago",
		},
		{
			name:     "2 weeks ago",
			input:    now.Add(-14 * 24 * time.Hour),
			expected: "2 weeks ago",
		},
		{
			name:     "1 month ago",
			input:    now.Add(-30 * 24 * time.Hour),
			expected: "1 month ago",
		},
		{
			name:     "2 months ago",
			input:    now.Add(-60 * 24 * time.Hour),
			expected: "2 months ago",
		},
		{
			name:     "6 months ago",
			input:    now.Add(-180 * 24 * time.Hour),
			expected: "6 months ago",
		},
		{
			name:     "1 year ago",
			input:    now.Add(-365 * 24 * time.Hour),
			expected: "1 year ago",
		},
		{
			name:     "zero time",
			input:    time.Time{},
			expected: "never",
		},
		{
			name:     "future time",
			input:    now.Add(5 * time.Minute),
			expected: "just now",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelativeTime(tt.input)
			if result != tt.expected {
				t.Errorf("FormatRelativeTime(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatRelativeTimePluralization(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "plural minutes",
			input:    now.Add(-2 * time.Minute),
			expected: "2 minutes ago",
		},
		{
			name:     "plural hours",
			input:    now.Add(-3 * time.Hour),
			expected: "3 hours ago",
		},
		{
			name:     "plural days",
			input:    now.Add(-5 * 24 * time.Hour),
			expected: "5 days ago",
		},
		{
			name:     "plural weeks",
			input:    now.Add(-3 * 7 * 24 * time.Hour),
			expected: "3 weeks ago",
		},
		{
			name:     "plural months",
			input:    now.Add(-4 * 30 * 24 * time.Hour),
			expected: "4 months ago",
		},
		{
			name:     "plural years",
			input:    now.Add(-2 * 365 * 24 * time.Hour),
			expected: "2 years ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelativeTime(tt.input)
			if result != tt.expected {
				t.Errorf("FormatRelativeTime(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}