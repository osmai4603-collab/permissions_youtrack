package rbac_test

import (
	"context"
	"errors"
	"testing"

	"youtrack/internal/services/rbac"
)

// mockImpliedResolver simulates an external catalog resolving implied permissions.
type mockImpliedResolver struct {
	impliedMap map[string][]string
}

func (m *mockImpliedResolver) ResolveImplied(permissions []string) []string {
	resultMap := make(map[string]struct{})
	for _, p := range permissions {
		resultMap[p] = struct{}{}
		if implied, ok := m.impliedMap[p]; ok {
			for _, imp := range implied {
				resultMap[imp] = struct{}{}
			}
		}
	}
	var res []string
	for p := range resultMap {
		res = append(res, p)
	}
	return res
}

func TestPDP_DirectAndInheritedEvaluation(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph()

	// Base role: Viewer
	_ = graph.AddRole(&rbac.Role{
		ID:          "VIEWER",
		Permissions: []string{"READ_ISSUE"},
	})

	// Child role: Editor (inherits Viewer)
	_ = graph.AddRole(&rbac.Role{
		ID:          "EDITOR",
		Permissions: []string{"UPDATE_ISSUE"},
		Parents:     []string{"VIEWER"},
	})

	pdp := rbac.NewPDPEngine(graph, nil)

	// 1. Direct permission match on Viewer
	dec := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "alice",
		RoleIDs:    []string{"VIEWER"},
		Permission: "READ_ISSUE",
	})
	if !dec.Allowed {
		t.Errorf("expected allowed for direct permission, got reason: %s", dec.Reason)
	}

	// 2. Inherited permission match on Editor (holds READ_ISSUE via VIEWER)
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "bob",
		RoleIDs:    []string{"EDITOR"},
		Permission: "READ_ISSUE",
	})
	if !dec.Allowed {
		t.Errorf("expected allowed for inherited permission, got reason: %s", dec.Reason)
	}

	// 3. Permission not granted (VIEWER attempting UPDATE_ISSUE)
	dec = pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "alice",
		RoleIDs:    []string{"VIEWER"},
		Permission: "UPDATE_ISSUE",
	})
	if dec.Allowed {
		t.Errorf("expected denied for unauthorized permission")
	}
}

func TestPDP_ImpliedPermissions(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph()

	_ = graph.AddRole(&rbac.Role{
		ID:          "PROJECT_MANAGER",
		Permissions: []string{"UPDATE_PROJECT"},
	})

	// When UPDATE_PROJECT is granted, READ_PROJECT is implied
	resolver := &mockImpliedResolver{
		impliedMap: map[string][]string{
			"UPDATE_PROJECT": {"READ_PROJECT", "READ_ISSUES"},
		},
	}

	pdp := rbac.NewPDPEngine(graph, resolver)

	dec := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "pm_user",
		RoleIDs:    []string{"PROJECT_MANAGER"},
		Permission: "READ_PROJECT",
	})
	if !dec.Allowed {
		t.Errorf("expected allowed for implied permission READ_PROJECT, reason: %s", dec.Reason)
	}
}

func TestPDP_SuperAdminWildcard(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph()

	_ = graph.AddRole(&rbac.Role{
		ID:          "SUPER_ADMIN",
		Permissions: []string{"*"},
	})
	_ = graph.AddRole(&rbac.Role{
		ID:          "SYS_ALL",
		Permissions: []string{"ALL"},
	})

	pdp := rbac.NewPDPEngine(graph, nil)

	// Test wildcard '*'
	decStar := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "root",
		RoleIDs:    []string{"SUPER_ADMIN"},
		Permission: "ANY_CUSTOM_ACTION_UNKNOWN",
	})
	if !decStar.Allowed {
		t.Errorf("expected wildcard '*' to grant any permission, reason: %s", decStar.Reason)
	}

	// Test wildcard 'ALL'
	decAll := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "admin",
		RoleIDs:    []string{"SYS_ALL"},
		Permission: "DELETE_ANYTHING",
	})
	if !decAll.Allowed {
		t.Errorf("expected 'ALL' permission to grant any permission, reason: %s", decAll.Reason)
	}
}

