package rbac

import (
	"context"
	"fmt"
	"time"
)

// RoleRepository abstracts persistent storage operations for roles (e.g. PostgreSQL).
type RoleRepository interface {
	GetRole(ctx context.Context, id string) (*Role, error)
	GetRolesWhereIDs(ctx context.Context, ids []string) ([]*Role, error)
	CreateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, roleID string) (bool, error)
}

// PermissionCatalog abstracts external permission catalog operations, such as expanding implied permissions.
type PermissionCatalog interface {
	ResolveImplied(permissionIDs []string) []string
}

// IRBACService is the unified facade interface providing comprehensive RBAC capabilities,
// integrating PAP, PDP, PEP, SoD, Session Management, JIT, and Symmetric Auditing.
type IRBACService interface {
	// --- Policy Administration Point (PAP) ---
	CreateRole(ctx context.Context, role *Role) (*Role, error)
	GetRole(ctx context.Context, id RoleID) (*Role, error)
	UpdateRole(ctx context.Context, role *Role) (*Role, error)
	DeleteRole(ctx context.Context, id RoleID) error
	AddRoleInheritance(ctx context.Context, childRoleID, parentRoleID RoleID) error
	RemoveRoleInheritance(ctx context.Context, childRoleID, parentRoleID RoleID) error
	GetAllRoles(ctx context.Context) ([]*Role, error)

	// --- Separation of Duties (Constrained RBAC) ---
	AddSoDConstraint(ctx context.Context, constraint SoDConstraint) error
	RemoveSoDConstraint(ctx context.Context, id string) error
	GetSoDConstraint(ctx context.Context, id string) (SoDConstraint, error)
	ListSoDConstraints(ctx context.Context) ([]SoDConstraint, error)

	// --- User Assignments (Provisioning) ---
	AssignRoleToUser(ctx context.Context, ua UserAssignment) error
	RevokeRoleFromUser(ctx context.Context, userID string, roleID RoleID, scope string) error
	GetUserAssignments(ctx context.Context, userID string) ([]UserAssignment, error)

	// --- Sessions & JIT Activation ---
	CreateSession(ctx context.Context, userID string, initialRoles []RoleID, ttl time.Duration) (*Session, error)
	ActivateRoleInSession(ctx context.Context, sessionID string, role RoleID, ttl time.Duration) error
	DeactivateRoleInSession(ctx context.Context, sessionID string, role RoleID) error
	TerminateSession(ctx context.Context, sessionID string) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// --- Policy Decision Point (PDP) & Enforcement (PEP) ---
	Evaluate(ctx context.Context, req AccessRequest) Decision
	Authorize(ctx context.Context, req AccessRequest) error
	AuthorizePermission(ctx context.Context, subjectID string, roleIDs []RoleID, permission string) error
	AuthorizeAny(ctx context.Context, subjectID string, roleIDs []RoleID, permissions ...string) error
	AuthorizeAll(ctx context.Context, subjectID string, roleIDs []RoleID, permissions ...string) error

	// --- Symmetric RBAC & Governance ---
	GetEffectivePermissions(ctx context.Context, roleID RoleID) ([]string, error)
	GetUserEffectivePermissions(ctx context.Context, userID string) ([]string, error)
	GetRolesWithPermission(ctx context.Context, permission string) ([]RoleID, error)
	GetUsersWithPermission(ctx context.Context, permission string) ([]string, error)
	FindOrphanedRoles(ctx context.Context) ([]RoleID, error)
	GenerateComplianceReport(ctx context.Context) (*ComplianceReport, error)

	// --- Context Manager Configuration ---
	GetContextManager() *ContextManager
}

// service implements IRBACService.
type service struct {
	graph      *RoleGraph
	sodMgr     *SoDManager
	pdp        *PDPEngine
	enforcer   *Enforcer
	contextMgr *ContextManager
	sessionMgr *SessionManager
	auditor    *SymmetricAuditor
	repo       RoleRepository
	catalog    PermissionCatalog
}

// NewService constructs a fully-wired, production-ready RBAC service facade.
func NewService(repo RoleRepository, catalog PermissionCatalog) IRBACService {
	graph := NewRoleGraph()
	sodMgr := NewSoDManager(graph)

	var impliedResolver ImpliedPermissionResolver
	if catalog != nil {
		impliedResolver = catalog
	}

	pdp := NewPDPEngine(graph, impliedResolver)
	contextMgr := NewContextManager()
	pdp.SetContextManager(contextMgr)

	enforcer := NewEnforcer(pdp)
	sessionMgr := NewSessionManager(graph, sodMgr)
	auditor := NewSymmetricAuditor(graph, sodMgr)

	return &service{
		graph:      graph,
		sodMgr:     sodMgr,
		pdp:        pdp,
		enforcer:   enforcer,
		contextMgr: contextMgr,
		sessionMgr: sessionMgr,
		auditor:    auditor,
		repo:       repo,
		catalog:    catalog,
	}
}

// --- Policy Administration Point (PAP) ---

func (s *service) CreateRole(ctx context.Context, role *Role) (*Role, error) {
	if role == nil {
		return nil, fmt.Errorf("%w: role cannot be nil", ErrInvalidInput)
	}

	// 1. Add to in-memory graph (validates uniqueness and parents cycle check)
	if err := s.graph.AddRole(role); err != nil {
		return nil, err
	}

	// 2. Persist to repository if configured
	if s.repo != nil {
		persisted, err := s.repo.CreateRole(ctx, role)
		if err != nil {
			// Rollback from graph
			_ = s.graph.DeleteRole(role.ID)
			return nil, err
		}
		return persisted, nil
	}

	return role, nil
}

