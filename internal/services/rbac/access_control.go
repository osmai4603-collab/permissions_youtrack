package rbac

import "context"

type IAccessControl interface {
	GetRole(ctx context.Context, id string) (*Role, error)
	GetRolesWhereIDs(ctx context.Context, ids []string) ([]*Role, error)
	CreateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, roleID string) (bool, error)
}
