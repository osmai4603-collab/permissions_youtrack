package domain

import (
	"context"
	"youtrack/internal/services/rbac"
)

// IssueRepository represents the outbound port for persisting issues.
// Outer layers (adapters/infrastructure) must implement this interface.
type IssueRepository interface {
	Save(ctx context.Context, issue *Issue) error
	FindByID(ctx context.Context, id string) (*Issue, error)
	Delete(ctx context.Context, id string) error
	FindByProject(ctx context.Context, projectID string) ([]*Issue, error)
}

// Authorizer represents the security outbound port required by use cases.
// By defining this interface in the domain/usecase boundary, use cases depend
// only on abstractions, satisfying the Dependency Inversion Principle (DIP).
type Authorizer interface {
	Authorize(ctx context.Context, req rbac.AccessRequest) error
}
