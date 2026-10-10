package rbac

import (
	"fmt"
	"sort"
	"sync"
)

// RoleGraph represents an in-memory, thread-safe Directed Acyclic Graph (DAG)
// governing role inheritance according to ANSI/INCITS 359 Hierarchical RBAC.
type RoleGraph struct {
	mu sync.RWMutex

	// roles maps each RoleID to its definition.
	roles map[RoleID]*Role

	// parents maps a child RoleID to its immediate parent RoleIDs (roles it inherits from).
	// Edge: child -> parent (Child inherits all permissions of Parent)
	parents map[RoleID]map[RoleID]struct{}

	// children maps a parent RoleID to its immediate child RoleIDs (roles that inherit from it).
	// Edge: parent -> child
	children map[RoleID]map[RoleID]struct{}
}

// NewRoleGraph creates and initializes an empty thread-safe role hierarchy graph.
func NewRoleGraph() *RoleGraph {
	return &RoleGraph{
		roles:    make(map[RoleID]*Role),
		parents:  make(map[RoleID]map[RoleID]struct{}),
		children: make(map[RoleID]map[RoleID]struct{}),
	}
}

// AddRole inserts a new role into the hierarchy graph.
// If the role defines Parents, each inheritance edge is validated against cycles.
func (g *RoleGraph) AddRole(role *Role) error {
	if role == nil || role.ID == "" {
		return fmt.Errorf("%w: role and role ID must not be empty", ErrInvalidInput)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.roles[role.ID]; exists {
		return fmt.Errorf("%w: %s", ErrRoleAlreadyExists, role.ID)
	}

	// Make a defensive copy
	clonedRole := cloneRole(role)

	// Validate parent roles exist and won't create cycles
	for _, parentID := range clonedRole.Parents {
		if parentID == clonedRole.ID {
			return fmt.Errorf("%w: role %s cannot inherit from itself", ErrCircularInheritance, clonedRole.ID)
		}
		if _, exists := g.roles[parentID]; !exists {
			return fmt.Errorf("%w: parent role %s does not exist", ErrRoleNotFound, parentID)
		}
		if g.hasPath(parentID, clonedRole.ID) {
			return fmt.Errorf("%w: path exists from %s to %s", ErrCircularInheritance, parentID, clonedRole.ID)
		}
	}

	// Register role
	g.roles[clonedRole.ID] = clonedRole
	if _, ok := g.parents[clonedRole.ID]; !ok {
		g.parents[clonedRole.ID] = make(map[RoleID]struct{})
	}
	if _, ok := g.children[clonedRole.ID]; !ok {
		g.children[clonedRole.ID] = make(map[RoleID]struct{})
	}

	// Add edges
	for _, parentID := range clonedRole.Parents {
		g.parents[clonedRole.ID][parentID] = struct{}{}
		if _, ok := g.children[parentID]; !ok {
			g.children[parentID] = make(map[RoleID]struct{})
		}
		g.children[parentID][clonedRole.ID] = struct{}{}
	}

	return nil
}

// GetRole retrieves a copy of the specified role by ID.
func (g *RoleGraph) GetRole(id RoleID) (*Role, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	role, exists := g.roles[id]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, id)
	}
	return cloneRole(role), nil
}

