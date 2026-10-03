package roles

import (
	"testing"

	perms "youtrack/internal/services/permissions_services"
)

// Helper to provide a system admin actor permissions set
func getAdminActorPerms(permSvc perms.IPermissionService) []string {
	allPerms := permSvc.GetAllPermissions()
	ids := make([]string, len(allPerms))
	for i, p := range allPerms {
		ids[i] = p.ID
	}
	return ids
}

func setupTestRolesService(_ *testing.T) (*Service, *InMemoryHierarchyProvider, perms.IPermissionService, []string) {
	permSvc := perms.NewService()
	idProvider := NewInMemoryHierarchyProvider()
	svc := NewService(permSvc, idProvider, idProvider)
	adminPerms := getAdminActorPerms(permSvc)
	return svc, idProvider, permSvc, adminPerms
}

func TestDefaultRolesInitialization(t *testing.T) {
	svc, _, permSvc, _ := setupTestRolesService(t)

	allRoles := svc.GetAllRoles()
	if len(allRoles) < 6 {
		t.Fatalf("expected at least 6 default roles, got %d", len(allRoles))
	}

	// 1. Check System Admin
	sysAdmin, err := svc.GetRole(RoleSystemAdmin)
	if err != nil {
		t.Fatalf("failed to get System Admin role: %v", err)
	}
	if !sysAdmin.IsReadOnly {
		t.Errorf("expected System Admin to be read-only")
	}
	if len(sysAdmin.Permissions) != len(permSvc.GetAllPermissions()) {
		t.Errorf("expected System Admin to have all %d permissions, got %d", len(permSvc.GetAllPermissions()), len(sysAdmin.Permissions))
	}

	// 2. Check Project Admin
	projAdmin, err := svc.GetRole(RoleProjectAdmin)
	if err != nil {
		t.Fatalf("failed to get Project Admin role: %v", err)
	}
	if !projAdmin.IsReadOnly {
		t.Errorf("expected Project Admin to be read-only")
	}
	if !projAdmin.HasPermission(perms.PermUpdateProject) {
		t.Errorf("Project Admin should have UPDATE_PROJECT")
	}
	if !projAdmin.HasPermission(perms.PermUpdateNotOwnIssueComment) {
		t.Errorf("Project Admin should have UPDATE_NOT_OWN_COMMENT")
	}
	if !projAdmin.HasPermission(perms.PermDeleteArticle) {
		t.Errorf("Project Admin should have DELETE_ARTICLE")
	}

	// 3. Check Contributor
	contributor, err := svc.GetRole(RoleContributor)
	if err != nil {
		t.Fatalf("failed to get Contributor role: %v", err)
	}
	if !contributor.IsReadOnly {
		t.Errorf("expected Contributor to be read-only")
	}
	if !contributor.HasPermission(perms.PermCreateIssue) {
		t.Errorf("Contributor should have CREATE_ISSUE")
	}
	if contributor.HasPermission(perms.PermUpdateProject) {
		t.Errorf("Contributor should NOT have UPDATE_PROJECT")
	}
	if contributor.HasPermission(perms.PermUpdateNotOwnIssueComment) {
		t.Errorf("Contributor should NOT have UPDATE_NOT_OWN_COMMENT")
	}
	if contributor.HasPermission(perms.PermDeleteArticle) {
		t.Errorf("Contributor should NOT have DELETE_ARTICLE")
	}

	// 4. Check Observer
	observer, err := svc.GetRole(RoleObserver)
	if err != nil {
		t.Fatalf("failed to get Observer role: %v", err)
	}
	if !observer.HasPermission(perms.PermReadUserDetails) {
		t.Errorf("Observer should have READ_USER")
	}
	if !observer.HasPermission(perms.PermUpdateSelf) {
		t.Errorf("Observer should have UPDATE_PROFILE")
	}

	// 5. Check User Manager
	userManager, err := svc.GetRole(RoleUserManager)
	if err != nil {
		t.Fatalf("failed to get User Manager role: %v", err)
	}
	if !userManager.HasPermission(perms.PermCreateUser) {
		t.Errorf("User Manager should have CREATE_USER")
	}

	// 6. Check Project Creator
	projCreator, err := svc.GetRole(RoleProjectCreator)
	if err != nil {
		t.Fatalf("failed to get Project Creator role: %v", err)
	}
	if !projCreator.HasPermission(perms.PermCreateProject) {
		t.Errorf("Project Creator should have CREATE_PROJECT")
	}
}

