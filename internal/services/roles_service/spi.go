package roles

import (
	"fmt"

	perms "youtrack/internal/services/permissions_services"
)

type AssignRef struct {
	AssignType AssignType `json:"principal_type"` // USER or GROUP type
	AssignerID string     `json:"principal_id"`   // User IDs or Group IDs
}

func UserAssign(userID string) AssignRef {
	return AssignRef{
		AssignerID: userID,
		AssignType: AssignUser,
	}
}

func GroupAssign(groupID string) AssignRef {
	return AssignRef{
		AssignerID: groupID,
		AssignType: AssignGroup,
	}
}

// ScopeRef represents an abstract scope reference identified by its level and target ID.
type ScopeRef struct {
	Scope    perms.ScopeLevel `json:"level"`
	TargetID string           `json:"target_id"`
}

// GlobalScope creates a ScopeRef for the Global scope.
func GlobalScope() ScopeRef {
	return ScopeRef{
		Scope:    perms.ScopeGlobal,
		TargetID: "",
	}
	//return ScopedTarget(perms.ScopeGlobal, "")
}

func ScopedOrganization(organizationID string) ScopeRef {
	return ScopeRef{
		Scope:    perms.ScopeOrganization,
		TargetID: organizationID,
	}
	//return ScopedTarget(perms.ScopeOrganization, organizationID)
}

func ScopedProject(projectID string) ScopeRef {
	return ScopeRef{
		Scope:    perms.ScopeProject,
		TargetID: projectID,
	}
	// return ScopedTarget(perms.ScopeProject, projectID)
}

// ScopedTarget creates a ScopeRef for a specific scoped target (e.g. Organization or Project).
// func ScopedTarget(level perms.ScopeLevel, targetID string) ScopeRef {
// 	return ScopeRef{
// 		Level:    level,
// 		TargetID: targetID,
// 	}
// }

// String returns a readable representation of the scope.
func (s ScopeRef) String() string {
	if s.Scope == perms.ScopeGlobal || s.TargetID == "" {
		return string(s.Scope)
	}
	return fmt.Sprintf("%s:%s", s.Scope, s.TargetID)
}

// PrincipalHierarchyProvider is the SPI for resolving inherited or associated principal identities.
// Given a principal type and ID, it returns all identities that inherit or represent this principal.
// For example: For a User, it returns [UserID, DirectGroupID1, ParentGroupID2, AllUsersGroupID].
// For a Group, it returns [GroupID, ParentGroupID].
type PrincipalHierarchyProvider interface {
	ResolveIdentities(principalType AssignType, principalID string) ([]string, error)
}

// ScopeHierarchyProvider is the SPI for resolving hierarchical scope relationships and target validity.
// For example: For a Project scope, it returns [OrganizationScopeRef, GlobalScopeRef].
type ScopeHierarchyProvider interface {
	GetParentScopes(scope ScopeRef) ([]ScopeRef, error)
	ValidateScope(scope ScopeRef) error
}

// InMemoryHierarchyProvider is a reference in-memory implementation of both
// PrincipalHierarchyProvider and ScopeHierarchyProvider, useful for testing and standalone usage.
type InMemoryHierarchyProvider struct {
	// userGroups: userID -> list of direct group IDs
	userGroups map[string][]string
	// groupParents: groupID -> parentGroupID
	groupParents map[string]string
	// validScopes: "Level:TargetID" -> exists
	validScopes map[string]bool
	// scopeParents: ScopeRef.String() -> list of parent ScopeRefs
	scopeParents map[string][]ScopeRef
	// allUsersGroupID: default virtual group ID
	allUsersGroupID string
}

// NewInMemoryHierarchyProvider creates a new InMemoryHierarchyProvider instance.
func NewInMemoryHierarchyProvider() *InMemoryHierarchyProvider {
	return &InMemoryHierarchyProvider{
		userGroups:      make(map[string][]string),
		groupParents:    make(map[string]string),
		validScopes:     make(map[string]bool),
		scopeParents:    make(map[string][]ScopeRef),
		allUsersGroupID: AllUsersGroupID,
	}
}

// SetAllUsersGroupID configures the virtual all-users group ID.
func (p *InMemoryHierarchyProvider) SetAllUsersGroupID(id string) {
	p.allUsersGroupID = id
}

// RegisterScope registers a scope as valid and optionally defines its parent hierarchy.
func (p *InMemoryHierarchyProvider) RegisterScope(scope ScopeRef, parents ...ScopeRef) {
	p.validScopes[scope.String()] = true
	if len(parents) > 0 {
		p.scopeParents[scope.String()] = parents
	}
}

// AddUserToGroup associates a user with a group.
func (p *InMemoryHierarchyProvider) AddUserToGroup(userID string, groupID string) {
	p.userGroups[userID] = append(p.userGroups[userID], groupID)
}

// SetGroupParent defines nested group inheritance (childGroup inherits parentGroup).
func (p *InMemoryHierarchyProvider) SetGroupParent(childGroupID string, parentGroupID string) {
	p.groupParents[childGroupID] = parentGroupID
}

// ResolveIdentities resolves all identities associated with the given principal.
func (p *InMemoryHierarchyProvider) ResolveIdentities(principalType AssignType, principalID string) ([]string, error) {
	identitySet := make(map[string]struct{})
	identitySet[principalID] = struct{}{}

	var queue []string

	switch principalType {
	case AssignUser:
		if p.allUsersGroupID != "" {
			identitySet[p.allUsersGroupID] = struct{}{}
		}
		// Add direct groups
		for _, gid := range p.userGroups[principalID] {
			if _, exists := identitySet[gid]; !exists {
				identitySet[gid] = struct{}{}
				queue = append(queue, gid)
			}
		}
	case AssignGroup:
		queue = append(queue, principalID)
	}

	// Traverse parent groups
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if parentID, ok := p.groupParents[curr]; ok && parentID != "" {
			if _, exists := identitySet[parentID]; !exists {
				identitySet[parentID] = struct{}{}
				queue = append(queue, parentID)
			}
		}
	}

	result := make([]string, 0, len(identitySet))
	for id := range identitySet {
		result = append(result, id)
	}
	return result, nil
}

// GetParentScopes returns the parent scopes for a given scope.
func (p *InMemoryHierarchyProvider) GetParentScopes(scope ScopeRef) ([]ScopeRef, error) {
	if scope.Scope == perms.ScopeGlobal {
		return nil, nil
	}
	parents, ok := p.scopeParents[scope.String()]
	if !ok {
		// By default, every scoped target's ancestor is GlobalScope
		return []ScopeRef{GlobalScope()}, nil
	}
	return parents, nil
}

// ValidateScope checks if a scope is valid.
func (p *InMemoryHierarchyProvider) ValidateScope(scope ScopeRef) error {
	if scope.Scope == perms.ScopeGlobal {
		return nil
	}
	if scope.TargetID == "" {
		return ErrTargetRequiredForScope
	}
	if !p.validScopes[scope.String()] {
		return ErrTargetNotFound
	}
	return nil
}
