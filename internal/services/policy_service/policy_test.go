package policy

import (
	"testing"

	perms "youtrack/internal/services/permissions_services"
	roles "youtrack/internal/services/roles_service"
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

func setupTestContext(_ *testing.T) (*Service, roles.IRolesService, *roles.InMemoryHierarchyProvider, perms.IPermissionService, []string) {
	permSvc := perms.NewService()
	idProvider := roles.NewInMemoryHierarchyProvider()
	rolesSvc := roles.NewService(permSvc, idProvider, idProvider)
	policySvc := NewService(rolesSvc, permSvc, idProvider)
	adminPerms := getAdminActorPerms(permSvc)
	return policySvc, rolesSvc, idProvider, permSvc, adminPerms
}

func TestContextualVisibility_Issue(t *testing.T) {
	policySvc, rolesSvc, idProvider, _, adminPerms := setupTestContext(t)

	projVisScope := roles.ScopedProject("PROJ_VIS")
	idProvider.RegisterScope(projVisScope)

	// User Alice has Contributor on PROJ_VIS
	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleContributor, roles.UserAssign("alice"), projVisScope)
	// User Bob has Contributor on PROJ_VIS
	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleContributor, roles.UserAssign("bob"), projVisScope)
	// User Admin has System Admin globally
	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleSystemAdmin, roles.UserAssign("admin-user"), roles.GlobalScope())

	issueUnrestricted := IssueContext{
		ID:         "ISSUE-1",
		ProjectID:  "PROJ_VIS",
		ReporterID: "alice",
	}

	issueRestricted := IssueContext{
		ID:         "ISSUE-2",
		ProjectID:  "PROJ_VIS",
		ReporterID: "alice",
		VisibleTo:  []string{"alice"}, // Restricted to Alice
	}

	// 1. Charlie has no READ_ISSUE -> cannot view unrestricted
	canCharlie, _ := policySvc.CanViewIssue("charlie", issueUnrestricted)
	if canCharlie {
		t.Errorf("Charlie has no project permission, should not view issue")
	}

	// 2. Bob has READ_ISSUE, can view unrestricted
	canBobUnrestricted, _ := policySvc.CanViewIssue("bob", issueUnrestricted)
	if !canBobUnrestricted {
		t.Errorf("Bob should be able to view unrestricted issue")
	}

	// 3. Bob is not in VisibleTo for restricted issue -> cannot view
	canBobRestricted, _ := policySvc.CanViewIssue("bob", issueRestricted)
	if canBobRestricted {
		t.Errorf("Bob should NOT be able to view issue restricted to Alice")
	}

	// 4. Alice is the reporter and in VisibleTo -> can view
	canAlice, _ := policySvc.CanViewIssue("alice", issueRestricted)
	if !canAlice {
		t.Errorf("Alice is reporter and in VisibleTo, should be able to view")
	}

	// 5. Admin has READ_HIDDEN_STUFF -> can override restriction
	canAdmin, _ := policySvc.CanViewIssue("admin-user", issueRestricted)
	if !canAdmin {
		t.Errorf("Admin with READ_HIDDEN_STUFF should override visibility restriction")
	}
}