func TestRoleLifecycle_Create_Update_Delete(t *testing.T) {
	svc, _, _, adminPerms := setupTestRolesService(t)

	// Unauthorized creation
	_, err := svc.CreateRole([]string{perms.PermReadIssue}, Role{
		ID:          "CUSTOM_TEST",
		Name:        "Custom Test",
		Permissions: []string{perms.PermReadIssue},
	})
	if err != ErrPermissionDenied {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}

	// Escalation check: actor lacks CREATE_USER
	restrictedAdminPerms := []string{perms.PermLowLevelAdminWrite, perms.PermReadIssue}
	_, err = svc.CreateRole(restrictedAdminPerms, Role{
		ID:          "CUSTOM_ESCALATE",
		Name:        "Escalate",
		Permissions: []string{perms.PermCreateUser},
	})
	if err == nil {
		t.Fatalf("expected privilege escalation error, got nil")
	}

	// Valid creation: CREATE_ISSUE automatically implies READ_PROJECT_BASIC
	customRole, err := svc.CreateRole(adminPerms, Role{
		ID:          "CUSTOM_ISSUE_CREATOR",
		Name:        "Issue Creator",
		Description: "Can only create issues",
		Permissions: []string{perms.PermCreateIssue},
		Scope:       perms.ScopeProject,
	})
	if err != nil {
		t.Fatalf("unexpected error creating role: %v", err)
	}
	if customRole.IsReadOnly {
		t.Errorf("new custom role should not be read-only")
	}
	if !customRole.HasPermission(perms.PermReadProjectBasic) {
		t.Errorf("CREATE_ISSUE should have automatically implied READ_PROJECT_BASIC")
	}

	// Update built-in role should fail
	_, err = svc.UpdateRole(adminPerms, RoleContributor, "Hacked Contributor", "Hacked", []string{perms.PermCreateUser})
	if err != ErrReadOnlyRole {
		t.Fatalf("expected ErrReadOnlyRole, got %v", err)
	}

	// Valid update: Add CREATE_ARTICLE
	updatedRole, err := svc.UpdateRole(adminPerms, customRole.ID, "Lead Issue Creator", "Updated Description", []string{perms.PermCreateIssue, perms.PermCreateArticle})
	if err != nil {
		t.Fatalf("failed updating role: %v", err)
	}
	if !updatedRole.HasPermission(perms.PermCreateArticle) {
		t.Errorf("expected updated role to have CREATE_ARTICLE")
	}

	// Delete custom role
	if err := svc.DeleteRole(adminPerms, customRole.ID); err != nil {
		t.Fatalf("failed deleting custom role: %v", err)
	}
	if _, err := svc.GetRole(customRole.ID); err != ErrRoleNotFound {
		t.Fatalf("expected ErrRoleNotFound after deletion, got %v", err)
	}

	// Delete built-in role should fail
	if err := svc.DeleteRole(adminPerms, RoleContributor); err != ErrReadOnlyRole {
		t.Fatalf("expected ErrReadOnlyRole when deleting built-in role, got %v", err)
	}
}

func TestRoleLifecycle_Clone(t *testing.T) {
	svc, _, _, adminPerms := setupTestRolesService(t)

	// Clone Contributor into custom role
	cloned, err := svc.CloneRole(adminPerms, RoleContributor, "CUSTOM_CONTRIBUTOR", "Custom Contributor", "Editable Contributor Clone")
	if err != nil {
		t.Fatalf("failed to clone role: %v", err)
	}
	if cloned.IsReadOnly {
		t.Errorf("cloned role must be editable (IsReadOnly = false)")
	}
	if len(cloned.Permissions) == 0 {
		t.Errorf("cloned role permissions should not be empty")
	}

	// Edit cloned role (e.g. add Apply Commands Silently)
	updatedCloned, err := svc.UpdateRole(adminPerms, "CUSTOM_CONTRIBUTOR", "Enhanced Contributor", "Has silent command", append(cloned.Permissions, perms.PermApplyCommandsSilently))
	if err != nil {
		t.Fatalf("failed to update cloned role: %v", err)
	}
	if !updatedCloned.HasPermission(perms.PermApplyCommandsSilently) {
		t.Errorf("expected enhanced contributor to have APPLY_COMMANDS_SILENTLY")
	}

	// Verify original Contributor was untouched
	original, err := svc.GetRole(RoleContributor)
	if err != nil {
		t.Fatalf("failed to get original Contributor: %v", err)
	}
	if original.HasPermission(perms.PermApplyCommandsSilently) {
		t.Errorf("original Contributor should not be modified by clone changes")
	}
}

