package rbac_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"youtrack/internal/services/rbac"
)

func TestHierarchy_BasicInheritance(t *testing.T) {
	graph := rbac.NewRoleGraph()

	// 1. Create Base Role (Viewer)
	viewer := &rbac.Role{
		ID:          "VIEWER",
		Name:        "Viewer",
		Permissions: []string{"READ_ISSUE", "READ_PROJECT"},
	}
	if err := graph.AddRole(viewer); err != nil {
		t.Fatalf("failed to add VIEWER: %v", err)
	}

	// 2. Create Intermediate Role (Editor) inheriting from Viewer
	editor := &rbac.Role{
		ID:          "EDITOR",
		Name:        "Editor",
		Permissions: []string{"UPDATE_ISSUE"},
		Parents:     []string{"VIEWER"},
	}
	if err := graph.AddRole(editor); err != nil {
		t.Fatalf("failed to add EDITOR: %v", err)
	}

	// 3. Create Admin Role inheriting from Editor
	admin := &rbac.Role{
		ID:          "ADMIN",
		Name:        "Admin",
		Permissions: []string{"DELETE_ISSUE"},
		Parents:     []string{"EDITOR"},
	}
	if err := graph.AddRole(admin); err != nil {
		t.Fatalf("failed to add ADMIN: %v", err)
	}

	// Verify Admin effective permissions: should include DELETE_ISSUE, UPDATE_ISSUE, READ_ISSUE, READ_PROJECT
	perms, err := graph.GetEffectivePermissions("ADMIN")
	if err != nil {
		t.Fatalf("failed to get effective permissions for ADMIN: %v", err)
	}

	expected := []string{"DELETE_ISSUE", "READ_ISSUE", "READ_PROJECT", "UPDATE_ISSUE"}
	if len(perms) != len(expected) {
		t.Fatalf("expected %d permissions, got %d: %v", len(expected), len(perms), perms)
	}
	for i, exp := range expected {
		if perms[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, perms[i])
		}
	}

	// Verify HasPermission
	hasRead, err := graph.HasPermission("ADMIN", "READ_ISSUE")
	if err != nil || !hasRead {
		t.Errorf("ADMIN should have READ_ISSUE permission")
	}

	hasDelete, err := graph.HasPermission("VIEWER", "DELETE_ISSUE")
	if err != nil || hasDelete {
		t.Errorf("VIEWER should NOT have DELETE_ISSUE permission")
	}
}

func TestHierarchy_DiamondInheritance(t *testing.T) {
	graph := rbac.NewRoleGraph()

	// A: Base
	if err := graph.AddRole(&rbac.Role{ID: "A", Permissions: []string{"PERM_A"}}); err != nil {
		t.Fatal(err)
	}
	// B -> A
	if err := graph.AddRole(&rbac.Role{ID: "B", Permissions: []string{"PERM_B"}, Parents: []string{"A"}}); err != nil {
		t.Fatal(err)
	}
	// C -> A
	if err := graph.AddRole(&rbac.Role{ID: "C", Permissions: []string{"PERM_C"}, Parents: []string{"A"}}); err != nil {
		t.Fatal(err)
	}
	// D -> B and C (Diamond)
	if err := graph.AddRole(&rbac.Role{ID: "D", Permissions: []string{"PERM_D"}, Parents: []string{"B", "C"}}); err != nil {
		t.Fatal(err)
	}

	perms, err := graph.GetEffectivePermissions("D")
	if err != nil {
		t.Fatal(err)
	}

	// Expect PERM_A (deduplicated), PERM_B, PERM_C, PERM_D
	expected := []string{"PERM_A", "PERM_B", "PERM_C", "PERM_D"}
	if len(perms) != len(expected) {
		t.Fatalf("expected %d permissions, got %d: %v", len(expected), len(perms), perms)
	}
	for i, exp := range expected {
		if perms[i] != exp {
			t.Errorf("at index %d: expected %s, got %s", i, exp, perms[i])
		}
	}
}