// GetAllRoles returns copies of all registered roles sorted by ID.
func (g *RoleGraph) GetAllRoles() []*Role {
	g.mu.RLock()
	defer g.mu.RUnlock()

	res := make([]*Role, 0, len(g.roles))
	for _, r := range g.roles {
		res = append(res, cloneRole(r))
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res
}

// UpdateRole updates an existing role. Immutable roles cannot be altered.
func (g *RoleGraph) UpdateRole(updated *Role) error {
	if updated == nil || updated.ID == "" {
		return fmt.Errorf("%w: role and role ID must not be empty", ErrInvalidInput)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	existing, exists := g.roles[updated.ID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrRoleNotFound, updated.ID)
	}

	if existing.IsImmutable {
		return fmt.Errorf("%w: cannot update immutable role %s", ErrRoleImmutable, updated.ID)
	}

	// Validate parent changes
	for _, pID := range updated.Parents {
		if pID == updated.ID {
			return fmt.Errorf("%w: role %s cannot inherit from itself", ErrCircularInheritance, updated.ID)
		}
		if _, exists := g.roles[pID]; !exists {
			return fmt.Errorf("%w: parent role %s does not exist", ErrRoleNotFound, pID)
		}
		// Temporarily check if path exists excluding current direct edge
		if g.hasPathExcludingEdge(pID, updated.ID, updated.ID, pID) {
			return fmt.Errorf("%w: cycle would be formed between %s and %s", ErrCircularInheritance, updated.ID, pID)
		}
	}

	// Remove old parent edges
	if oldParents, ok := g.parents[updated.ID]; ok {
		for pID := range oldParents {
			delete(g.children[pID], updated.ID)
		}
	}
	g.parents[updated.ID] = make(map[RoleID]struct{})

	// Re-add new parent edges
	for _, pID := range updated.Parents {
		g.parents[updated.ID][pID] = struct{}{}
		g.children[pID][updated.ID] = struct{}{}
	}

	g.roles[updated.ID] = cloneRole(updated)
	return nil
}

// DeleteRole removes a role from the graph. Immutable roles cannot be deleted.
func (g *RoleGraph) DeleteRole(id RoleID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	role, exists := g.roles[id]
	if !exists {
		return fmt.Errorf("%w: %s", ErrRoleNotFound, id)
	}

	if role.IsImmutable {
		return fmt.Errorf("%w: cannot delete immutable role %s", ErrRoleImmutable, id)
	}

	// Remove as child from parents
	if parents, ok := g.parents[id]; ok {
		for pID := range parents {
			delete(g.children[pID], id)
		}
	}

	// Remove as parent from children
	if children, ok := g.children[id]; ok {
		for cID := range children {
			delete(g.parents[cID], id)
			// Also update Parents slice on the child role model
			if childRole, ok := g.roles[cID]; ok {
				newParents := make([]RoleID, 0, len(childRole.Parents))
				for _, p := range childRole.Parents {
					if p != id {
						newParents = append(newParents, p)
					}
				}
				childRole.Parents = newParents
			}
		}
	}

	delete(g.parents, id)
	delete(g.children, id)
	delete(g.roles, id)

	return nil
}

// AddInheritance registers an inheritance relationship: childRoleID inherits from parentRoleID.
// If the edge introduces a cycle, ErrCircularInheritance is returned.
func (g *RoleGraph) AddInheritance(childRoleID, parentRoleID RoleID) error {
	if childRoleID == "" || parentRoleID == "" {
		return fmt.Errorf("%w: role IDs must not be empty", ErrInvalidInput)
	}
	if childRoleID == parentRoleID {
		return fmt.Errorf("%w: role %s cannot inherit from itself", ErrCircularInheritance, childRoleID)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	childRole, childExists := g.roles[childRoleID]
	if !childExists {
		return fmt.Errorf("%w: child role %s", ErrRoleNotFound, childRoleID)
	}
	_, parentExists := g.roles[parentRoleID]
	if !parentExists {
		return fmt.Errorf("%w: parent role %s", ErrRoleNotFound, parentRoleID)
	}

	// Check if already inherited
	if _, alreadyInherited := g.parents[childRoleID][parentRoleID]; alreadyInherited {
		return nil // idempotent
	}

	// Cycle check: If parentRoleID already reaches childRoleID, adding child -> parent creates a cycle
	if g.hasPath(parentRoleID, childRoleID) {
		return fmt.Errorf("%w: adding %s -> %s creates a circular dependency", ErrCircularInheritance, childRoleID, parentRoleID)
	}

	// Add edge
	g.parents[childRoleID][parentRoleID] = struct{}{}
	g.children[parentRoleID][childRoleID] = struct{}{}

	// Update role model
	childRole.Parents = append(childRole.Parents, parentRoleID)

	return nil
}

// RemoveInheritance removes an inheritance relationship between child and parent.
func (g *RoleGraph) RemoveInheritance(childRoleID, parentRoleID RoleID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	childRole, exists := g.roles[childRoleID]
	if !exists {
		return fmt.Errorf("%w: child role %s", ErrRoleNotFound, childRoleID)
	}

	delete(g.parents[childRoleID], parentRoleID)
	delete(g.children[parentRoleID], childRoleID)

	// Update slice in role struct
	newParents := make([]RoleID, 0, len(childRole.Parents))
	for _, p := range childRole.Parents {
		if p != parentRoleID {
			newParents = append(newParents, p)
		}
	}
	childRole.Parents = newParents

	return nil
}

// GetInheritedRoles returns all ancestor roles inherited directly or transitively by roleID.
// The returned list is topologically ordered (immediate parents first, then higher ancestors), deduplicated.
func (g *RoleGraph) GetInheritedRoles(roleID RoleID) ([]RoleID, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.roles[roleID]; !exists {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, roleID)
	}

	var result []RoleID
	visited := make(map[RoleID]bool)
	queue := []RoleID{roleID}
	visited[roleID] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for parentID := range g.parents[curr] {
			if !visited[parentID] {
				visited[parentID] = true
				result = append(result, parentID)
				queue = append(queue, parentID)
			}
		}
	}

	return result, nil
}

