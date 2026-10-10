package usecases

import (
	"context"
	"fmt"
	"youtrack/examples/clean_architecture/1_domain"
	"youtrack/internal/services/rbac"
)

// DeleteIssueCommand encapsulates parameters required to delete an issue.
type DeleteIssueCommand struct {
	UserID    string
	UserRoles []string
	IssueID   string
}

// DeleteIssueUseCase coordinates issue deletion with dual RBAC and Inherent Ownership evaluation.
type DeleteIssueUseCase struct {
	repo       domain.IssueRepository
	authorizer domain.Authorizer
}

// NewDeleteIssueUseCase injects the repository and authorizer dependencies.
func NewDeleteIssueUseCase(repo domain.IssueRepository, authorizer domain.Authorizer) *DeleteIssueUseCase {
	return &DeleteIssueUseCase{
		repo:       repo,
		authorizer: authorizer,
	}
}

// Execute enforces that either:
// A) The user holds an administrative role granting DELETE_ISSUE for this project.
// B) The user is the inherent author/owner of the issue (hybrid context attribute).
func (uc *DeleteIssueUseCase) Execute(ctx context.Context, cmd DeleteIssueCommand) error {
	// 1. Fetch issue from repository
	issue, err := uc.repo.FindByID(ctx, cmd.IssueID)
	if err != nil {
		return err
	}

	// 2. Internal Policy Enforcement Point (PEP):
	authReq := rbac.AccessRequest{
		SubjectID:  cmd.UserID,
		RoleIDs:    cmd.UserRoles,
		Permission: "DELETE_ISSUE",
		Resource:   issue.ID,
		Scope:      "PROJECT:" + issue.ProjectID,
		Context: map[string]any{
			"resource_scope": "PROJECT:" + issue.ProjectID,
			"owner_id":       issue.AuthorID,
			"author_id":      issue.AuthorID,
			"is_author":      cmd.UserID == issue.AuthorID,
		},
	}

	if err := uc.authorizer.Authorize(ctx, authReq); err != nil {
		return fmt.Errorf("authorization denied to delete issue: %w", err)
	}

	// 3. Delete from repository
	return uc.repo.Delete(ctx, cmd.IssueID)
}
