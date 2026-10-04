package perms

import (
	"sort"
)

// IPermissionService defines the interface for managing and evaluating permissions.
type IPermissionService interface {
	GetPermission(id string) (Permission, bool)
	GetAllPermissions() []Permission
	GetPermissionsByScope(scope ScopeLevel) []Permission
	GetPermissionsByEntity(entity EntityType) []Permission
	ResolveImplied(permissionIDs []string) []string
	ResolveRevocation(activePermissionIDs []string, permissionToRemove string) []string
	ValidatePermissionsForScope(permissionIDs []string, targetScope ScopeLevel) []string
	HasPermission(grantedPermissions []string, requiredPermission string) bool
	CheckInherentAccess(action InherentAction, isAuthorOrReporter bool, hasPermission func(string) bool) bool
}

// Service provides YouTrack permission management, validation, and evaluation.
type Service struct {
	catalog     map[string]Permission
	listOrdered []Permission
}

// NewService instantiates a new permissions service populated with the YouTrack catalog.
func NewService() *Service {
	cat := BuildDefaultCatalog()
	list := make([]Permission, 0, len(cat))
	for _, p := range cat {
		list = append(list, p)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].Entity != list[j].Entity {
			return list[i].Entity < list[j].Entity
		}
		return list[i].ID < list[j].ID
	})

	return &Service{
		catalog:     cat,
		listOrdered: list,
	}
}

// GetPermission returns a permission definition by its ID.
func (s *Service) GetPermission(id string) (Permission, bool) {

	p, ok := s.catalog[id]
	return p, ok
}

// GetAllPermissions returns all registered permissions in deterministic order.
func (s *Service) GetAllPermissions() []Permission {

	out := make([]Permission, len(s.listOrdered))
	copy(out, s.listOrdered)
	return out
}

// GetPermissionsByScope retrieves all permissions belonging to a specific scope level.
func (s *Service) GetPermissionsByScope(scope ScopeLevel) []Permission {

	var res []Permission
	for _, p := range s.listOrdered {
		if p.Scope == scope {
			res = append(res, p)
		}
	}
	return res
}


// GetPermissionsByEntity retrieves permissions targeting a specific entity type.
func (s *Service) GetPermissionsByEntity(entity EntityType) []Permission {

	var res []Permission
	for _, p := range s.listOrdered {
		if p.Entity == entity {
			res = append(res, p)
		}
	}
	return res
}

// ResolveImplied computes the transitive closure of implied permissions.
// As defined by YouTrack: when you add a permission with implied permissions to a role,
// the implied permissions are added automatically.
func (s *Service) ResolveImplied(permissionIDs []string) []string {

	resultMap := make(map[string]struct{})
	var queue []string

	for _, id := range permissionIDs {
		if _, exists := s.catalog[id]; exists {
			if _, seen := resultMap[id]; !seen {
				resultMap[id] = struct{}{}
				queue = append(queue, id)
			}
		}
	}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if p, ok := s.catalog[currentID]; ok {
			for _, impliedID := range p.ImpliedPerms {
				if _, seen := resultMap[impliedID]; !seen {
					resultMap[impliedID] = struct{}{}
					queue = append(queue, impliedID)
				}
			}
		}
	}

	out := make([]string, 0, len(resultMap))
	for id := range resultMap {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// ResolveRevocation determines which permissions remain after revoking a target permission.
// As defined by YouTrack: when you remove a permission with dependent permissions from a role,
// the dependent permissions are removed automatically.
func (s *Service) ResolveRevocation(activePermissionIDs []string, permissionToRemove string) []string {

	activeSet := make(map[string]struct{}, len(activePermissionIDs))
	for _, id := range activePermissionIDs {
		activeSet[id] = struct{}{}
	}

	// Queue permissions to delete (starting with the target)
	toDelete := make(map[string]struct{})
	queue := []string{permissionToRemove}
	toDelete[permissionToRemove] = struct{}{}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// Find any permission that directly or transitively depends on curr
		if p, ok := s.catalog[curr]; ok {
			for _, depID := range p.DependentPerms {
				if _, alreadyMarked := toDelete[depID]; !alreadyMarked {
					toDelete[depID] = struct{}{}
					queue = append(queue, depID)
				}
			}
		}
	}

	var remaining []string
	for id := range activeSet {
		if _, deleted := toDelete[id]; !deleted {
			remaining = append(remaining, id)
		}
	}
	sort.Strings(remaining)
	return remaining
}

// ValidatePermissionsForScope filters permissions based on YouTrack scope isolation rules:
// - Global assignment: all permissions apply.
// - Organization assignment: global permissions no longer propagate and have no effect.
// - Project assignment: global and organization-level permissions have no effect.
func (s *Service) ValidatePermissionsForScope(permissionIDs []string, targetScope ScopeLevel) []string {

	var valid []string
	for _, id := range permissionIDs {
		p, ok := s.catalog[id]
		if !ok {
			continue
		}

		switch targetScope {
		case ScopeGlobal:
			// Global assignments allow all scopes to apply
			valid = append(valid, id)
		case ScopeOrganization:
			// Only organization and project permissions take effect
			if p.Scope == ScopeOrganization || p.Scope == ScopeProject {
				valid = append(valid, id)
			}
		case ScopeProject:
			// Only project-level permissions take effect
			if p.Scope == ScopeProject {
				valid = append(valid, id)
			}
		}
	}
	sort.Strings(valid)
	return valid
}

// HasPermission checks if the required permission is present in the list of granted permissions.
func (s *Service) HasPermission(grantedPermissions []string, requiredPermission string) bool {
	for _, p := range grantedPermissions {
		if p == requiredPermission {
			return true
		}
	}
	return false
}

// CheckInherentAccess evaluates inherent rights granted to creators/reporters without explicit permissions:
// - Issue reporters inherit permission to view public fields, update public fields, and add links to their issues.
// - File attachers inherit modifying their files and restricting visibility without Update Attachment.
// - All users can delete files they attached themselves without Delete Attachment.
// - Issue comment creators inherit reading their own comments.
// - Issue work item creators inherit reading their own work items.
// - Article comment creators inherit reading and updating their own comments.
// - Users cannot delete their own issues without Delete Issue.
func (s *Service) CheckInherentAccess(action InherentAction, isAuthorOrReporter bool, hasPermission func(string) bool) bool {
	if !isAuthorOrReporter {
		// Not the author/reporter; must have explicit permission
		return false
	}

	switch action {
	case InherentReadOwnIssuePublicFields, InherentUpdateOwnIssuePublicFields, InherentLinkOwnIssue:
		// Requires CREATE_ISSUE permission in the project
		return hasPermission(PermCreateIssue)

	case InherentModifyOwnAttachment, InherentRestrictOwnAttachment:
		// Requires CREATE_ATTACHMENT_ISSUE permission
		return hasPermission(PermAddAttachment)

	case InherentDeleteOwnAttachment:
		// Any user can delete files they attached themselves
		return true

	case InherentReadOwnIssueComment:
		// Requires CREATE_COMMENT permission
		return hasPermission(PermCreateIssueComment)

	case InherentReadOwnWorkItem:
		// Requires CREATE_WORK_ITEM permission
		return hasPermission(PermCreateWorkItem)

	case InherentReadOwnArticleComment, InherentUpdateOwnArticleComment, InherentDeleteOwnArticleComment:
		// Requires CREATE_ARTICLE_COMMENT permission
		return hasPermission(PermCreateArticleComment)

	default:
		return false
	}
}
