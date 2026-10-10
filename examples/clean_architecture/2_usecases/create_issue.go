package usecases

import (
	"context"
	"fmt"
	"youtrack/examples/clean_architecture/1_domain"
	"youtrack/internal/services/rbac"
)

// CreateIssueCommand encapsulates input parameters for creating an issue.
type CreateIssueCommand struct {
	UserID      string
	UserRoles   []string
	IssueID     string
	ProjectID   string
	Title       string
	Description string
}

// CreateIssueUseCase coordinates issue creation while enforcing RBAC authorization.
type CreateIssueUseCase struct {
	repo       domain.IssueRepository
	authorizer domain.Authorizer
}

// NewCreateIssueUseCase injects the repository and authorizer dependencies.
func NewCreateIssueUseCase(repo domain.IssueRepository, authorizer domain.Authorizer) *CreateIssueUseCase {
	return &CreateIssueUseCase{
		repo:       repo,
		authorizer: authorizer,
	}
}

// Execute performs the use case workflow with Defense in Depth authorization.
func (uc *CreateIssueUseCase) Execute(ctx context.Context, cmd CreateIssueCommand) (*domain.Issue, error) {
	// 1. Internal Policy Enforcement Point (PEP):
	// Verifies the user's role grants CREATE_ISSUE within the targeted Project scope.
	authReq := rbac.AccessRequest{
		SubjectID:  cmd.UserID,
		RoleIDs:    cmd.UserRoles,
		Permission: "CREATE_ISSUE",
		Resource:   cmd.ProjectID,
		Scope:      "PROJECT:" + cmd.ProjectID,
		Context: map[string]any{
			"resource_scope": "PROJECT:" + cmd.ProjectID,
		},
	}

	if err := uc.authorizer.Authorize(ctx, authReq); err != nil {
		return nil, fmt.Errorf("authorization denied: %w", err)
	}

	// 2. Instantiate domain entity enforcing business invariants
	issue, err := domain.NewIssue(cmd.IssueID, cmd.ProjectID, cmd.Title, cmd.Description, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// 3. Save via repository outbound port
	if err := uc.repo.Save(ctx, issue); err != nil {
		return nil, err
	}

	return issue, nil
}
