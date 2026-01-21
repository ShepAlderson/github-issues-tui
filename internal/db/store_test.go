package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/shepbook/ghissues/internal/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	assert.NotNil(t, store)
}

func TestStore_SaveAndGetIssue(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	issue := github.Issue{
		Number:       1,
		Title:        "Test Issue",
		Body:         "This is a test issue body",
		Author:       github.User{Login: "testuser"},
		CreatedAt:    "2024-01-15T10:30:00Z",
		UpdatedAt:    "2024-01-16T12:00:00Z",
		CommentCount: 5,
		Labels: []github.Label{
			{Name: "bug", Color: "ff0000"},
			{Name: "priority", Color: "00ff00"},
		},
		Assignees: []github.User{
			{Login: "assignee1"},
			{Login: "assignee2"},
		},
	}

	ctx := context.Background()
	err = store.SaveIssue(ctx, &issue)
	require.NoError(t, err)

	// Retrieve the issue
	retrieved, err := store.GetIssue(ctx, 1)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, issue.Number, retrieved.Number)
	assert.Equal(t, issue.Title, retrieved.Title)
	assert.Equal(t, issue.Body, retrieved.Body)
	assert.Equal(t, issue.Author.Login, retrieved.Author.Login)
	assert.Equal(t, issue.CommentCount, retrieved.CommentCount)
	assert.Len(t, retrieved.Labels, 2)
	assert.Len(t, retrieved.Assignees, 2)
}

func TestStore_SaveIssues_Batch(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	issues := []github.Issue{
		{Number: 1, Title: "Issue 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
		{Number: 2, Title: "Issue 2", Author: github.User{Login: "user2"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
		{Number: 3, Title: "Issue 3", Author: github.User{Login: "user3"}, CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
	}

	ctx := context.Background()
	err = store.SaveIssues(ctx, issues)
	require.NoError(t, err)

	// Retrieve all issues
	retrieved, err := store.GetAllIssues(ctx)
	require.NoError(t, err)
	assert.Len(t, retrieved, 3)
}

func TestStore_SaveAndGetComments(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Save an issue first
	issue := github.Issue{
		Number:    1,
		Title:     "Test Issue",
		Author:    github.User{Login: "testuser"},
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	ctx := context.Background()
	err = store.SaveIssue(ctx, &issue)
	require.NoError(t, err)

	// Save comments
	comments := []github.Comment{
		{ID: "c1", Body: "Comment 1", Author: github.User{Login: "user1"}, CreatedAt: "2024-01-15T11:00:00Z"},
		{ID: "c2", Body: "Comment 2", Author: github.User{Login: "user2"}, CreatedAt: "2024-01-15T12:00:00Z"},
	}

	err = store.SaveComments(ctx, 1, comments)
	require.NoError(t, err)

	// Retrieve comments
	retrieved, err := store.GetComments(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, retrieved, 2)
	assert.Equal(t, "Comment 1", retrieved[0].Body)
}

func TestStore_UpdateIssue(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Save initial issue
	issue := github.Issue{
		Number:    1,
		Title:     "Original Title",
		Body:      "Original body",
		Author:    github.User{Login: "testuser"},
		CreatedAt: "2024-01-15T10:30:00Z",
		UpdatedAt: "2024-01-15T10:30:00Z",
	}

	ctx := context.Background()
	err = store.SaveIssue(ctx, &issue)
	require.NoError(t, err)

	// Update the issue
	issue.Title = "Updated Title"
	issue.Body = "Updated body"
	issue.UpdatedAt = "2024-01-16T12:00:00Z"

	err = store.SaveIssue(ctx, &issue)
	require.NoError(t, err)

	// Verify update
	retrieved, err := store.GetIssue(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", retrieved.Title)
	assert.Equal(t, "Updated body", retrieved.Body)
}

func TestStore_ClearIssues(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Save some issues
	issues := []github.Issue{
		{Number: 1, Title: "Issue 1", CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
		{Number: 2, Title: "Issue 2", CreatedAt: "2024-01-15T10:30:00Z", UpdatedAt: "2024-01-15T10:30:00Z"},
	}

	ctx := context.Background()
	err = store.SaveIssues(ctx, issues)
	require.NoError(t, err)

	// Clear all issues
	err = store.ClearIssues(ctx)
	require.NoError(t, err)

	// Verify empty
	retrieved, err := store.GetAllIssues(ctx)
	require.NoError(t, err)
	assert.Empty(t, retrieved)
}

func TestStore_GetLastSyncTime(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()

	// Initially should be zero
	lastSync, err := store.GetLastSyncTime(ctx)
	require.NoError(t, err)
	assert.True(t, lastSync.IsZero())

	// Update sync time
	now := time.Now().Truncate(time.Second)
	err = store.SetLastSyncTime(ctx, now)
	require.NoError(t, err)

	// Verify
	lastSync, err = store.GetLastSyncTime(ctx)
	require.NoError(t, err)
	assert.Equal(t, now.UTC(), lastSync.UTC())
}

func TestStore_NonExistentIssue(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	store, err := NewStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	ctx := context.Background()
	issue, err := store.GetIssue(ctx, 999)
	require.NoError(t, err)
	assert.Nil(t, issue)
}
