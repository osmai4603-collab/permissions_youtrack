package roles

// IRolesService defines the decoupled, generic contract for managing roles, assignments,
// and resolving aggregated permissions according to YouTrack rules.
type IRolesService interface {
	// Role Lifecycle
	GetRole(id string) (Role, error)
	GetAllRoles() []Role
	CreateRole(actorPerms []string, role Role) (*Role, error)
	UpdateRole(actorPerms []string, id string, name string, description string, permissions []string) (*Role, error)
	DeleteRole(actorPerms []string, id string) error
}
