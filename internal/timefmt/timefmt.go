package timefmt

import (
	"fmt"
	"time"
)

// FormatRelative formats a time as a relative string like "5 minutes ago"
// If the time is zero, returns "never"
func FormatRelative(t time.Time, now time.Time) string {
	// Handle zero time
	if t.IsZero() {
		return "never"
	}

	// Calculate duration
	duration := now.Sub(t)

	// Handle future times (shouldn't happen but be defensive)
	if duration < 0 {
		return "just now"
	}

	seconds := int(duration.Seconds())

	// Less than a minute
	if seconds < 60 {
		return "just now"
	}

	minutes := seconds / 60

	// Less than an hour
	if minutes < 60 {
		if minutes == 1 {
			return "1 minute ago"
		}
		return formatPlural(minutes, "minute")
	}

	hours := minutes / 60

	// Less than a day
	if hours < 24 {
		if hours == 1 {
			return "1 hour ago"
		}
		return formatPlural(hours, "hour")
	}

	days := hours / 24

	// Less than a week
	if days < 7 {
		if days == 1 {
			return "1 day ago"
		}
		return formatPlural(days, "day")
	}

	weeks := days / 7

	// Less than a month (approximately 30 days)
	if days < 30 {
		if weeks == 1 {
			return "1 week ago"
		}
		return formatPlural(weeks, "week")
	}

	// Months (approximately 30 days)
	months := days / 30
	if months == 1 {
		return "1 month ago"
	}
	return formatPlural(months, "month")
}

// formatPlural formats a count with the appropriate plural form
func formatPlural(count int, unit string) string {
	return fmt.Sprintf("%d %ss ago", count, unit)
}
