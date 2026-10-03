package roles

import (
	"errors"
)

// Predefined identifiers for virtual principals.
const (
	// AllUsersGroupID is the default virtual group in YouTrack that every registered user belongs to automatically.
	AllUsersGroupID = "ALL_USERS_GROUP"
)

// AssignType identifies whether a role assignment targets a User or a Group (or any extended principal).
type AssignType string

const (
	AssignUser  AssignType = "USER"
	AssignGroup AssignType = "GROUP"
)

// RoleAssignment models an assignment of a role to a principal within a given scope reference.
type RoleAssignment struct {
	ID     string `json:"id"`
	RoleID string `json:"role_id"`
	AssignRef
	ScopeRef
}

// Common error definitions for roles and assignments.
var (
	ErrRoleNotFound             = errors.New("role not found")
	ErrRoleAlreadyExists        = errors.New("role already exists")
	ErrReadOnlyRole             = errors.New("built-in role is read-only and cannot be modified or deleted directly; clone it first")
	ErrPermissionDenied         = errors.New("permission denied: caller lacks Low-level Admin Write (ADMIN_UPDATE_APP)")
	ErrPrivilegeEscalation      = errors.New("privilege escalation: cannot grant or configure permissions the caller does not hold")
	ErrNoProjectScopePermission = errors.New("role cannot be assigned to project: must contain at least one project-scoped permission")
	ErrTargetNotFound           = errors.New("target scope not found")
	ErrAssignmentNotFound       = errors.New("role assignment not found")
	ErrCannotMergeBuiltInRole   = errors.New("cannot merge or delete a built-in role as source")
	ErrInvalidPrincipalType     = errors.New("invalid principal type")
	ErrTargetRequiredForScope   = errors.New("target ID is required for scoped targets")

	// Backward-compatible aliases
	ErrProjectNotFound      = ErrTargetNotFound
	ErrOrganizationNotFound = ErrTargetNotFound
	ErrGroupNotFound        = errors.New("group not found")
)
