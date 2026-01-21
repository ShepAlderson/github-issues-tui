package tui

import (
	"strconv"
	"time"
)

// FormatRelativeTime formats a time as a human-readable relative string
// Examples: "just now", "5 minutes ago", "2 hours ago", "3 days ago"
func FormatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	now := time.Now()
	diff := now.Sub(t)

	// Handle future times (shouldn't happen with sync timestamps, but good to handle)
	if diff < 0 {
		return "just now"
	}

	// Just now (less than a minute)
	if diff < time.Minute {
		return "just now"
	}

	// Minutes
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return formatPlural(minutes, "minute", "minutes")
	}

	// Hours
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return formatPlural(hours, "hour", "hours")
	}

	// Days
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return formatPlural(days, "day", "days")
	}

	// Weeks (up to 4 weeks)
	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / (24 * 7))
		if weeks == 1 {
			return "1 week ago"
		}
		return formatPlural(weeks, "week", "weeks")
	}

	// Months (up to 12 months)
	if diff < 365*24*time.Hour {
		months := int(diff.Hours() / (24 * 30))
		if months == 1 {
			return "1 month ago"
		}
		return formatPlural(months, "month", "months")
	}

	// Years
	years := int(diff.Hours() / (24 * 365))
	if years == 1 {
		return "1 year ago"
	}
	return formatPlural(years, "year", "years")
}

// formatPlural returns the correct singular or plural form
func formatPlural(count int, singular, plural string) string {
	if count == 1 {
		return "1 " + singular + " ago"
	}
	return formatInt(count) + " " + plural + " ago"
}

// formatInt formats an integer as a string
func formatInt(n int) string {
	return strconv.Itoa(n)
}