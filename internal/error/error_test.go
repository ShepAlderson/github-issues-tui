package error

import (
	"errors"
	"net"
	"testing"
)

func TestErrorSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity ErrorSeverity
		isMinor  bool
		isCritical bool
	}{
		{
			name:       "Minor severity",
			severity:   SeverityMinor,
			isMinor:    true,
			isCritical: false,
		},
		{
			name:       "Critical severity",
			severity:   SeverityCritical,
			isMinor:    false,
			isCritical: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.severity.IsMinor() != tt.isMinor {
				t.Errorf("IsMinor() = %v, want %v", tt.severity.IsMinor(), tt.isMinor)
			}
			if tt.severity.IsCritical() != tt.isCritical {
				t.Errorf("IsCritical() = %v, want %v", tt.severity.IsCritical(), tt.isCritical)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name             string
		err              error
		wantSeverity     ErrorSeverity
		wantActionable   bool
		wantContainsGuidance bool
	}{
		{
			name:             "Network timeout error",
			err:              &net.OpError{Op: "dial", Err: errors.New("timeout")},
			wantSeverity:     SeverityMinor,
			wantActionable:   true,
			wantContainsGuidance: true,
		},
		{
			name:             "Network connection refused",
			err:              &net.OpError{Op: "dial", Err: errors.New("connection refused")},
			wantSeverity:     SeverityMinor,
			wantActionable:   true,
			wantContainsGuidance: true,
		},
		{
			name:             "Rate limit error string",
			err:              errors.New("API error: rate limit exceeded (status 403)"),
			wantSeverity:     SeverityMinor,
			wantActionable:   true,
			wantContainsGuidance: true,
		},
		{
			name:             "Invalid token (401)",
			err:              errors.New("API error: Bad credentials (status 401)"),
			wantSeverity:     SeverityCritical,
			wantActionable:   true,
			wantContainsGuidance: true,
		},
		{
			name:             "Database corruption error",
			err:              errors.New("database disk image is malformed"),
			wantSeverity:     SeverityCritical,
			wantActionable:   true,
			wantContainsGuidance: true,
		},
		{
			name:             "Generic error",
			err:              errors.New("something went wrong"),
			wantSeverity:     SeverityMinor,
			wantActionable:   false,
			wantContainsGuidance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := Classify(tt.err)

			if appErr.Severity != tt.wantSeverity {
				t.Errorf("Severity = %v, want %v", appErr.Severity, tt.wantSeverity)
			}

			if appErr.Actionable != tt.wantActionable {
				t.Errorf("Actionable = %v, want %v", appErr.Actionable, tt.wantActionable)
			}

			if (appErr.Guidance != "") != tt.wantContainsGuidance {
				t.Errorf("HasGuidance = %v, want %v", appErr.Guidance != "", tt.wantContainsGuidance)
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		appErr  AppError
		want    string
	}{
		{
			name: "Error with guidance",
			appErr: AppError{
				Original: errors.New("connection timeout"),
				Severity: SeverityMinor,
				Display:  "Could not connect to GitHub",
				Guidance: "Check your internet connection and try again",
			},
			want: "Could not connect to GitHub: connection timeout",
		},
		{
			name: "Error without guidance",
			appErr: AppError{
				Original: errors.New("unknown error"),
				Severity: SeverityMinor,
				Display:  "An error occurred",
			},
			want: "An error occurred: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_DisplayMessage(t *testing.T) {
	tests := []struct {
		name    string
		appErr  AppError
		want    string
	}{
		{
			name: "Error with guidance",
			appErr: AppError{
				Original: errors.New("connection timeout"),
				Severity: SeverityMinor,
				Display:  "Could not connect to GitHub",
				Guidance: "Check your internet connection and try again",
			},
			want: "Could not connect to GitHub\n\nCheck your internet connection and try again",
		},
		{
			name: "Error without guidance",
			appErr: AppError{
				Original: errors.New("unknown error"),
				Severity: SeverityMinor,
				Display:  "An error occurred",
			},
			want: "An error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appErr.DisplayMessage(); got != tt.want {
				t.Errorf("DisplayMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifyNetworkErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantInGuidance string
	}{
		{
			name:           "Network timeout",
			err:            errors.New("operation timed out"),
			wantInGuidance: "Check your internet connection",
		},
		{
			name:           "DNS error",
			err:            errors.New("no such host"),
			wantInGuidance: "Check your internet connection",
		},
		{
			name:           "Connection reset",
			err:            errors.New("connection reset by peer"),
			wantInGuidance: "retry",
		},
		{
			name:           "Temporary failure",
			err:            errors.New("temporary failure in name resolution"),
			wantInGuidance: "temporary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := Classify(tt.err)
			if !appErr.Actionable {
				t.Errorf("Network errors should be actionable")
			}
			if appErr.Severity != SeverityMinor {
				t.Errorf("Network errors should be minor severity")
			}
		})
	}
}

func TestClassifyDatabaseErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "Database malformed",
			err:  errors.New("database disk image is malformed"),
		},
		{
			name: "Database corrupted",
			err:  errors.New("database corruption detected"),
		},
		{
			name: "Database not writable",
			err:  errors.New("attempt to write a readonly database"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := Classify(tt.err)
			if appErr.Severity != SeverityCritical {
				t.Errorf("Database corruption should be critical severity, got %v", appErr.Severity)
			}
		})
	}
}

func TestClassifyAuthErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantInGuidance string
	}{
		{
			name:           "Bad credentials",
			err:            errors.New("API error: Bad credentials (status 401)"),
			wantInGuidance: "gh auth login",
		},
		{
			name:           "Token invalid",
			err:            errors.New("token is invalid or expired"),
			wantInGuidance: "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := Classify(tt.err)
			if appErr.Severity != SeverityCritical {
				t.Errorf("Auth errors should be critical severity, got %v", appErr.Severity)
			}
			if !appErr.Actionable {
				t.Errorf("Auth errors should be actionable")
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retryable bool
	}{
		{
			name:      "Network timeout is retryable",
			err:       errors.New("connection timeout"),
			retryable: true,
		},
		{
			name:      "Rate limit is retryable",
			err:       errors.New("rate limit exceeded"),
			retryable: true,
		},
		{
			name:      "Auth error not retryable",
			err:       errors.New("Bad credentials"),
			retryable: false,
		},
		{
			name:      "Database error not retryable",
			err:       errors.New("database malformed"),
			retryable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := Classify(tt.err)
			if appErr.Retryable != tt.retryable {
				t.Errorf("Retryable = %v, want %v (error: %v)", appErr.Retryable, tt.retryable, tt.err)
			}
		})
	}
}
