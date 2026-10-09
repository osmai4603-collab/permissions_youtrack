package roles

// import (
// 	"fmt"
// 	"sort"
// 	perms "youtrack/internal/services/permissions_services"
// )

// // Service is the concrete implementation of the generic Role and Authorization Management engine.
// type Service struct {
// 	permSvc       perms.IPermissionService
// 	idProvider    PrincipalHierarchyProvider
// 	scopeProvider ScopeHierarchyProvider
// }

// // NewService creates and initializes a new YouTrack roles service.
// // It is completely decoupled from any external entities (Group, Project, Organization, Issue)
// // and relies on PrincipalHierarchyProvider and ScopeHierarchyProvider SPIs.
// // If providers are nil, a default InMemoryHierarchyProvider is used.
// func NewService(permSvc perms.IPermissionService, idProvider PrincipalHierarchyProvider, scopeProvider ScopeHierarchyProvider) *Service {
// 	if permSvc == nil {
// 		permSvc = perms.NewService()
// 	}

// 	inMem := NewInMemoryHierarchyProvider()
// 	if idProvider == nil {
// 		idProvider = inMem
// 	}
// 	if scopeProvider == nil {
// 		scopeProvider = inMem
// 	}

// 	svc := &Service{
// 		permSvc:       permSvc,
// 		idProvider:    idProvider,
// 		scopeProvider: scopeProvider,
// 	}

// 	// Baseline assignment: "All Users" virtual group gets RoleObserver globally
// 	svc.assignmentSeq++
// 	baseAssignmentID := fmt.Sprintf("assign-%d", svc.assignmentSeq)
// 	svc.assignments[baseAssignmentID] = RoleAssignment{
// 		ID:        baseAssignmentID,
// 		RoleID:    RoleObserver,
// 		ScopeRef:  GlobalScope(),
// 		AssignRef: UserAssign(AllUsersGroupID),
// 	}

// 	return svc
// }

// // hasAdminWrite checks if the actor possesses the administrative permission
// // required to perform role and assignment management operations.
// func (svc *Service) hasAdminWrite(actorPerms []string) bool {
// 	return svc.permSvc.HasPermission(actorPerms, perms.PermLowLevelAdminWrite)
// }

// // checkEscalation validates that the actor is not attempting to grant or configure
// // permissions that they do not personally hold (privilege escalation protection).
// func (svc *Service) checkEscalation(actorPerms []string, targetPerms []string) error {
// 	actorPermSet := make(map[string]struct{}, len(actorPerms))
// 	for _, p := range actorPerms {
// 		actorPermSet[p] = struct{}{}
// 	}

// 	for _, p := range targetPerms {
// 		if _, ok := actorPermSet[p]; !ok {
// 			return fmt.Errorf("%w: missing permission %s", ErrPrivilegeEscalation, p)
// 		}
// 	}
// 	return nil
// }

// // GetRole retrieves a role by ID.
// func (svc *Service) GetRole(id string) (Role, error) {
// 	svc.mu.RLock()
// 	defer svc.mu.RUnlock()

// 	r, ok := svc.roles[id]
// 	if !ok {
// 		return Role{}, ErrRoleNotFound
// 	}
// 	return r, nil
// }

// // GetAllRoles retrieves all registered roles sorted by ID.
// func (svc *Service) GetAllRoles() []Role {
// 	svc.mu.RLock()
// 	defer svc.mu.RUnlock()

// 	rolesList := make([]Role, 0, len(svc.roles))
// 	for _, r := range svc.roles {
// 		rolesList = append(rolesList, r)
// 	}

// 	sort.Slice(rolesList, func(i, j int) bool {
// 		return rolesList[i].ID < rolesList[j].ID
// 	})

// 	return rolesList
// }

// // CreateRole adds a new custom role to the system.
// // Built-in roles are read-only and predefined.
// // The actor must possess ADMIN_UPDATE_APP and cannot grant permissions they do not possess.
// func (svc *Service) CreateRole(actorPerms []string, role Role) (*Role, error) {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return nil, ErrPermissionDenied
// 	}