func (s *service) GetRole(ctx context.Context, id RoleID) (*Role, error) {
	// First check memory graph
	role, err := s.graph.GetRole(id)
	if err == nil {
		return role, nil
	}

	// Fallback to repository
	if s.repo != nil {
		persisted, repoErr := s.repo.GetRole(ctx, id)
		if repoErr == nil && persisted != nil {
			_ = s.graph.AddRole(persisted)
			return persisted, nil
		}
	}

	return nil, err
}

func (s *service) UpdateRole(ctx context.Context, role *Role) (*Role, error) {
	if err := s.graph.UpdateRole(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *service) DeleteRole(ctx context.Context, id RoleID) error {
	// 1. Delete from graph (checks immutable protection)
	if err := s.graph.DeleteRole(id); err != nil {
		return err
	}

	// 2. Delete from repository if configured
	if s.repo != nil {
		_, err := s.repo.DeleteRole(ctx, id)
		return err
	}

	return nil
}

func (s *service) AddRoleInheritance(ctx context.Context, childRoleID, parentRoleID RoleID) error {
	return s.graph.AddInheritance(childRoleID, parentRoleID)
}

func (s *service) RemoveRoleInheritance(ctx context.Context, childRoleID, parentRoleID RoleID) error {
	return s.graph.RemoveInheritance(childRoleID, parentRoleID)
}

func (s *service) GetAllRoles(ctx context.Context) ([]*Role, error) {
	return s.graph.GetAllRoles(), nil
}

// --- Separation of Duties (Constrained RBAC) ---

func (s *service) AddSoDConstraint(ctx context.Context, constraint SoDConstraint) error {
	return s.sodMgr.AddConstraint(constraint)
}

func (s *service) RemoveSoDConstraint(ctx context.Context, id string) error {
	return s.sodMgr.RemoveConstraint(id)
}

func (s *service) GetSoDConstraint(ctx context.Context, id string) (SoDConstraint, error) {
	return s.sodMgr.GetConstraint(id)
}

func (s *service) ListSoDConstraints(ctx context.Context) ([]SoDConstraint, error) {
	return s.sodMgr.ListConstraints(), nil
}

// --- User Assignments (Provisioning) ---

func (s *service) AssignRoleToUser(ctx context.Context, ua UserAssignment) error {
	return s.auditor.AssignRole(ua)
}

func (s *service) RevokeRoleFromUser(ctx context.Context, userID string, roleID RoleID, scope string) error {
	return s.auditor.RevokeRole(userID, roleID, scope)
}

func (s *service) GetUserAssignments(ctx context.Context, userID string) ([]UserAssignment, error) {
	return s.auditor.GetUserAssignments(userID), nil
}

// --- Sessions & JIT Activation ---

func (s *service) CreateSession(ctx context.Context, userID string, initialRoles []RoleID, ttl time.Duration) (*Session, error) {
	return s.sessionMgr.CreateSession(userID, initialRoles, ttl)
}

func (s *service) ActivateRoleInSession(ctx context.Context, sessionID string, role RoleID, ttl time.Duration) error {
	return s.sessionMgr.ActivateRole(sessionID, role, ttl)
}

func (s *service) DeactivateRoleInSession(ctx context.Context, sessionID string, role RoleID) error {
	return s.sessionMgr.DeactivateRole(sessionID, role)
}

func (s *service) TerminateSession(ctx context.Context, sessionID string) error {
	return s.sessionMgr.TerminateSession(sessionID)
}

func (s *service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	return s.sessionMgr.GetSession(sessionID)
}

// --- Policy Decision Point (PDP) & Enforcement (PEP) ---

func (s *service) Evaluate(ctx context.Context, req AccessRequest) Decision {
	return s.pdp.Evaluate(ctx, req)
}

func (s *service) Authorize(ctx context.Context, req AccessRequest) error {
	return s.enforcer.Authorize(ctx, req)
}

func (s *service) AuthorizePermission(ctx context.Context, subjectID string, roleIDs []RoleID, permission string) error {
	return s.enforcer.AuthorizePermission(ctx, subjectID, roleIDs, permission)
}

func (s *service) AuthorizeAny(ctx context.Context, subjectID string, roleIDs []RoleID, permissions ...string) error {
	return s.enforcer.AuthorizeAny(ctx, subjectID, roleIDs, permissions...)
}

func (s *service) AuthorizeAll(ctx context.Context, subjectID string, roleIDs []RoleID, permissions ...string) error {
	return s.enforcer.AuthorizeAll(ctx, subjectID, roleIDs, permissions...)
}

// --- Symmetric RBAC & Governance ---

func (s *service) GetEffectivePermissions(ctx context.Context, roleID RoleID) ([]string, error) {
	return s.graph.GetEffectivePermissions(roleID)
}

func (s *service) GetUserEffectivePermissions(ctx context.Context, userID string) ([]string, error) {
	return s.auditor.GetUserEffectivePermissions(userID)
}

func (s *service) GetRolesWithPermission(ctx context.Context, permission string) ([]RoleID, error) {
	return s.auditor.GetRolesWithPermission(permission)
}

func (s *service) GetUsersWithPermission(ctx context.Context, permission string) ([]string, error) {
	return s.auditor.GetUsersWithPermission(permission)
}

func (s *service) FindOrphanedRoles(ctx context.Context) ([]RoleID, error) {
	return s.auditor.FindOrphanedRoles(), nil
}

func (s *service) GenerateComplianceReport(ctx context.Context) (*ComplianceReport, error) {
	sessions := make([]Session, 0)
	return s.auditor.GenerateComplianceReport(sessions), nil
}

func (s *service) GetContextManager() *ContextManager {
	return s.contextMgr
}
