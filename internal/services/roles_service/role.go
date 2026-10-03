package roles

import (
	"sort"

	perms "youtrack/internal/services/permissions_services"
)

// Predefined Role Identifiers as defined in YouTrack official documentation.
const (
	RoleSystemAdmin    = "SYSTEM_ADMIN"
	RoleProjectAdmin   = "PROJECT_ADMIN"
	RoleContributor    = "CONTRIBUTOR"
	RoleObserver       = "OBSERVER"
	RoleUserManager    = "USER_MANAGER"
	RoleProjectCreator = "PROJECT_CREATOR"
	RoleDeveloper      = "DEVELOPER" // Legacy role in upgraded installations
)

// Role models a YouTrack role, which acts as a container for permissions.
type Role struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Permissions []string         `json:"permissions"`
	Scope       perms.ScopeLevel `json:"scope"`
	IsReadOnly  bool             `json:"is_read_only"`
}

// HasPermission checks if the role contains the given permission ID.
func (r *Role) HasPermission(permissionID string) bool {
	for _, p := range r.Permissions {
		if p == permissionID {
			return true
		}
	}
	return false
}

// HasProjectScopePermission checks if the role contains at least one permission
// that has Project scope in the YouTrack permission catalog.
// This is required by YouTrack's Project Scope Guard.
func (r *Role) HasProjectScopePermission(permSvc perms.IPermissionService) bool {
	for _, permID := range r.Permissions {
		if p, ok := permSvc.GetPermission(permID); ok {
			if p.Scope == perms.ScopeProject {
				return true
			}
		}
	}
	return false
}