func TestContextualVisibility_ArticleInheritance(t *testing.T) {
	policySvc, rolesSvc, idProvider, _, adminPerms := setupTestContext(t)

	projKbScope := roles.ScopedProject("PROJ_KB")
	idProvider.RegisterScope(projKbScope)

	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleContributor, roles.UserAssign("alice"), projKbScope)
	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleContributor, roles.UserAssign("bob"), projKbScope)

	// Parent article is restricted to Alice
	parentArticle := ArticleContext{
		ID:        "ARTICLE-PARENT",
		ProjectID: "PROJ_KB",
		AuthorID:  "alice",
		VisibleTo: []string{"alice"},
	}

	// Child article has no restrictions of its own, but inherits from parent!
	childArticle := ArticleContext{
		ID:            "ARTICLE-CHILD",
		ProjectID:     "PROJ_KB",
		AuthorID:      "alice",
		VisibleTo:     nil,
		ParentArticle: &parentArticle,
	}

	// Bob should NOT be able to view child article because parent is restricted to Alice
	canBob, err := policySvc.CanViewArticle("bob", childArticle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if canBob {
		t.Errorf("Bob should NOT be able to view child article due to parent visibility inheritance")
	}

	// Alice can view both
	canAlice, err := policySvc.CanViewArticle("alice", childArticle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !canAlice {
		t.Errorf("Alice should be able to view child article")
	}
}

func TestPrivateCustomFieldsAccess(t *testing.T) {
	policySvc, rolesSvc, idProvider, _, adminPerms := setupTestContext(t)

	projFieldsScope := roles.ScopedProject("PROJ_FIELDS")
	idProvider.RegisterScope(projFieldsScope)

	// Contributor has READ_ISSUE and PRIVATE_READ_ISSUE, and UPDATE_ISSUE and PRIVATE_UPDATE_ISSUE
	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleContributor, roles.UserAssign("dev-user"), projFieldsScope)

	canRead, err := policySvc.CanAccessIssuePrivateField("dev-user", "PROJ_FIELDS", false)
	if err != nil || !canRead {
		t.Errorf("dev-user should have private fields read access")
	}

	canUpdate, err := policySvc.CanAccessIssuePrivateField("dev-user", "PROJ_FIELDS", true)
	if err != nil || !canUpdate {
		t.Errorf("dev-user should have private fields update access")
	}

	// Create user with only READ_ISSUE (no PRIVATE_READ_ISSUE)
	customPublicReader, _ := rolesSvc.CreateRole(adminPerms, roles.Role{
		ID:          "ROLE_PUBLIC_ONLY",
		Name:        "Public Only",
		Permissions: []string{perms.PermReadIssue},
		Scope:       perms.ScopeProject,
	})
	_, _ = rolesSvc.AssignRole(adminPerms, customPublicReader.ID, roles.UserAssign("guest-user"), projFieldsScope)

	canGuestRead, _ := policySvc.CanAccessIssuePrivateField("guest-user", "PROJ_FIELDS", false)
	if canGuestRead {
		t.Errorf("guest-user without PRIVATE_READ_ISSUE should not access private fields")
	}
}

func TestInherentAccessIntegration(t *testing.T) {
	policySvc, rolesSvc, idProvider, _, adminPerms := setupTestContext(t)

	projInherentScope := roles.ScopedProject("PROJ_INHERENT")
	idProvider.RegisterScope(projInherentScope)

	// User has CREATE_ISSUE
	_, _ = rolesSvc.AssignRole(adminPerms, roles.RoleContributor, roles.UserAssign("reporter-user"), projInherentScope)

	// Reporter reading own issue public fields -> Allowed
	canReadOwn, err := policySvc.CheckInherentAccess(perms.InherentReadOwnIssuePublicFields, "reporter-user", "reporter-user", "PROJ_INHERENT")
	if err != nil || !canReadOwn {
		t.Errorf("reporter should have inherent read access to own issue public fields")
	}

	// Non-reporter -> Denied inherent access
	canOtherRead, err := policySvc.CheckInherentAccess(perms.InherentReadOwnIssuePublicFields, "other-user", "reporter-user", "PROJ_INHERENT")
	if err != nil || canOtherRead {
		t.Errorf("non-reporter should not have inherent access")
	}

	// Deleting own attachment: any user can delete attachment they uploaded
	canDeleteOwnAttach, err := policySvc.CheckInherentAccess(perms.InherentDeleteOwnAttachment, "uploader-user", "uploader-user", "PROJ_INHERENT")
	if err != nil || !canDeleteOwnAttach {
		t.Errorf("any user can delete own attachment via inherent access")
	}
}
