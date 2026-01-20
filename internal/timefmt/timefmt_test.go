package timefmt

import (
	"testing"
	"time"
)

func TestFormatRelative(t *testing.T) {
	tests := []struct {
		name     string
		t        time.Time
		now      time.Time
		expected string
	}{
		{
			name:     "just now",
			t:        time.Now().Add(-30 * time.Second),
			now:      time.Now(),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			t:        time.Now().Add(-1 * time.Minute),
			now:      time.Now(),
			expected: "1 minute ago",
		},
		{
			name:     "5 minutes ago",
			t:        time.Now().Add(-5 * time.Minute),
			now:      time.Now(),
			expected: "5 minutes ago",
		},
		{
			name:     "1 hour ago",
			t:        time.Now().Add(-1 * time.Hour),
			now:      time.Now(),
			expected: "1 hour ago",
		},
		{
			name:     "3 hours ago",
			t:        time.Now().Add(-3 * time.Hour),
			now:      time.Now(),
			expected: "3 hours ago",
		},
		{
			name:     "1 day ago",
			t:        time.Now().Add(-24 * time.Hour),
			now:      time.Now(),
			expected: "1 day ago",
		},
		{
			name:     "5 days ago",
			t:        time.Now().Add(-5 * 24 * time.Hour),
			now:      time.Now(),
			expected: "5 days ago",
		},
		{
			name:     "1 week ago",
			t:        time.Now().Add(-7 * 24 * time.Hour),
			now:      time.Now(),
			expected: "1 week ago",
		},
		{
			name:     "4 weeks ago",
			t:        time.Now().Add(-28 * 24 * time.Hour),
			now:      time.Now(),
			expected: "4 weeks ago",
		},
		{
			name:     "1 month ago",
			t:        time.Now().Add(-30 * 24 * time.Hour),
			now:      time.Now(),
			expected: "1 month ago",
		},
		{
			name:     "12 months ago",
			t:        time.Now().Add(-365 * 24 * time.Hour),
			now:      time.Now(),
			expected: "12 months ago",
		},
		{
			name:     "zero time",
			t:        time.Time{},
			now:      time.Now(),
			expected: "never",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelative(tt.t, tt.now)
			if result != tt.expected {
				t.Errorf("FormatRelative(%v) = %q, want %q", tt.t, result, tt.expected)
			}
		})
	}
}

func TestFormatRelativeNoNow(t *testing.T) {
	// Test the convenience function that uses time.Now()
	past := time.Now().Add(-5 * time.Minute)
	result := FormatRelative(past, time.Now())
	if result != "5 minutes ago" {
		t.Errorf("FormatRelative(5 minutes ago) = %q, want %q", result, "5 minutes ago")
	}
}