func TestHierarchy_CycleDetection(t *testing.T) {
	graph := rbac.NewRoleGraph()

	_ = graph.AddRole(&rbac.Role{ID: "R1"})
	_ = graph.AddRole(&rbac.Role{ID: "R2"})
	_ = graph.AddRole(&rbac.Role{ID: "R3"})

	// Self inheritance
	if err := graph.AddInheritance("R1", "R1"); !errors.Is(err, rbac.ErrCircularInheritance) {
		t.Errorf("expected ErrCircularInheritance for self-loop, got: %v", err)
	}

	// R1 -> R2
	if err := graph.AddInheritance("R1", "R2"); err != nil {
		t.Fatalf("failed to add R1 -> R2: %v", err)
	}

	// Direct cycle: R2 -> R1
	if err := graph.AddInheritance("R2", "R1"); !errors.Is(err, rbac.ErrCircularInheritance) {
		t.Errorf("expected ErrCircularInheritance for direct cycle, got: %v", err)
	}

	// R2 -> R3
	if err := graph.AddInheritance("R2", "R3"); err != nil {
		t.Fatalf("failed to add R2 -> R3: %v", err)
	}

	// Transitive cycle: R3 -> R1 (would make R1 -> R2 -> R3 -> R1)
	if err := graph.AddInheritance("R3", "R1"); !errors.Is(err, rbac.ErrCircularInheritance) {
		t.Errorf("expected ErrCircularInheritance for transitive cycle, got: %v", err)
	}
}

func TestHierarchy_ImmutableRoleProtection(t *testing.T) {
	graph := rbac.NewRoleGraph()

	sysAdmin := &rbac.Role{
		ID:          "SYSTEM_ADMIN",
		Name:        "System Administrator",
		IsImmutable: true,
		Permissions: []string{"ALL"},
	}
	if err := graph.AddRole(sysAdmin); err != nil {
		t.Fatal(err)
	}

	// Deletion must fail
	err := graph.DeleteRole("SYSTEM_ADMIN")
	if !errors.Is(err, rbac.ErrRoleImmutable) {
		t.Errorf("expected ErrRoleImmutable when deleting immutable role, got: %v", err)
	}

	// Update must fail
	updated := &rbac.Role{
		ID:          "SYSTEM_ADMIN",
		Name:        "Modified Admin",
		IsImmutable: true,
	}
	err = graph.UpdateRole(updated)
	if !errors.Is(err, rbac.ErrRoleImmutable) {
		t.Errorf("expected ErrRoleImmutable when updating immutable role, got: %v", err)
	}
}

func TestHierarchy_CloneIndependence(t *testing.T) {
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{ID: "PARENT", Permissions: []string{"READ"}})
	_ = graph.AddRole(&rbac.Role{ID: "CHILD", Permissions: []string{"WRITE"}, Parents: []string{"PARENT"}})

	clone := graph.Clone()

	// Modify clone
	_ = clone.AddRole(&rbac.Role{ID: "NEW_ROLE", Permissions: []string{"EXEC"}})
	_ = clone.AddInheritance("NEW_ROLE", "CHILD")

	// Original graph should not have NEW_ROLE
	if _, err := graph.GetRole("NEW_ROLE"); !errors.Is(err, rbac.ErrRoleNotFound) {
		t.Errorf("original graph should not be affected by changes to clone")
	}

	// Clone should have NEW_ROLE
	if _, err := clone.GetRole("NEW_ROLE"); err != nil {
		t.Errorf("clone should contain NEW_ROLE")
	}
}

func TestHierarchy_Concurrency(t *testing.T) {
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{ID: "ROOT", Permissions: []string{"ROOT_PERM"}})

	var wg sync.WaitGroup
	const workers = 30

	// Concurrent role creation and querying
	for i := 0; i < workers; i++ {
		wg.Add(1)
		roleID := fmt.Sprintf("WORKER_ROLE_%d", i)
		go func(id string) {
			defer wg.Done()
			_ = graph.AddRole(&rbac.Role{
				ID:          id,
				Permissions: []string{"WORKER_PERM"},
				Parents:     []string{"ROOT"},
			})
			_, _ = graph.GetEffectivePermissions(id)
			_, _ = graph.GetInheritedRoles(id)
		}(roleID)
	}

	wg.Wait()

	// Verify all roles created
	all := graph.GetAllRoles()
	if len(all) != workers+1 {
		t.Errorf("expected %d roles, got %d", workers+1, len(all))
	}
}

func BenchmarkGetEffectivePermissions(b *testing.B) {
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{ID: "L0", Permissions: []string{"P0", "P1"}})
	_ = graph.AddRole(&rbac.Role{ID: "L1", Permissions: []string{"P2"}, Parents: []string{"L0"}})
	_ = graph.AddRole(&rbac.Role{ID: "L2", Permissions: []string{"P3"}, Parents: []string{"L1"}})
	_ = graph.AddRole(&rbac.Role{ID: "L3", Permissions: []string{"P4"}, Parents: []string{"L2"}})
	_ = graph.AddRole(&rbac.Role{ID: "L4", Permissions: []string{"P5"}, Parents: []string{"L3"}})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = graph.GetEffectivePermissions("L4")
	}
}