// 	if err := svc.checkEscalation(actorPerms, role.Permissions); err != nil {
// 		return nil, err
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	if _, exists := svc.roles[role.ID]; exists {
// 		return nil, ErrRoleAlreadyExists
// 	}

// 	// Resolve implied permissions automatically
// 	resolvedPerms := svc.permSvc.ResolveImplied(role.Permissions)

// 	newRole := Role{
// 		ID:          role.ID,
// 		Name:        role.Name,
// 		Description: role.Description,
// 		Permissions: resolvedPerms,
// 		IsImmutable: false, // Custom roles are always editable
// 	}

// 	svc.roles[role.ID] = newRole
// 	return &newRole, nil
// }

// // UpdateRole modifies the details and permissions of a custom role.
// // Built-in roles cannot be modified.
// // When permissions are removed, dependent permissions are cascade-revoked.
// func (svc *Service) UpdateRole(actorPerms []string, id string, name string, description string, newPerms []string) (*Role, error) {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return nil, ErrPermissionDenied
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	existing, ok := svc.roles[id]
// 	if !ok {
// 		return nil, ErrRoleNotFound
// 	}
// 	if existing.IsImmutable {
// 		return nil, ErrReadOnlyRole
// 	}

// 	// Validate privilege escalation for new permissions
// 	if err := svc.checkEscalation(actorPerms, newPerms); err != nil {
// 		return nil, err
// 	}

// 	// Resolve implied permissions for all requested permissions
// 	resolvedPerms := svc.permSvc.ResolveImplied(newPerms)

// 	// For any permission explicitly removed, ensure dependent permissions are also revoked
// 	newSet := make(map[string]struct{}, len(resolvedPerms))
// 	for _, p := range resolvedPerms {
// 		newSet[p] = struct{}{}
// 	}

// 	activePerms := resolvedPerms
// 	for _, oldPerm := range existing.Permissions {
// 		if _, stillPresent := newSet[oldPerm]; !stillPresent {
// 			activePerms = svc.permSvc.ResolveRevocation(activePerms, oldPerm)
// 		}
// 	}

// 	existing.Name = name
// 	existing.Description = description
// 	existing.Permissions = activePerms
// 	svc.roles[id] = existing

// 	return &existing, nil
// }

// // DeleteRole removes a custom role and cleans up any assignments pointing to it.
// // Built-in roles cannot be deleted.
// func (svc *Service) DeleteRole(actorPerms []string, id string) error {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return ErrPermissionDenied
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	r, ok := svc.roles[id]
// 	if !ok {
// 		return ErrRoleNotFound
// 	}
// 	if r.IsImmutable {
// 		return ErrReadOnlyRole
// 	}

// 	// Remove all assignments for this role
// 	for aid, assignment := range svc.assignments {
// 		if assignment.RoleID == id {
// 			delete(svc.assignments, aid)
// 		}
// 	}

// 	delete(svc.roles, id)
// 	return nil
// }

// // CloneRole creates an independent, fully editable copy of an existing role.
// // This is the standard YouTrack mechanism for customizing built-in roles.
// func (svc *Service) CloneRole(actorPerms []string, sourceRoleID string, newRoleID string, newName string, newDescription string) (*Role, error) {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return nil, ErrPermissionDenied
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	source, ok := svc.roles[sourceRoleID]
// 	if !ok {
// 		return nil, ErrRoleNotFound
// 	}

// 	if _, exists := svc.roles[newRoleID]; exists {
// 		return nil, ErrRoleAlreadyExists
// 	}

// 	// Validate privilege escalation: caller cannot clone a role containing permissions they don't hold
// 	if err := svc.checkEscalation(actorPerms, source.Permissions); err != nil {
// 		return nil, err
// 	}

// 	clonedPerms := make([]string, len(source.Permissions))
// 	copy(clonedPerms, source.Permissions)

// 	clonedRole := Role{
// 		ID:          newRoleID,
// 		Name:        newName,
// 		Description: newDescription,
// 		Permissions: clonedPerms,
// 		IsImmutable: false, // Cloned roles are always editable
// 	}

// 	svc.roles[newRoleID] = clonedRole
// 	return &clonedRole, nil
// }

// // MergeRoles merges one or more source roles into a target role according to YouTrack rules:
// // - All assignments of source roles are migrated to the target role.
// // - The target role's permissions are unchanged.
// // - Source roles are permanently deleted.
// // - Built-in roles cannot be merged/deleted as sources.
// func (svc *Service) MergeRoles(actorPerms []string, targetRoleID string, sourceRoleIDs []string) error {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return ErrPermissionDenied
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	if _, targetExists := svc.roles[targetRoleID]; !targetExists {
// 		return fmt.Errorf("%w: target role %s", ErrRoleNotFound, targetRoleID)
// 	}

// 	// Verify all source roles exist and are not built-in
// 	for _, srcID := range sourceRoleIDs {
// 		src, exists := svc.roles[srcID]
// 		if !exists {
// 			return fmt.Errorf("%w: source role %s", ErrRoleNotFound, srcID)
// 		}
// 		if src.IsImmutable {
// 			return fmt.Errorf("%w: role %s", ErrCannotMergeBuiltInRole, srcID)
// 		}
// 	}

// 	// Re-assign all assignments referencing source roles to targetRoleID
// 	sourceMap := make(map[string]struct{}, len(sourceRoleIDs))
// 	for _, srcID := range sourceRoleIDs {
// 		sourceMap[srcID] = struct{}{}
// 	}

// 	for aid, assignment := range svc.assignments {
// 		if _, isSource := sourceMap[assignment.RoleID]; isSource {
// 			assignment.RoleID = targetRoleID
// 			svc.assignments[aid] = assignment
// 		}
// 	}

// 	// Delete source roles
// 	for _, srcID := range sourceRoleIDs {
// 		delete(svc.roles, srcID)
// 	}

// 	return nil
// }

// // AssignRole grants a role to a principal (User or Group) in a target scope.
// // Enforces YouTrack's Project Scope Guard:
// // Roles assigned at Project scope must have at least one project-scoped permission.
// func (svc *Service) AssignRole(actorPerms []string, roleID string, assign AssignRef, scope ScopeRef) (*RoleAssignment, error) {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return nil, ErrPermissionDenied
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	role, ok := svc.roles[roleID]
// 	if !ok {
// 		return nil, ErrRoleNotFound
// 	}

// 	// Validate target scope using the ScopeHierarchyProvider SPI
// 	if err := svc.scopeProvider.ValidateScope(scope); err != nil {
// 		return nil, err
// 	}

// 	// Enforce Project Scope Guard
// 	if scope.Scope == perms.ScopeProject {
// 		if !role.HasProjectScopePermission(svc.permSvc) {
// 			return nil, ErrNoProjectScopePermission
// 		}
// 	}

// 	svc.assignmentSeq++
// 	aid := fmt.Sprintf("assign-%d", svc.assignmentSeq)
// 	assignment := RoleAssignment{
// 		ID:        aid,
// 		RoleID:    roleID,
// 		AssignRef: assign,
// 		ScopeRef:  scope,
// 	}

// 	svc.assignments[aid] = assignment
// 	return &assignment, nil
// }

// // RevokeAssignment removes a role assignment by its assignment ID.
// func (svc *Service) RevokeAssignment(actorPerms []string, assignmentID string) error {
// 	if !svc.hasAdminWrite(actorPerms) {
// 		return ErrPermissionDenied
// 	}

// 	svc.mu.Lock()
// 	defer svc.mu.Unlock()

// 	if _, ok := svc.assignments[assignmentID]; !ok {
// 		return ErrAssignmentNotFound
// 	}

// 	delete(svc.assignments, assignmentID)
// 	return nil
// }

// // GetAssignmentsForPrincipal returns all role assignments granted to a principal.
// func (svc *Service) GetAssignmentsForPrincipal(assign AssignRef) []RoleAssignment {
// 	svc.mu.RLock()
// 	defer svc.mu.RUnlock()

// 	var out []RoleAssignment
// 	for _, a := range svc.assignments {
// 		if a.AssignType == assign.AssignType && a.AssignerID == assign.AssignerID {
// 			out = append(out, a)
// 		}
// 	}
// 	return out
// }

// // GetAssignmentsForScope returns all role assignments configured for a scope.
// func (svc *Service) GetAssignmentsForScope(scope ScopeRef) []RoleAssignment {
// 	svc.mu.RLock()
// 	defer svc.mu.RUnlock()

// 	var out []RoleAssignment
// 	for _, a := range svc.assignments {
// 		if a.Scope == scope.Scope && a.TargetID == scope.TargetID {
// 			out = append(out, a)
// 		}
// 	}
// 	return out
// }