func TestRoleLifecycle_Merge(t *testing.T) {
	svc, idProvider, _, adminPerms := setupTestRolesService(t)

	proj1Scope := ScopedProject("PROJ_1")
	idProvider.RegisterScope(proj1Scope)

	// Create Target custom role
	targetRole, err := svc.CreateRole(adminPerms, Role{
		ID:          "TARGET_ROLE",
		Name:        "Target Role",
		Permissions: []string{perms.PermReadIssue, perms.PermCreateIssue},
		Scope:       perms.ScopeProject,
	})
	if err != nil {
		t.Fatalf("failed to create target role: %v", err)
	}

	// Create Source custom role 1
	_, err = svc.CreateRole(adminPerms, Role{
		ID:          "SRC_ROLE_1",
		Name:        "Source Role 1",
		Permissions: []string{perms.PermReadIssue},
		Scope:       perms.ScopeProject,
	})
	if err != nil {
		t.Fatalf("failed to create src role 1: %v", err)
	}

	// Assign SRC_ROLE_1 to a user
	assign, err := svc.AssignRole(adminPerms, "SRC_ROLE_1", UserAssign("user-bob"), proj1Scope)
	if err != nil {
		t.Fatalf("failed to assign src role 1: %v", err)
	}

	// Attempting to merge built-in role should fail
	err = svc.MergeRoles(adminPerms, targetRole.ID, []string{RoleContributor})
	if err == nil {
		t.Fatalf("expected error merging built-in role as source, got nil")
	}

	// Valid merge: merge SRC_ROLE_1 into TARGET_ROLE
	if err := svc.MergeRoles(adminPerms, targetRole.ID, []string{"SRC_ROLE_1"}); err != nil {
		t.Fatalf("failed to merge roles: %v", err)
	}

	// SRC_ROLE_1 should be deleted
	if _, err := svc.GetRole("SRC_ROLE_1"); err != ErrRoleNotFound {
		t.Fatalf("expected SRC_ROLE_1 to be deleted, got %v", err)
	}

	// Assignment of user-bob should now point to TARGET_ROLE
	userAssignments := svc.GetAssignmentsForPrincipal(UserAssign("user-bob"))
	if len(userAssignments) != 1 {
		t.Fatalf("expected 1 assignment for user-bob, got %d", len(userAssignments))
	}
	if userAssignments[0].RoleID != targetRole.ID {
		t.Errorf("expected assignment role ID to be %s, got %s", targetRole.ID, userAssignments[0].RoleID)
	}
	if userAssignments[0].ID != assign.ID {
		t.Errorf("assignment ID should be preserved")
	}
}

func TestProjectScopeGuard(t *testing.T) {
	svc, idProvider, _, adminPerms := setupTestRolesService(t)

	projTestScope := ScopedProject("PROJ_TEST")
	idProvider.RegisterScope(projTestScope)

	// RoleUserManager has only CREATE_USER (Global scope) -> Should be rejected by Project Scope Guard
	_, err := svc.AssignRole(adminPerms, RoleUserManager, UserAssign("user-alice"), projTestScope)
	if err != ErrNoProjectScopePermission {
		t.Fatalf("expected ErrNoProjectScopePermission, got %v", err)
	}

	// RoleObserver has only user profile permissions (Global scope) -> Should be rejected
	_, err = svc.AssignRole(adminPerms, RoleObserver, UserAssign("user-alice"), projTestScope)
	if err != ErrNoProjectScopePermission {
		t.Fatalf("expected ErrNoProjectScopePermission, got %v", err)
	}

	// RoleContributor has project-scoped permissions -> Should succeed
	_, err = svc.AssignRole(adminPerms, RoleContributor, UserAssign("user-alice"), projTestScope)
	if err != nil {
		t.Fatalf("expected assigning Contributor to project to succeed, got %v", err)
	}
}

