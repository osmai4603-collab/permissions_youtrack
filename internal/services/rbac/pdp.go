package rbac

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ImpliedPermissionResolver is an optional hook allowing the PDP to expand permissions
// with their implied permissions (e.g. YouTrack implied permissions hierarchy).
type ImpliedPermissionResolver interface {
	ResolveImplied(permissions []string) []string
}

// IPDPEngine defines the contract for an authoritative Policy Decision Point.
type IPDPEngine interface {
	// Evaluate assesses an incoming access request and renders an authoritative decision.
	// It guarantees Fail-Closed (Default Deny) semantics.
	Evaluate(ctx context.Context, req AccessRequest) Decision
}

// PDPEngine is the reference implementation of the Policy Decision Point (PDP)
// according to ANSI/INCITS 359 and NIST SP 800-207 Zero Trust Architecture.
type PDPEngine struct {
	graph           *RoleGraph
	impliedResolver ImpliedPermissionResolver
	contextMgr      *ContextManager
}

// NewPDPEngine creates a new Policy Decision Point engine backed by a role graph.
func NewPDPEngine(graph *RoleGraph, resolver ImpliedPermissionResolver) *PDPEngine {
	if graph == nil {
		graph = NewRoleGraph()
	}
	return &PDPEngine{
		graph:           graph,
		impliedResolver: resolver,
	}
}

// SetImpliedResolver updates or configures the implied permission resolver.
func (p *PDPEngine) SetImpliedResolver(resolver ImpliedPermissionResolver) {
	p.impliedResolver = resolver
}

// SetContextManager binds a contextual ABAC attribute evaluator manager to the PDP.
func (p *PDPEngine) SetContextManager(cm *ContextManager) {
	p.contextMgr = cm
}

// Evaluate evaluates an access request against the role graph and policies.
// Strict Fail-Closed (Default Deny) is enforced at all times.
func (p *PDPEngine) Evaluate(ctx context.Context, req AccessRequest) Decision {
	now := time.Now()

	// 1. Guard check: Permission must not be empty
	targetPerm := strings.TrimSpace(req.Permission)
	if targetPerm == "" {
		return Decision{
			Allowed:     false,
			Reason:      "fail-closed: target permission is empty",
			EvaluatedAt: now,
		}
	}

	// 2. Guard check: At least one role must be provided
	if len(req.RoleIDs) == 0 {
		return Decision{
			Allowed:     false,
			Reason:      "fail-closed: no active roles supplied in request",
			EvaluatedAt: now,
		}
	}

	// 3. Collect all effective permissions across all provided roles
	effectivePermsSet := make(map[string]struct{})
	var rolesEvaluated []string

	for _, roleID := range req.RoleIDs {
		rID := strings.TrimSpace(roleID)
		if rID == "" {
			continue
		}
		rolesEvaluated = append(rolesEvaluated, rID)

		// Fetch direct + inherited permissions from role graph
		perms, err := p.graph.GetEffectivePermissions(rID)
		if err != nil {
			// If role is unknown, skip it (do not fail entirely if other valid roles are present)
			continue
		}

		for _, perm := range perms {
			effectivePermsSet[perm] = struct{}{}
		}
	}

	if len(effectivePermsSet) == 0 {
		return Decision{
			Allowed: false,
			Reason: fmt.Sprintf("no effective permissions resolved for roles: %v",
				rolesEvaluated),
			EvaluatedAt: now,
		}
	}

	// 4. Check for Wildcard / Super-admin permission ("ALL" or "*")
	if _, hasWildcard := effectivePermsSet["*"]; hasWildcard {
		return Decision{
			Allowed:     true,
			Reason:      "granted via global wildcard '*' permission",
			EvaluatedAt: now,
		}
	}
	if _, hasAll := effectivePermsSet["ALL"]; hasAll {
		return Decision{
			Allowed:     true,
			Reason:      "granted via global super-admin 'ALL' permission",
			EvaluatedAt: now,
		}
	}

	// 5. Expand implied permissions if resolver is configured
	if p.impliedResolver != nil {
		rawList := make([]string, 0, len(effectivePermsSet))
		for p := range effectivePermsSet {
			rawList = append(rawList, p)
		}
		expanded := p.impliedResolver.ResolveImplied(rawList)
		for _, ep := range expanded {
			effectivePermsSet[ep] = struct{}{}
		}
	}

	// 6. Direct permission match check
	if _, ok := effectivePermsSet[targetPerm]; ok {
		// Contextual ABAC Evaluation (Hybrid RBAC+ABAC per NIST SP 800-162)
		if p.contextMgr != nil {
			if allowed, reason := p.contextMgr.EvaluateContext(ctx, req); !allowed {
				return Decision{
					Allowed:     false,
					Reason:      fmt.Sprintf("contextual evaluation failed: %s", reason),
					EvaluatedAt: now,
				}
			}
		}

		return Decision{
			Allowed: true,
			Reason: fmt.Sprintf("permission '%s' granted via roles: %v",
				targetPerm, rolesEvaluated),
			EvaluatedAt: now,
		}
	}

	// 7. Default Deny
	return Decision{
		Allowed: false,
		Reason: fmt.Sprintf("access denied: permission '%s' is not held by roles %v",
			targetPerm, rolesEvaluated),
		EvaluatedAt: now,
	}
}
