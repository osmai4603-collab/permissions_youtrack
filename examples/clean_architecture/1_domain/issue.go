package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrIssueNotFound    = errors.New("domain: issue not found")
	ErrInvalidIssueData = errors.New("domain: invalid issue title or project ID")
)

// Issue represents the core business entity in our Issue Tracking domain.
// It is completely free of external dependencies (pure Go struct).
type Issue struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	AuthorID    string    `json:"author_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewIssue constructs a valid Issue entity enforcing domain invariants.
func NewIssue(id, projectID, title, description, authorID string) (*Issue, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(projectID) == "" {
		return nil, ErrInvalidIssueData
	}
	if strings.TrimSpace(title) == "" || strings.TrimSpace(authorID) == "" {
		return nil, ErrInvalidIssueData
	}

	return &Issue{
		ID:          id,
		ProjectID:   projectID,
		Title:       title,
		Description: description,
		AuthorID:    authorID,
		Status:      "OPEN",
		CreatedAt:   time.Now(),
	}, nil
}
