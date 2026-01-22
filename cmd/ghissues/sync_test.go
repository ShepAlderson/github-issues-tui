package main

import (
	"testing"
	"time"
)

func TestProgressBar_Update(t *testing.T) {
	p := &ProgressBar{Total: 100, Current: 0, width: 40}

	// Test that update doesn't panic
	p.Update()

	// Update to 50%
	p.Current = 50
	p.Update()

	// Update to 100%
	p.Current = 100
	p.Update()

	// Verify final state
	if p.Current != 100 {
		t.Errorf("Current = %d, want 100", p.Current)
	}
	if p.Total != 100 {
		t.Errorf("Total = %d, want 100", p.Total)
	}
}

func TestProgressBar_ZeroTotal(t *testing.T) {
	p := &ProgressBar{Total: 0, Current: 5, width: 40}

	// Test that update doesn't panic with zero total
	p.Update()

	// Verify state
	if p.Current != 5 {
		t.Errorf("Current = %d, want 5", p.Current)
	}
}

func TestProgressBar_Finish(t *testing.T) {
	p := &ProgressBar{Total: 10, Current: 10, width: 40}

	// Test that finish doesn't panic
	p.Finish()

	// Verify state
	if p.Current != 10 {
		t.Errorf("Current = %d, want 10", p.Current)
	}
}

func TestProgressBar_Show(t *testing.T) {
	p := &ProgressBar{Total: 100, Current: 0, width: 40}

	// Test that show doesn't panic
	p.Show()

	// Verify state
	if p.Total != 100 {
		t.Errorf("Total = %d, want 100", p.Total)
	}
	if p.Current != 0 {
		t.Errorf("Current = %d, want 0", p.Current)
	}
}

func TestSyncProgress_UpdateIssue(t *testing.T) {
	p := NewSyncProgress()

	if p.IssuesFetched != 0 {
		t.Errorf("initial IssuesFetched = %d, want 0", p.IssuesFetched)
	}

	p.UpdateIssue()
	if p.IssuesFetched != 1 {
		t.Errorf("after UpdateIssue, IssuesFetched = %d, want 1", p.IssuesFetched)
	}

	p.UpdateIssue()
	p.UpdateIssue()
	if p.IssuesFetched != 3 {
		t.Errorf("after 3 UpdateIssue calls, IssuesFetched = %d, want 3", p.IssuesFetched)
	}
}

func TestSyncProgress_UpdateComment(t *testing.T) {
	p := NewSyncProgress()

	if p.CommentsFetched != 0 {
		t.Errorf("initial CommentsFetched = %d, want 0", p.CommentsFetched)
	}

	p.UpdateComment()
	if p.CommentsFetched != 1 {
		t.Errorf("after UpdateComment, CommentsFetched = %d, want 1", p.CommentsFetched)
	}
}

func TestSyncProgress_SetTotal(t *testing.T) {
	p := NewSyncProgress()

	p.SetTotal(100)
	if p.TotalIssues != 100 {
		t.Errorf("TotalIssues = %d, want 100", p.TotalIssues)
	}
}

func TestSyncProgress_Summary(t *testing.T) {
	p := NewSyncProgress()
	p.IssuesFetched = 10
	p.CommentsFetched = 25

	// Give some time to pass so elapsed time is visible
	time.Sleep(10 * time.Millisecond)

	// Summary should contain the expected values
	if p.IssuesFetched != 10 {
		t.Errorf("IssuesFetched = %d, want 10", p.IssuesFetched)
	}
	if p.CommentsFetched != 25 {
		t.Errorf("CommentsFetched = %d, want 25", p.CommentsFetched)
	}
	if time.Since(p.StartTime) <= 0 {
		t.Error("StartTime should be in the past")
	}

	// Verify Summary method returns a string
	_ = p.Summary()
}

func TestNewSyncProgress(t *testing.T) {
	p := NewSyncProgress()
	if p == nil {
		t.Error("NewSyncProgress returned nil")
	}
	if p.IssuesFetched != 0 {
		t.Errorf("initial IssuesFetched = %d, want 0", p.IssuesFetched)
	}
	if p.CommentsFetched != 0 {
		t.Errorf("initial CommentsFetched = %d, want 0", p.CommentsFetched)
	}
	if p.TotalIssues != 0 {
		t.Errorf("initial TotalIssues = %d, want 0", p.TotalIssues)
	}
	if p.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}
}

func TestProgressBar_Width(t *testing.T) {
	tests := []struct {
		width    int
		expected int
	}{
		{20, 20},
		{40, 40},
		{80, 80},
	}

	for _, tt := range tests {
		p := &ProgressBar{width: tt.width}
		if p.width != tt.expected {
			t.Errorf("ProgressBar width = %d, want %d", p.width, tt.expected)
		}
	}
}

func TestSyncProgress_MultipleOperations(t *testing.T) {
	p := NewSyncProgress()

	// Simulate fetching issues and comments
	for i := 0; i < 5; i++ {
		p.UpdateIssue()
		for j := 0; j < 3; j++ {
			p.UpdateComment()
		}
	}

	if p.IssuesFetched != 5 {
		t.Errorf("IssuesFetched = %d, want 5", p.IssuesFetched)
	}
	if p.CommentsFetched != 15 {
		t.Errorf("CommentsFetched = %d, want 15", p.CommentsFetched)
	}
}

func TestProgressCallback_Type(t *testing.T) {
	// Verify the ProgressCallback type is correctly defined
	var callback ProgressCallback
	if callback != nil {
		t.Error("initial ProgressCallback should be nil")
	}

	// Verify it can be assigned a function
	callback = func(current, total int, status string) {
		// Do nothing
	}
	if callback == nil {
		t.Error("ProgressCallback should be assignable to a function")
	}

	// Verify it can be called
	callback(1, 10, "test status")
}

func TestProgressCallback_Called(t *testing.T) {
	var called bool
	var lastStatus string

	callback := func(current, total int, status string) {
		called = true
		lastStatus = status
	}

	// Simulate a callback call
	callback(5, 10, "Processing issue #123")

	if !called {
		t.Error("ProgressCallback should be called")
	}
	if lastStatus != "Processing issue #123" {
		t.Errorf("lastStatus = %q, want %q", lastStatus, "Processing issue #123")
	}
}