func TestAssignmentsAndAggregation_AllUsersBaseline(t *testing.T) {
	svc, _, _, _ := setupTestRolesService(t)

	// Brand new user with no explicit assignments at all
	hasReadUserBasic, err := svc.HasPermission(UserAssign("brand-new-user"), GlobalScope(), perms.PermReadUserBasic)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasReadUserBasic {
		t.Errorf("every registered user should have READ_USER_BASIC via All Users group")
	}

	hasUpdateProfile, err := svc.HasPermission(UserAssign("brand-new-user"), GlobalScope(), perms.PermUpdateSelf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasUpdateProfile {
		t.Errorf("every registered user should have UPDATE_PROFILE via All Users group")
	}
}

func TestNestedGroupInheritance(t *testing.T) {
	svc, idProvider, _, adminPerms := setupTestRolesService(t)

	orgScope := ScopedOrganization("ORG_TECH")
	projScope := ScopedProject("PROJ_APP")
	idProvider.RegisterScope(orgScope)
	idProvider.RegisterScope(projScope, orgScope)

	// Group Hierarchy: Group "GRP_BACKEND" has Parent "GRP_ENGINEERS"
	idProvider.SetGroupParent("GRP_BACKEND", "GRP_ENGINEERS")
	idProvider.AddUserToGroup("user-dave", "GRP_BACKEND")

	// Assign Contributor to parent group GRP_ENGINEERS in PROJ_APP
	_, err := svc.AssignRole(adminPerms, RoleContributor, GroupAssign("GRP_ENGINEERS"), projScope)
	if err != nil {
		t.Fatalf("failed to assign role to engineers: %v", err)
	}

	// Dave should inherit Contributor in PROJ_APP
	hasCreateIssue, err := svc.HasPermission(UserAssign("user-dave"), projScope, perms.PermCreateIssue)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !hasCreateIssue {
		t.Errorf("user-dave should inherit CREATE_ISSUE from parent group GRP_ENGINEERS")
	}
}

func TestDocumentationScenario1_Ahmed(t *testing.T) {
	// Scenario 1: User Ahmed belongs to QA Team with Contributor role on Project Alpha.
	// Admin grants Ahmed Project Admin directly on Project Alpha.
	// Result: Ahmed effectively gets full Project Admin + Contributor (Pure additive union).
	svc, idProvider, _, adminPerms := setupTestRolesService(t)

	projAlpha := ScopedProject("PROJ_ALPHA")
	idProvider.RegisterScope(projAlpha)
	idProvider.AddUserToGroup("ahmed", "GRP_QA")

	// QA Team gets Contributor
	_, err := svc.AssignRole(adminPerms, RoleContributor, GroupAssign("GRP_QA"), projAlpha)
	if err != nil {
		t.Fatalf("failed assigning role: %v", err)
	}

	// Ahmed gets Project Admin directly
	_, err = svc.AssignRole(adminPerms, RoleProjectAdmin, UserAssign("ahmed"), projAlpha)
	if err != nil {
		t.Fatalf("failed assigning role: %v", err)
	}

	// Check administrative abilities
	hasUpdateProject, err := svc.HasPermission(UserAssign("ahmed"), projAlpha, perms.PermUpdateProject)
	if err != nil || !hasUpdateProject {
		t.Errorf("Ahmed should have UPDATE_PROJECT from Project Admin")
	}

	hasDeleteArticle, err := svc.HasPermission(UserAssign("ahmed"), projAlpha, perms.PermDeleteArticle)
	if err != nil || !hasDeleteArticle {
		t.Errorf("Ahmed should have DELETE_ARTICLE from Project Admin")
	}
}

func TestDocumentationScenario2_Sara(t *testing.T) {
	// Scenario 2: Admin creates Custom Lead role with mixed scopes:
	// - CREATE_USER (Global)
	// - UPDATE_ORGANIZATION (Org)
	// - CREATE_ISSUE (Project)
	// - UPDATE_ISSUE (Project)
	// Assigned to Sara on Mobile App project.
	// Result:
	// - Assignment is allowed because role has project permissions.
	// - CREATE_USER and UPDATE_ORGANIZATION are disregarded in project scope.
	// - Sara effectively has CREATE_ISSUE, UPDATE_ISSUE, and implied READ_PROJECT_BASIC.
	svc, idProvider, _, adminPerms := setupTestRolesService(t)

	orgCorp := ScopedOrganization("ORG_CORP")
	projMobile := ScopedProject("PROJ_MOBILE")
	idProvider.RegisterScope(orgCorp)
	idProvider.RegisterScope(projMobile, orgCorp)

	customLead, err := svc.CreateRole(adminPerms, Role{
		ID:   "CUSTOM_LEAD",
		Name: "Custom Lead",
		Permissions: []string{
			perms.PermCreateUser,
			perms.PermUpdateOrganization,
			perms.PermCreateIssue,
			perms.PermUpdateIssue,
		},
		Scope: perms.ScopeProject,
	})
	if err != nil {
		t.Fatalf("failed creating Custom Lead role: %v", err)
	}

	_, err = svc.AssignRole(adminPerms, customLead.ID, UserAssign("sara"), projMobile)
	if err != nil {
		t.Fatalf("failed assigning Custom Lead to project: %v", err)
	}

	// In PROJ_MOBILE, Sara has CREATE_ISSUE
	hasCreateIssue, err := svc.HasPermission(UserAssign("sara"), projMobile, perms.PermCreateIssue)
	if err != nil || !hasCreateIssue {
		t.Errorf("Sara should have CREATE_ISSUE in PROJ_MOBILE")
	}

	// In PROJ_MOBILE, CREATE_USER should be disregarded!
	hasCreateUser, err := svc.HasPermission(UserAssign("sara"), projMobile, perms.PermCreateUser)
	if err != nil || hasCreateUser {
		t.Errorf("CREATE_USER should be disregarded in project assignment")
	}

	// UPDATE_ORGANIZATION should also be disregarded in project assignment!
	hasUpdateOrg, err := svc.HasPermission(UserAssign("sara"), projMobile, perms.PermUpdateOrganization)
	if err != nil || hasUpdateOrg {
		t.Errorf("UPDATE_ORGANIZATION should be disregarded in project assignment")
	}
}

func TestDocumentationScenario3_Khaled(t *testing.T) {
	// Scenario 3: Organization TechCorp has Backend, Frontend, DevOps projects.
	// Khaled has Contributor at TechCorp Org level.
	// Khaled has Observer directly on DevOps project.
	// Result:
	// - Backend & Frontend: Khaled has Contributor.
	// - DevOps: Khaled has Contributor + Observer (additive; still has Contributor rights in DevOps).
	svc, idProvider, _, adminPerms := setupTestRolesService(t)

	orgTech := ScopedOrganization("ORG_TECHCORP")
	projBackend := ScopedProject("PROJ_BACKEND")
	projFrontend := ScopedProject("PROJ_FRONTEND")
	projDevOps := ScopedProject("PROJ_DEVOPS")

	idProvider.RegisterScope(orgTech)
	idProvider.RegisterScope(projBackend, orgTech)
	idProvider.RegisterScope(projFrontend, orgTech)
	idProvider.RegisterScope(projDevOps, orgTech)

	// Assign Contributor to Khaled at Organization level
	_, err := svc.AssignRole(adminPerms, RoleContributor, UserAssign("khaled"), orgTech)
	if err != nil {
		t.Fatalf("failed assigning role at org level: %v", err)
	}

	// Assign custom project observer to Khaled directly on DevOps project
	customProjObserver, err := svc.CreateRole(adminPerms, Role{
		ID:          "CUSTOM_PROJECT_OBSERVER",
		Name:        "Project Observer",
		Permissions: []string{perms.PermReadProjectBasic},
		Scope:       perms.ScopeProject,
	})
	if err != nil {
		t.Fatalf("failed creating project observer: %v", err)
	}

	_, err = svc.AssignRole(adminPerms, customProjObserver.ID, UserAssign("khaled"), projDevOps)
	if err != nil {
		t.Fatalf("failed assigning to devops: %v", err)
	}

	// In Backend: Khaled has CREATE_ISSUE via org inheritance
	hasInBackend, err := svc.HasPermission(UserAssign("khaled"), projBackend, perms.PermCreateIssue)
	if err != nil || !hasInBackend {
		t.Errorf("Khaled should have CREATE_ISSUE in Backend via org inheritance")
	}

	// In Frontend: Khaled has CREATE_ISSUE via org inheritance
	hasInFrontend, err := svc.HasPermission(UserAssign("khaled"), projFrontend, perms.PermCreateIssue)
	if err != nil || !hasInFrontend {
		t.Errorf("Khaled should have CREATE_ISSUE in Frontend via org inheritance")
	}

	// In DevOps: Khaled STILL has CREATE_ISSUE because Contributor is inherited and additive!
	hasInDevops, err := svc.HasPermission(UserAssign("khaled"), projDevOps, perms.PermCreateIssue)
	if err != nil || !hasInDevops {
		t.Errorf("Khaled should still have CREATE_ISSUE in DevOps (additive permissions)")
	}
}
