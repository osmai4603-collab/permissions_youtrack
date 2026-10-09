package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"youtrack/internal/postgres"
	"youtrack/internal/repositories"
	perms "youtrack/internal/services/permissions_services"
	"youtrack/internal/services/rbac"
	"youtrack/migrations"
)

func TestPostgresIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := postgres.DefaultConfig()
	cfg.Database = getEnv("TEST_DB_NAME", "youtrack_test")

	// Initialize pool
	pool, err := postgres.NewPool(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	// Run migrations
	if err := postgres.RunMigrations(ctx, pool, migrations.FS, ".", nil); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Health check
	health := postgres.NewHealthChecker(pool)
	if err := health.Check(ctx); err != nil {
		t.Fatalf("health check failed: %v", err)
	}

	// Repository CRUD test
	permSvc := perms.NewService()
	repo := repositories.NewRolesService(permSvc, pool)

	// Fetch SYSTEM_ADMIN seeded role
	sysAdmin, err := repo.GetRole(ctx, "SYSTEM_ADMIN")
	if err != nil {
		t.Fatalf("failed to get SYSTEM_ADMIN role: %v", err)
	}
	if sysAdmin.Name != "System Admin" {
		t.Errorf("expected role name 'System Admin', got %q", sysAdmin.Name)
	}

	// Test CreateRole
	testRole := &rbac.Role{
		ID:          "TEST_ROLE_INTEGRATION",
		Name:        "Test Role Integration",
		Description: "Integration test role",
		Permissions: []string{"READ_USER"},
		IsImmutable: false,
	}

	created, err := repo.CreateRole(ctx, testRole)
	if err != nil {
		t.Fatalf("failed to create test role: %v", err)
	}
	if created.ID != testRole.ID {
		t.Errorf("expected role ID %q, got %q", testRole.ID, created.ID)
	}

	// Test GetRolesWhereIDs
	roles, err := repo.GetRolesWhereIDs(ctx, []string{"SYSTEM_ADMIN", "TEST_ROLE_INTEGRATION"})
	if err != nil {
		t.Fatalf("failed to get roles by IDs: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}

	// Clean up test role
	deleted, err := repo.DeleteRole(ctx, "TEST_ROLE_INTEGRATION")
	if err != nil {
		t.Fatalf("failed to delete test role: %v", err)
	}
	if !deleted {
		t.Error("expected delete to return true")
	}
}

func TestProjectScopeRoleQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg := postgres.DefaultConfig()
	cfg.Database = getEnv("TEST_DB_NAME", "youtrack_test")
	pool, err := postgres.NewPool(ctx, cfg, nil)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := postgres.RunMigrations(ctx, pool, migrations.FS, ".", nil); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin fixture transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	fixtureID := fmt.Sprintf("project-role-%d", time.Now().UnixNano())
	userID := fixtureID + "-user"
	projectID := fixtureID + "-project"
	otherProjectID := fixtureID + "-other-project"
	directGroupID := fixtureID + "-direct-group"
	roleGroupID := fixtureID + "-role-group"
	middleGroupID := fixtureID + "-middle-group"
	leafGroupID := fixtureID + "-leaf-group"

	if _, err := tx.Exec(ctx, `INSERT INTO users (id, login) VALUES ($1, $2)`, userID, fixtureID); err != nil {
		t.Fatalf("insert fixture user: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO projects (id, name, short_name) VALUES ($1, $2, $3)`, projectID, fixtureID, fixtureID); err != nil {
		t.Fatalf("insert fixture project: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO projects (id, name, short_name) VALUES ($1, $2, $3)`, otherProjectID, fixtureID+" other", fixtureID+"-other"); err != nil {
		t.Fatalf("insert other fixture project: %v", err)
	}

	for _, group := range []struct {
		id   string
		name string
	}{
		{directGroupID, fixtureID + " direct group"},
		{roleGroupID, fixtureID + " role group"},
		{middleGroupID, fixtureID + " middle group"},
		{leafGroupID, fixtureID + " leaf group"},
	} {
		if _, err := tx.Exec(ctx, `INSERT INTO groups (id, name) VALUES ($1, $2)`, group.id, group.name); err != nil {
			t.Fatalf("insert fixture group %q: %v", group.id, err)
		}
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO group_members (group_id, member_type, user_id) VALUES
			($1, 'USER', $2),
			($3, 'USER', $2)
	`, directGroupID, userID, leafGroupID); err != nil {
		t.Fatalf("insert direct fixture group memberships: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO group_members (group_id, member_type, member_group_id) VALUES ($1, 'GROUP', $2)`, roleGroupID, middleGroupID); err != nil {
		t.Fatalf("nest middle fixture group: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO group_members (group_id, member_type, member_group_id) VALUES ($1, 'GROUP', $2)`, middleGroupID, leafGroupID); err != nil {
		t.Fatalf("nest leaf fixture group: %v", err)
	}

	var teamID string
	if err := tx.QueryRow(ctx, `SELECT team_id FROM projects WHERE id = $1`, projectID).Scan(&teamID); err != nil {
		t.Fatalf("read fixture project team: %v", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO group_members (group_id, member_type, user_id) VALUES ($1, 'USER', $2)`, teamID, userID); err != nil {
		t.Fatalf("add user to fixture project team: %v", err)
	}

	for _, assignment := range []struct {
		roleID  string
		project string
		user    string
		group   string
	}{
		{"PROJECT_ADMIN", projectID, userID, ""},
		{"SYSTEM_ADMIN", projectID, "", directGroupID},
		{"DEVELOPER", projectID, "", roleGroupID},
		{"OBSERVER", otherProjectID, userID, ""},
	} {
		if _, err := tx.Exec(ctx, `
			INSERT INTO assigned_roles (role_id, scope_type, project_id, user_id, group_id)
			VALUES ($1, 'PROJECT', $2, NULLIF($3, ''), NULLIF($4, ''))
		`, assignment.roleID, assignment.project, assignment.user, assignment.group); err != nil {
			t.Fatalf("insert fixture role assignment %q: %v", assignment.roleID, err)
		}
	}

	query, err := os.ReadFile("../../queries/get_permissions_by_project_scope.sql")
	if err != nil {
		t.Fatalf("read project scope role query: %v", err)
	}
	rows, err := tx.Query(ctx, string(query), projectID, userID)
	if err != nil {
		t.Fatalf("execute project scope role query: %v", err)
	}
	defer rows.Close()

	got := make(map[string]bool)
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			t.Fatalf("scan project scope role: %v", err)
		}
		got[roleID] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate project scope roles: %v", err)
	}

	for _, roleID := range []string{"PROJECT_ADMIN", "SYSTEM_ADMIN", "DEVELOPER", "CONTRIBUTOR"} {
		if !got[roleID] {
			t.Errorf("expected role %q from direct, group, nested-group, or project-team assignment", roleID)
		}
	}
	if got["OBSERVER"] {
		t.Error("did not expect a role assigned in another project")
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