func TestPDP_FailClosedGuards(t *testing.T) {
	ctx := context.Background()
	pdp := rbac.NewPDPEngine(nil, nil)

	// 1. Empty permission
	decEmptyPerm := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "user_1",
		RoleIDs:    []string{"SOME_ROLE"},
		Permission: "",
	})
	if decEmptyPerm.Allowed {
		t.Errorf("expected Fail-Closed for empty permission")
	}

	// 2. Empty roles
	decEmptyRoles := pdp.Evaluate(ctx, rbac.AccessRequest{
		SubjectID:  "user_1",
		RoleIDs:    []string{},
		Permission: "READ_PROJECT",
	})
	if decEmptyRoles.Allowed {
		t.Errorf("expected Fail-Closed for empty role list")
	}
}

func TestPEP_Enforcer(t *testing.T) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{
		ID:          "DEVELOPER",
		Permissions: []string{"CODE_READ", "CODE_WRITE"},
	})

	pdp := rbac.NewPDPEngine(graph, nil)
	enforcer := rbac.NewEnforcer(pdp)

	// 1. Authorize success
	err := enforcer.Authorize(ctx, rbac.AccessRequest{
		SubjectID:  "dev1",
		RoleIDs:    []string{"DEVELOPER"},
		Permission: "CODE_READ",
	})
	if err != nil {
		t.Errorf("expected authorization success, got: %v", err)
	}

	// 2. Authorize failure
	err = enforcer.Authorize(ctx, rbac.AccessRequest{
		SubjectID:  "dev1",
		RoleIDs:    []string{"DEVELOPER"},
		Permission: "DEPLOY_PROD",
	})
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied for unauthorized request, got: %v", err)
	}

	// 3. AuthorizeAny: one matches -> Success
	err = enforcer.AuthorizeAny(ctx, "dev1", []string{"DEVELOPER"}, "DEPLOY_PROD", "CODE_WRITE")
	if err != nil {
		t.Errorf("AuthorizeAny should succeed if at least one matches, got: %v", err)
	}

	// 4. AuthorizeAny: none matches -> Failure
	err = enforcer.AuthorizeAny(ctx, "dev1", []string{"DEVELOPER"}, "DEPLOY_PROD", "DROP_DATABASE")
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("expected ErrAccessDenied when none match, got: %v", err)
	}

	// 5. AuthorizeAll: all match -> Success
	err = enforcer.AuthorizeAll(ctx, "dev1", []string{"DEVELOPER"}, "CODE_READ", "CODE_WRITE")
	if err != nil {
		t.Errorf("AuthorizeAll should succeed when all match, got: %v", err)
	}

	// 6. AuthorizeAll: one missing -> Failure
	err = enforcer.AuthorizeAll(ctx, "dev1", []string{"DEVELOPER"}, "CODE_READ", "DEPLOY_PROD")
	if !errors.Is(err, rbac.ErrAccessDenied) {
		t.Errorf("AuthorizeAll should fail if one is missing, got: %v", err)
	}
}

func BenchmarkPDPEvaluate(b *testing.B) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{ID: "R0", Permissions: []string{"P0", "P1"}})
	_ = graph.AddRole(&rbac.Role{ID: "R1", Permissions: []string{"P2"}, Parents: []string{"R0"}})
	_ = graph.AddRole(&rbac.Role{ID: "R2", Permissions: []string{"P3"}, Parents: []string{"R1"}})

	pdp := rbac.NewPDPEngine(graph, nil)
	req := rbac.AccessRequest{
		SubjectID:  "benchmark_user",
		RoleIDs:    []string{"R2"},
		Permission: "P0",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pdp.Evaluate(ctx, req)
	}
}
