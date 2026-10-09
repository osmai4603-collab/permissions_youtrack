package rbac

import "slices"

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