// BuildDefaultRoles returns the 6 predefined built-in YouTrack roles
// (plus the legacy Developer role) as defined by YouTrack Cloud 2026.2.
func BuildDefaultRoles(permSvc perms.IPermissionService) map[string]Role {
	allPerms := permSvc.GetAllPermissions()
	allPermIDs := make([]string, len(allPerms))
	for i, p := range allPerms {
		allPermIDs[i] = p.ID
	}
	sort.Strings(allPermIDs)

	// 1. System Admin: full access across all scopes
	systemAdmin := Role{
		ID:          RoleSystemAdmin,
		Name:        "System Admin",
		Description: "Full administrative access to the entire YouTrack installation and all projects and organizations.",
		Permissions: allPermIDs,
		Scope:       perms.ScopeGlobal,
		IsReadOnly:  true,
	}

	// 2. Project Admin: project administration, issues, articles, comments, work items
	projectAdminPerms := []string{
		// Project
		perms.PermReadProjectBasic,
		perms.PermReadProjectFull,
		perms.PermUpdateProject,
		// Issue
		perms.PermReadIssue,
		perms.PermReadIssuePrivateFields,
		perms.PermUpdateIssue,
		perms.PermCreateIssue,
		perms.PermDeleteIssue,
		perms.PermLinkIssue,
		perms.PermUpdateIssuePrivateFields,
		perms.PermApplyCommandsSilently,
		perms.PermViewWatchers,
		perms.PermUpdateWatchers,
		perms.PermViewVoters,
		// Attachment
		perms.PermAddAttachment,
		perms.PermUpdateAttachment,
		perms.PermDeleteAttachment,
		// Comment
		perms.PermCreateIssueComment,
		perms.PermReadIssueComment,
		perms.PermUpdateIssueComment,
		perms.PermDeleteIssueComment,
		perms.PermUpdateNotOwnIssueComment,
		perms.PermDeleteNotOwnAndPermanentCommentDelete,
		perms.PermReadArticleComment,
		perms.PermCreateArticleComment,
		perms.PermUpdateArticleComment,
		perms.PermDeleteArticleComment,
		// Work Item
		perms.PermReadWorkItem,
		perms.PermUpdateWorkItem,
		perms.PermUpdateNotOwnWorkItem,
		perms.PermCreateWorkItem,
		perms.PermCreateNotOwnWorkItem,
		// Article
		perms.PermReadArticle,
		perms.PermCreateArticle,
		perms.PermUpdateArticle,
		perms.PermDeleteArticle,
		// App
		perms.PermReadAppContent,
		perms.PermUpdateAppContent,
		// Watch Folder
		perms.PermCreateWatchFolder,
		perms.PermUpdateWatchFolder,
		perms.PermDeleteWatchFolder,
		perms.PermShareWatchFolder,
	}
	resolvedProjectAdmin := permSvc.ResolveImplied(projectAdminPerms)

	projectAdmin := Role{
		ID:          RoleProjectAdmin,
		Name:        "Project Admin",
		Description: "Administrative access to project settings, issues, comments, work items, and articles.",
		Permissions: resolvedProjectAdmin,
		Scope:       perms.ScopeProject,
		IsReadOnly:  true,
	}

	// 3. Contributor: daily work on issues, articles, comments, and work items
	contributorPerms := []string{
		// Project
		perms.PermReadProjectBasic,
		// Issue
		perms.PermReadIssue,
		perms.PermReadIssuePrivateFields,
		perms.PermUpdateIssue,
		perms.PermCreateIssue,
		perms.PermDeleteIssue,
		perms.PermLinkIssue,
		perms.PermUpdateIssuePrivateFields,
		perms.PermViewWatchers,
		perms.PermUpdateWatchers,
		perms.PermViewVoters,
		// Attachment
		perms.PermAddAttachment,
		perms.PermUpdateAttachment,
		perms.PermDeleteAttachment,
		// Comment
		perms.PermCreateIssueComment,
		perms.PermReadIssueComment,
		perms.PermUpdateIssueComment,
		perms.PermDeleteIssueComment,
		perms.PermReadArticleComment,
		perms.PermCreateArticleComment,
		// Work Item
		perms.PermReadWorkItem,
		perms.PermUpdateWorkItem,
		perms.PermCreateWorkItem,
		// Article
		perms.PermReadArticle,
		perms.PermCreateArticle,
		// Watch Folder
		perms.PermCreateWatchFolder,
		perms.PermUpdateWatchFolder,
		perms.PermDeleteWatchFolder,
	}
	resolvedContributor := permSvc.ResolveImplied(contributorPerms)

	contributor := Role{
		ID:          RoleContributor,
		Name:        "Contributor",
		Description: "Standard role for project team members to create and edit issues, comments, work items, and articles.",
		Permissions: resolvedContributor,
		Scope:       perms.ScopeProject,
		IsReadOnly:  true,
	}

	// 4. Observer: basic user profile access
	observerPerms := []string{
		perms.PermUpdateSelf,
		perms.PermReadUserBasic,
		perms.PermReadUserDetails,
	}
	resolvedObserver := permSvc.ResolveImplied(observerPerms)

	observer := Role{
		ID:          RoleObserver,
		Name:        "Observer",
		Description: "Basic access to view user profiles and update own profile.",
		Permissions: resolvedObserver,
		Scope:       perms.ScopeGlobal,
		IsReadOnly:  true,
	}

	// 5. User Manager: creating users globally
	userManagerPerms := []string{
		perms.PermCreateUser,
	}
	resolvedUserManager := permSvc.ResolveImplied(userManagerPerms)

	userManager := Role{
		ID:          RoleUserManager,
		Name:        "User Manager",
		Description: "Ability to create new user accounts in YouTrack.",
		Permissions: resolvedUserManager,
		Scope:       perms.ScopeGlobal,
		IsReadOnly:  true,
	}

	// 6. Project Creator: creating projects globally
	projectCreatorPerms := []string{
		perms.PermCreateProject,
	}
	resolvedProjectCreator := permSvc.ResolveImplied(projectCreatorPerms)

	projectCreator := Role{
		ID:          RoleProjectCreator,
		Name:        "Project Creator",
		Description: "Ability to create new projects in YouTrack.",
		Permissions: resolvedProjectCreator,
		Scope:       perms.ScopeGlobal,
		IsReadOnly:  true,
	}

	// 7. Legacy Developer role (similar to Contributor for backward compatibility)
	developer := Role{
		ID:          RoleDeveloper,
		Name:        "Developer",
		Description: "Legacy role for project team members in upgraded installations.",
		Permissions: resolvedContributor,
		Scope:       perms.ScopeProject,
		IsReadOnly:  true,
	}

	return map[string]Role{
		RoleSystemAdmin:    systemAdmin,
		RoleProjectAdmin:   projectAdmin,
		RoleContributor:    contributor,
		RoleObserver:       observer,
		RoleUserManager:    userManager,
		RoleProjectCreator: projectCreator,
		RoleDeveloper:      developer,
	}
}
