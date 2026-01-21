package tui

import (
	"testing"
	"time"
)

func TestFormatRelativeTimeIntegration(t *testing.T) {
	// Test that FormatRelativeTime works with various time inputs
	now := time.Now()

	testCases := []struct {
		name     string
		input    time.Time
		contains string // We check if output contains this substring
	}{
		{
			name:     "zero time returns never",
			input:    time.Time{},
			contains: "never",
		},
		{
			name:     "recent time contains ago",
			input:    now.Add(-5 * time.Minute),
			contains: "ago",
		},
		{
			name:     "just now",
			input:    now.Add(-30 * time.Second),
			contains: "just now",
		},
		{
			name:     "future time handled gracefully",
			input:    now.Add(5 * time.Minute),
			contains: "just now",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatRelativeTime(tc.input)
			if tc.contains == "" {
				t.Errorf("Test case must specify 'contains' substring")
			}
			// Check if result contains the expected substring
			contains := false
			for i := 0; i <= len(result)-len(tc.contains); i++ {
				if result[i:i+len(tc.contains)] == tc.contains {
					contains = true
					break
				}
			}
			if !contains {
				t.Errorf("FormatRelativeTime(%v) = %q, expected to contain %q", tc.input, result, tc.contains)
			}
		})
	}
}