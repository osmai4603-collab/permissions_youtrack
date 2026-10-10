package rbac_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"youtrack/internal/services/rbac"
)

// BenchmarkDeepHierarchyEvaluation benchmarks resolving permissions through a 10-level inheritance chain.
func BenchmarkDeepHierarchyEvaluation(b *testing.B) {
	graph := rbac.NewRoleGraph()
	const depth = 10

	for i := 0; i < depth; i++ {
		roleID := fmt.Sprintf("LEVEL_%d", i)
		var parents []string
		if i > 0 {
			parents = []string{fmt.Sprintf("LEVEL_%d", i-1)}
		}
		_ = graph.AddRole(&rbac.Role{
			ID:          roleID,
			Permissions: []string{fmt.Sprintf("PERM_%d_A", i), fmt.Sprintf("PERM_%d_B", i)},
			Parents:     parents,
		})
	}

	deepestRole := fmt.Sprintf("LEVEL_%d", depth-1)

	b.ReportAllocs()
	for b.Loop() {
		_, _ = graph.GetEffectivePermissions(deepestRole)
	}
}

// BenchmarkPDPEvaluateParallel measures PDP evaluation throughput across multiple concurrent goroutines.
func BenchmarkPDPEvaluateParallel(b *testing.B) {
	ctx := context.Background()
	graph := rbac.NewRoleGraph()

	_ = graph.AddRole(&rbac.Role{
		ID:          "LEAD_DEV",
		Permissions: []string{"GIT_PUSH", "CODE_REVIEW"},
	})
	_ = graph.AddRole(&rbac.Role{
		ID:          "ARCHITECT",
		Permissions: []string{"DEPLOY_PROD"},
		Parents:     []string{"LEAD_DEV"},
	})

	pdp := rbac.NewPDPEngine(graph, nil)
	req := rbac.AccessRequest{
		SubjectID:  "architect_user",
		RoleIDs:    []string{"ARCHITECT"},
		Permission: "GIT_PUSH",
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			dec := pdp.Evaluate(ctx, req)
			if !dec.Allowed {
				b.Errorf("expected allowed decision in benchmark")
			}
		}
	})
}

// BenchmarkSoDValidation measures the performance of Static Separation of Duties (SSD) validation.
func BenchmarkSoDValidation(b *testing.B) {
	graph := rbac.NewRoleGraph()
	sodMgr := rbac.NewSoDManager(graph)

	_ = graph.AddRole(&rbac.Role{ID: "FINANCE_1"})
	_ = graph.AddRole(&rbac.Role{ID: "FINANCE_2"})
	_ = graph.AddRole(&rbac.Role{ID: "FINANCE_3"})

	_ = sodMgr.AddConstraint(rbac.SoDConstraint{
		ID:         "SSD_RULE",
		Name:       "Finance SSD",
		Type:       rbac.StaticSoD,
		Roles:      []string{"FINANCE_1", "FINANCE_2"},
		MaxAllowed: 1,
	})

	existing := []string{"FINANCE_1"}

	b.ReportAllocs()
	for b.Loop() {
		_ = sodMgr.ValidateUserAssignment(existing, "FINANCE_3")
	}
}

// BenchmarkReversePermissionLookup measures auditing reverse query (Who has permission X?)
// across 1,000 users and complex role hierarchies.
func BenchmarkReversePermissionLookup(b *testing.B) {
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{ID: "BASE_VIEWER", Permissions: []string{"AUDIT_LOG_READ"}})
	_ = graph.AddRole(&rbac.Role{ID: "SEC_ANALYST", Parents: []string{"BASE_VIEWER"}})
	_ = graph.AddRole(&rbac.Role{ID: "CISO", Parents: []string{"SEC_ANALYST"}})

	auditor := rbac.NewSymmetricAuditor(graph, nil)

	// Populate 1,000 assignments
	for i := 0; i < 1000; i++ {
		role := "BASE_VIEWER"
		if i%5 == 0 {
			role = "SEC_ANALYST"
		} else if i%20 == 0 {
			role = "CISO"
		}
		_ = auditor.AssignRole(rbac.UserAssignment{
			UserID:  fmt.Sprintf("user_%d", i),
			RoleIDs: []string{role},
		})
	}

	b.ReportAllocs()
	for b.Loop() {
		users, err := auditor.GetUsersWithPermission("AUDIT_LOG_READ")
		if err != nil || len(users) == 0 {
			b.Fatalf("failed reverse lookup in benchmark")
		}
	}
}

// BenchmarkSessionRoleActivation measures JIT role activation and active permissions resolution.
func BenchmarkSessionRoleActivation(b *testing.B) {
	graph := rbac.NewRoleGraph()
	_ = graph.AddRole(&rbac.Role{ID: "OPS", Permissions: []string{"SERVER_RESTART"}})

	sm := rbac.NewSessionManager(graph, nil)
	sess, _ := sm.CreateSession("admin", []string{}, time.Hour)

	b.ReportAllocs()
	for b.Loop() {
		_ = sm.ActivateRole(sess.ID, "OPS", 5*time.Minute)
		_, _ = sm.GetActivePermissions(sess.ID)
	}
}
