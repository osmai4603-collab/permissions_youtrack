package repositories

import (
	"context"
	"youtrack/internal/postgres"
	perms "youtrack/internal/services/permissions_services"
	"youtrack/internal/services/rbac"
)

type RolesService struct {
	permSRV perms.IPermissionService
	db      postgres.DBTX
}

func NewRolesService(srv perms.IPermissionService, db postgres.DBTX) *RolesService {
	return &RolesService{
		permSRV: srv,
		db:      db,
	}
}

func (srv *RolesService) GetRole(ctx context.Context, id string) (*rbac.Role, error) {
	const op = "repositories.RolesService.GetRole"

	query := `SELECT id, name, description, permissions, immutable FROM roles WHERE id = $1`
	var role rbac.Role
	err := srv.db.QueryRow(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &role.Permissions, &role.IsImmutable,
	)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	return &role, nil
}

func (srv *RolesService) GetRolesWhereIDs(ctx context.Context, ids []string) ([]*rbac.Role, error) {
	const op = "repositories.RolesService.GetRolesWhereIDs"

	if len(ids) == 0 {
		return []*rbac.Role{}, nil
	}

	query := `SELECT id, name, description, permissions, immutable FROM roles WHERE id = ANY($1)`
	rows, err := srv.db.Query(ctx, query, ids)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}
	defer rows.Close()

	var roles []*rbac.Role
	for rows.Next() {
		var role rbac.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.Permissions, &role.IsImmutable); err != nil {
			return nil, postgres.TranslateError(op, err)
		}
		roles = append(roles, &role)
	}

	if err := rows.Err(); err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	return roles, nil
}

func (srv *RolesService) CreateRole(ctx context.Context, role *rbac.Role) (*rbac.Role, error) {
	const op = "repositories.RolesService.CreateRole"

	query := `INSERT INTO roles (id, name, description, permissions, immutable) VALUES ($1, $2, $3, $4, $5)`
	_, err := srv.db.Exec(ctx, query, role.ID, role.Name, role.Description, role.Permissions, role.IsImmutable)
	if err != nil {
		return nil, postgres.TranslateError(op, err)
	}

	return role, nil
}

func (srv *RolesService) DeleteRole(ctx context.Context, roleID string) (bool, error) {
	const op = "repositories.RolesService.DeleteRole"

	query := `DELETE FROM roles WHERE id = $1`
	tag, err := srv.db.Exec(ctx, query, roleID)
	if err != nil {
		return false, postgres.TranslateError(op, err)
	}

	return tag.RowsAffected() > 0, nil
}
