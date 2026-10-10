package adapters

import (
	"context"
	"sync"
	"youtrack/examples/clean_architecture/1_domain"
)

// InMemoryIssueRepository implements domain.IssueRepository as an adapter.
type InMemoryIssueRepository struct {
	mu     sync.RWMutex
	issues map[string]*domain.Issue
}

// NewInMemoryIssueRepository initializes the in-memory repository.
func NewInMemoryIssueRepository() *InMemoryIssueRepository {
	return &InMemoryIssueRepository{
		issues: make(map[string]*domain.Issue),
	}
}

func (r *InMemoryIssueRepository) Save(ctx context.Context, issue *domain.Issue) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.issues[issue.ID] = issue
	return nil
}

func (r *InMemoryIssueRepository) FindByID(ctx context.Context, id string) (*domain.Issue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	issue, exists := r.issues[id]
	if !exists {
		return nil, domain.ErrIssueNotFound
	}
	return issue, nil
}

func (r *InMemoryIssueRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.issues[id]; !exists {
		return domain.ErrIssueNotFound
	}
	delete(r.issues, id)
	return nil
}

func (r *InMemoryIssueRepository) FindByProject(ctx context.Context, projectID string) ([]*domain.Issue, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []*domain.Issue
	for _, issue := range r.issues {
		if issue.ProjectID == projectID {
			res = append(res, issue)
		}
	}
	return res, nil
}