// GetChildRoles returns all descendant roles that inherit from roleID directly or transitively.
func (g *RoleGraph) GetChildRoles(roleID RoleID) ([]RoleID, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.roles[roleID]; !exists {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, roleID)
	}

	var result []RoleID
	visited := make(map[RoleID]bool)
	queue := []RoleID{roleID}
	visited[roleID] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for childID := range g.children[curr] {
			if !visited[childID] {
				visited[childID] = true
				result = append(result, childID)
				queue = append(queue, childID)
			}
		}
	}

	return result, nil
}

// GetEffectivePermissions resolves the transitive closure of all permissions granted to roleID,
// aggregating direct permissions plus all permissions inherited from ancestor roles.
// The output is deduplicated and sorted deterministically.
func (g *RoleGraph) GetEffectivePermissions(roleID RoleID) ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	role, exists := g.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrRoleNotFound, roleID)
	}

	permSet := make(map[string]struct{})

	// 1. Direct permissions
	for _, p := range role.Permissions {
		permSet[p] = struct{}{}
	}

	// 2. Transitive ancestors permissions (BFS)
	visited := make(map[RoleID]bool)
	queue := []RoleID{roleID}
	visited[roleID] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for parentID := range g.parents[curr] {
			if !visited[parentID] {
				visited[parentID] = true
				queue = append(queue, parentID)

				if parentRole, ok := g.roles[parentID]; ok {
					for _, p := range parentRole.Permissions {
						permSet[p] = struct{}{}
					}
				}
			}
		}
	}

	result := make([]string, 0, len(permSet))
	for p := range permSet {
		result = append(result, p)
	}
	sort.Strings(result)

	return result, nil
}

// HasPermission checks whether roleID holds targetPermission either directly or transitively.
func (g *RoleGraph) HasPermission(roleID RoleID, targetPermission string) (bool, error) {
	effective, err := g.GetEffectivePermissions(roleID)
	if err != nil {
		return false, err
	}
	for _, p := range effective {
		if p == targetPermission {
			return true, nil
		}
	}
	return false, nil
}

// hasPath performs a BFS to check if there is an inheritance path from `from` to `to`.
// (i.e. Does `from` inherit from `to` directly or transitively?)
// Assumes lock is already held.
func (g *RoleGraph) hasPath(from, to RoleID) bool {
	if from == to {
		return true
	}

	visited := make(map[RoleID]bool)
	queue := []RoleID{from}
	visited[from] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == to {
			return true
		}

		for parent := range g.parents[curr] {
			if !visited[parent] {
				visited[parent] = true
				queue = append(queue, parent)
			}
		}
	}

	return false
}

// hasPathExcludingEdge checks if a path exists from `from` to `to`, ignoring a specific edge.
// Assumes lock is already held.
func (g *RoleGraph) hasPathExcludingEdge(from, to, excludeFrom, excludeTo RoleID) bool {
	if from == to {
		return true
	}

	visited := make(map[RoleID]bool)
	queue := []RoleID{from}
	visited[from] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr == to {
			return true
		}

		for parent := range g.parents[curr] {
			if curr == excludeFrom && parent == excludeTo {
				continue
			}
			if !visited[parent] {
				visited[parent] = true
				queue = append(queue, parent)
			}
		}
	}

	return false
}

// Clone creates an isolated, deep copy of the role hierarchy graph.
// Useful for what-if simulation, SoD pre-assignment evaluation, and isolation.
func (g *RoleGraph) Clone() *RoleGraph {
	g.mu.RLock()
	defer g.mu.RUnlock()

	clone := NewRoleGraph()
	for id, r := range g.roles {
		clone.roles[id] = cloneRole(r)
	}

	for child, parents := range g.parents {
		clone.parents[child] = make(map[RoleID]struct{}, len(parents))
		for p := range parents {
			clone.parents[child][p] = struct{}{}
		}
	}

	for parent, children := range g.children {
		clone.children[parent] = make(map[RoleID]struct{}, len(children))
		for c := range children {
			clone.children[parent][c] = struct{}{}
		}
	}

	return clone
}

// cloneRole creates a deep copy of a Role instance.
func cloneRole(r *Role) *Role {
	if r == nil {
		return nil
	}
	cp := *r
	if r.Permissions != nil {
		cp.Permissions = make([]string, len(r.Permissions))
		copy(cp.Permissions, r.Permissions)
	}
	if r.Parents != nil {
		cp.Parents = make([]RoleID, len(r.Parents))
		copy(cp.Parents, r.Parents)
	}
	if r.Metadata != nil {
		cp.Metadata = make(map[string]string, len(r.Metadata))
		for k, v := range r.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}
