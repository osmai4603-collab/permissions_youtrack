package roles

// import (
// 	"sort"
// )

// // GetEffectivePermissions computes the cumulative effective permissions for a principal
// // (User, Group, etc.) in a target scope according to YouTrack rules:
// // - Resolves all associated identities via the PrincipalHierarchyProvider SPI.
// // - Resolves the target scope and its parent hierarchy via the ScopeHierarchyProvider SPI.
// // - Aggregates permissions from all matching assignments.
// // - Enforces YouTrack scope disregard rules per assignment scope level:
// //   - Global assignments preserve all permission scopes.
// //   - Organization assignments disregard global permissions.
// //   - Project assignments disregard global and organization permissions.
// func (svc *Service) GetEffectivePermissions(assign AssignRef, targetScope ScopeRef) ([]string, error) {
// 	svc.mu.RLock()
// 	defer svc.mu.RUnlock()

// 	if err := svc.scopeProvider.ValidateScope(targetScope); err != nil {
// 		return nil, err
// 	}

// 	// 1. Resolve all identities (direct ID, groups, virtual all users, etc.)
// 	identities, err := svc.idProvider.ResolveIdentities(assign.AssignType, assign.AssignerID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	identitySet := make(map[string]struct{}, len(identities))
// 	for _, id := range identities {
// 		identitySet[id] = struct{}{}
// 	}

// 	// 2. Resolve parent scopes hierarchy
// 	parentScopes, err := svc.scopeProvider.GetParentScopes(targetScope)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// A role assignment applies if its scope is targetScope, any parentScope, or GlobalScope
// 	applicableScopes := make(map[string]ScopeRef)
// 	applicableScopes[targetScope.String()] = targetScope
// 	applicableScopes[GlobalScope().String()] = GlobalScope()
// 	for _, ps := range parentScopes {
// 		applicableScopes[ps.String()] = ps
// 	}

// 	effectivePermsSet := make(map[string]struct{})

// 	// 3. Process role assignments
// 	for _, assignment := range svc.assignments {
// 		if _, matchesIdentity := identitySet[assignment.AssignerID]; !matchesIdentity {
// 			continue
// 		}

// 		role, exists := svc.roles[assignment.RoleID]
// 		if !exists {
// 			continue
// 		}

// 		// Check if assignment's scope applies to the target scope hierarchy
// 		if _, applies := applicableScopes[assignment.ScopeRef.String()]; !applies {
// 			continue
// 		}

// 		// Apply YouTrack Disregard Rules based on the assignment scope level
// 		filteredPerms := svc.permSvc.ValidatePermissionsForScope(role.Permissions, assignment.Scope)
// 		for _, permID := range filteredPerms {
// 			effectivePermsSet[permID] = struct{}{}
// 		}
// 	}

// 	out := make([]string, 0, len(effectivePermsSet))
// 	for pid := range effectivePermsSet {
// 		out = append(out, pid)
// 	}
// 	sort.Strings(out)
// 	return out, nil
// }

// // HasPermission checks if a principal possesses a specific permission in a target scope.
// func (svc *Service) HasPermission(assign AssignRef, scope ScopeRef, requiredPerm string) (bool, error) {
// 	effective, err := svc.GetEffectivePermissions(assign, scope)
// 	if err != nil {
// 		return false, err
// 	}
// 	return svc.permSvc.HasPermission(effective, requiredPerm), nil
// }
