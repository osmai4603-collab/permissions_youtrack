package roles

import (
	"slices"

	perms "youtrack/internal/services/permissions_services"
)

// Role models a YouTrack role, which acts as a container for permissions.
type Role struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsImmutable bool     `json:"immutable"`
}

// HasPermission checks if the role contains the given permission ID.
func (rl *Role) HasPermission(permissionID string) bool {
	return slices.Contains(rl.Permissions, permissionID)
}

// HasProjectScopePermission checks if the role contains at least one permission
// that has Project scope in the YouTrack permission catalog.
// This is required by YouTrack's Project Scope Guard.
func (rl *Role) HasProjectScopePermission(permSvc perms.IPermissionService) bool {
	for _, permID := range rl.Permissions {
		if perm := permSvc.GetPermission(permID); perm != nil {
			if perm.Scope == perms.ScopeProject {
				return true
			}
		}
	}
	return false
}
