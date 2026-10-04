package perms

import "sort"

// Standard YouTrack Permission Keys as defined in YouTrack Server Documentation and DevPortal
const (
	// System (Global)
	PermLowLevelAdminRead  = "ADMIN_READ_APP"
	PermLowLevelAdminWrite = "ADMIN_UPDATE_APP"

	// App (Project)
	PermReadAppContent   = "READ_APP_CONTENT"
	PermUpdateAppContent = "UPDATE_APP_CONTENT"

	// Article (Project)
	PermCreateArticle = "CREATE_ARTICLE"
	PermDeleteArticle = "DELETE_ARTICLE"
	PermReadArticle   = "READ_ARTICLE"
	PermUpdateArticle = "UPDATE_ARTICLE"

	// Article Comment (Project)
	PermCreateArticleComment = "CREATE_ARTICLE_COMMENT"
	PermDeleteArticleComment = "DELETE_ARTICLE_COMMENT"
	PermReadArticleComment   = "READ_ARTICLE_COMMENT"
	PermUpdateArticleComment = "UPDATE_ARTICLE_COMMENT"

	// Issue (Project)
	PermApplyCommandsSilently    = "APPLY_COMMANDS_SILENTLY"
	PermCreateIssue              = "CREATE_ISSUE"
	PermDeleteIssue              = "DELETE_ISSUE"
	PermLinkIssue                = "LINK_ISSUE"
	PermOverrideVisibility       = "READ_HIDDEN_STUFF"
	PermReadIssue                = "READ_ISSUE"
	PermReadIssuePrivateFields   = "PRIVATE_READ_ISSUE"
	PermUpdateIssue              = "UPDATE_ISSUE"
	PermUpdateIssuePrivateFields = "PRIVATE_UPDATE_ISSUE"
	PermUpdateWatchers           = "UPDATE_WATCHERS"
	PermViewVoters               = "VIEW_VOTERS"
	PermViewWatchers             = "VIEW_WATCHERS"

	// Issue Attachment (Project)
	PermAddAttachment    = "CREATE_ATTACHMENT_ISSUE"
	PermDeleteAttachment = "DELETE_ATTACHMENT_ISSUE"
	PermUpdateAttachment = "UPDATE_ATTACHMENT_ISSUE"

	// Issue Comment (Project)
	PermCreateIssueComment                    = "CREATE_COMMENT"
	PermDeleteIssueComment                    = "DELETE_COMMENT"
	PermDeleteNotOwnAndPermanentCommentDelete = "DELETE_NOT_OWN_COMMENT"
	PermReadIssueComment                      = "READ_COMMENT"
	PermUpdateIssueComment                    = "UPDATE_COMMENT"
	PermUpdateNotOwnIssueComment              = "UPDATE_NOT_OWN_COMMENT"

	// Issue Work Item (Project)
	PermCreateNotOwnWorkItem = "CREATE_NOT_OWN_WORK_ITEM"
	PermCreateWorkItem       = "CREATE_WORK_ITEM"
	PermReadWorkItem         = "READ_WORK_ITEM"
	PermUpdateNotOwnWorkItem = "UPDATE_NOT_OWN_WORK_ITEM"
	PermUpdateWorkItem       = "UPDATE_WORK_ITEM"

	// Organization (Global / Organization)
	PermCreateOrganization = "CREATE_ORGANIZATION"
	PermDeleteOrganization = "DELETE_ORGANIZATION"
	PermReadOrganization   = "READ_ORGANIZATION"
	PermUpdateOrganization = "UPDATE_ORGANIZATION"

	// Project (Global / Project)
	PermCreateProject    = "CREATE_PROJECT"
	PermDeleteProject    = "DELETE_PROJECT"
	PermReadProjectBasic = "READ_PROJECT_BASIC"
	PermReadProjectFull  = "READ_PROJECT"
	PermUpdateProject    = "UPDATE_PROJECT"

	// Users (Global)
	PermCreateUser      = "CREATE_USER"
	PermDeleteUser      = "DELETE_USER"
	PermReadUserBasic   = "READ_USER_BASIC"
	PermReadUserDetails = "READ_USER"
	PermUpdateSelf      = "UPDATE_PROFILE"
	PermUpdateUser      = "UPDATE_USER"

	// Watch Folder / Saved Searches & Tags
	PermCreateWatchFolder = "CREATE_WATCH_FOLDER"
	PermDeleteWatchFolder = "DELETE_WATCH_FOLDER"
	PermUpdateWatchFolder = "UPDATE_WATCH_FOLDER"
	PermShareWatchFolder  = "SHARE_WATCH_FOLDER"
)

// rawPermissionDefinitions contains the baseline YouTrack permissions.
var rawPermissionDefinitions = []Permission{
	// System (Global Scope)
	{
		ID:          PermLowLevelAdminRead,
		DisplayName: "Low-level Admin Read",
		Description: "Read-only access to low-level administrative settings. View integrations, metrics, groups, and roles.",
		IsGlobal:    true,
		Entity:      EntitySystem,
		Scope:       ScopeGlobal,
		Operation:   OpRead,
	},
	{
		ID:           PermLowLevelAdminWrite,
		DisplayName:  "Low-level Admin Write",
		Description:  "Manage low-level administrative actions, integrations, database backups, groups, and roles.",
		IsGlobal:     true,
		Entity:       EntitySystem,
		Scope:        ScopeGlobal,
		Operation:    OpAdmin,
		ImpliedPerms: []string{PermLowLevelAdminRead},
	},

	// App (Project Scope)
	{
		ID:          PermReadAppContent,
		DisplayName: "Read App Content",
		Description: "View app, workflow, and SLA policy scripts, configuration files, logs, and export archives.",
		IsGlobal:    false,
		Entity:      EntityApp,
		Scope:       ScopeProject,
		Operation:   OpRead,
	},
	{
		ID:           PermUpdateAppContent,
		DisplayName:  "Update App Content",
		Description:  "Create, import, update, and delete app, workflow, and SLA policy scripts and configuration files.",
		IsGlobal:     false,
		Entity:       EntityApp,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadAppContent},
	},

	// Article (Project Scope)
	{
		ID:           PermCreateArticle,
		DisplayName:  "Create Article",
		Description:  "Add articles to the knowledge base for a specific project.",
		IsGlobal:     false,
		Entity:       EntityArticle,
		Scope:        ScopeProject,
		Operation:    OpCreate,
		ImpliedPerms: []string{PermReadArticle},
	},
	{
		ID:           PermDeleteArticle,
		DisplayName:  "Delete Article",
		Description:  "Delete articles from the knowledge base in a specific project.",
		IsGlobal:     false,
		Entity:       EntityArticle,
		Scope:        ScopeProject,
		Operation:    OpDelete,
		ImpliedPerms: []string{PermReadArticle},
	},
	{
		ID:          PermReadArticle,
		DisplayName: "Read Article",
		Description: "View articles and article content in the knowledge base for a specific project.",
		IsGlobal:    false,
		Entity:      EntityArticle,
		Scope:       ScopeProject,
		Operation:   OpRead,
	},
	{
		ID:           PermUpdateArticle,
		DisplayName:  "Update Article",
		Description:  "Edit existing articles in the knowledge base for a specific project.",
		IsGlobal:     false,
		Entity:       EntityArticle,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadArticle},
	},

	// Article Comment (Project Scope)
	{
		ID:           PermCreateArticleComment,
		DisplayName:  "Create Article Comment",
		Description:  "Add comments to existing articles in the knowledge base for a specific project.",
		IsGlobal:     false,
		Entity:       EntityArticleComment,
		Scope:        ScopeProject,
		Operation:    OpCreate,
		ImpliedPerms: []string{PermReadArticleComment},
	},
	{
		ID:           PermDeleteArticleComment,
		DisplayName:  "Delete Article Comment",
		Description:  "Delete comments that have been posted to articles in the knowledge base for a specific project.",
		IsGlobal:     false,
		Entity:       EntityArticleComment,
		Scope:        ScopeProject,
		Operation:    OpDelete,
		ImpliedPerms: []string{PermReadArticleComment},
	},
	{
		ID:          PermReadArticleComment,
		DisplayName: "Read Article Comment",
		Description: "View comments that have been posted to articles in the knowledge base for a specific project.",
		IsGlobal:    false,
		Entity:      EntityArticleComment,
		Scope:       ScopeProject,
		Operation:   OpRead,
	},
	{
		ID:           PermUpdateArticleComment,
		DisplayName:  "Update Article Comment",
		Description:  "Edit comments that have been posted to articles in the knowledge base for a specific project.",
		IsGlobal:     false,
		Entity:       EntityArticleComment,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadArticleComment},
	},

	// Issue (Project Scope)
	{
		ID:          PermApplyCommandsSilently,
		DisplayName: "Apply Commands Silently",
		Description: "Update issue attributes using a command without sending update notification messages to subscribers.",
		IsGlobal:    false,
		Entity:      EntityIssue,
		Scope:       ScopeProject,
		Operation:   OpSpecial,
	},
	{
		ID:           PermCreateIssue,
		DisplayName:  "Create Issue",
		Description:  "Create (report) issues in a project. Reporters inherit viewing/updating public fields and linking their issues.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpCreate,
		ImpliedPerms: []string{PermReadProjectBasic},
	},
	{
		ID:          PermDeleteIssue,
		DisplayName: "Delete Issue",
		Description: "Delete issues.",
		IsGlobal:    false,
		Entity:      EntityIssue,
		Scope:       ScopeProject,
		Operation:   OpDelete,
	},
	{
		ID:          PermLinkIssue,
		DisplayName: "Link Issues",
		Description: "Add links that define relationships between issues.",
		IsGlobal:    false,
		Entity:      EntityIssue,
		Scope:       ScopeProject,
		Operation:   OpLink,
	},
	{
		ID:           PermOverrideVisibility,
		DisplayName:  "Override Visibility Restrictions",
		Description:  "View issues, comments, and attachments that are hidden by visibility settings.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpSpecial,
		ImpliedPerms: []string{PermReadIssuePrivateFields},
	},
	{
		ID:           PermReadIssue,
		DisplayName:  "Read Issue",
		Description:  "View issues and read public fields.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpRead,
		ImpliedPerms: []string{PermReadProjectBasic},
	},
	{
		ID:           PermReadIssuePrivateFields,
		DisplayName:  "Read Issue Private Fields",
		Description:  "View private fields in issues.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpRead,
		ImpliedPerms: []string{PermReadProjectBasic},
	},
	{
		ID:          PermUpdateIssue,
		DisplayName: "Update Issue",
		Description: "Update the values for public fields in issues.",
		IsGlobal:    false,
		Entity:      EntityIssue,
		Scope:       ScopeProject,
		Operation:   OpUpdate,
	},
	{
		ID:           PermUpdateIssuePrivateFields,
		DisplayName:  "Update Issue Private Fields",
		Description:  "Update the values for private fields in issues.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadIssuePrivateFields, PermUpdateIssue},
	},
	{
		ID:          PermUpdateWatchers,
		DisplayName: "Update Watchers",
		Description: "Add other users to the list of watchers for an issue.",
		IsGlobal:    false,
		Entity:      EntityIssue,
		Scope:       ScopeProject,
		Operation:   OpUpdate,
	},
	{
		ID:           PermViewVoters,
		DisplayName:  "View Voters",
		Description:  "View the list of users who have voted for an issue.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpRead,
		ImpliedPerms: []string{PermReadProjectBasic},
	},
	{
		ID:           PermViewWatchers,
		DisplayName:  "View Watchers",
		Description:  "View the list of users who are watching an issue.",
		IsGlobal:     false,
		Entity:       EntityIssue,
		Scope:        ScopeProject,
		Operation:    OpRead,
		ImpliedPerms: []string{PermReadProjectBasic},
	},

	// Issue Attachment (Project Scope)
	{
		ID:          PermAddAttachment,
		DisplayName: "Add Attachment",
		Description: "Attach files to issues. Uploaders inherit modifying and restricting their files without explicit update permission.",
		IsGlobal:    false,
		Entity:      EntityAttachment,
		Scope:       ScopeProject,
		Operation:   OpCreate,
	},
	{
		ID:          PermDeleteAttachment,
		DisplayName: "Delete Attachment",
		Description: "Delete any file attached to an issue. All users can delete files they attached themselves.",
		IsGlobal:    false,
		Entity:      EntityAttachment,
		Scope:       ScopeProject,
		Operation:   OpDelete,
	},
	{
		ID:          PermUpdateAttachment,
		DisplayName: "Update Attachment",
		Description: "Modify files attached to issues and restrict attachment visibility.",
		IsGlobal:    false,
		Entity:      EntityAttachment,
		Scope:       ScopeProject,
		Operation:   OpUpdate,
	},

	// Issue Comment (Project Scope)
	{
		ID:          PermCreateIssueComment,
		DisplayName: "Create Issue Comment",
		Description: "Add comments to issues. Users inherit permission to read their own comments.",
		IsGlobal:    false,
		Entity:      EntityComment,
		Scope:       ScopeProject,
		Operation:   OpCreate,
	},
	{
		ID:          PermDeleteIssueComment,
		DisplayName: "Delete Issue Comment",
		Description: "Delete comments that have been added to issues.",
		IsGlobal:    false,
		Entity:      EntityComment,
		Scope:       ScopeProject,
		Operation:   OpDelete,
	},
	{
		ID:           PermDeleteNotOwnAndPermanentCommentDelete,
		DisplayName:  "Delete Not Own and Permanent Comment Delete",
		Description:  "Delete comments that were added to issues by other users and delete comments permanently.",
		IsGlobal:     false,
		Entity:       EntityComment,
		Scope:        ScopeProject,
		Operation:    OpDelete,
		ImpliedPerms: []string{PermReadIssueComment},
	},
	{
		ID:          PermReadIssueComment,
		DisplayName: "Read Issue Comment",
		Description: "View comments that have been added to issues.",
		IsGlobal:    false,
		Entity:      EntityComment,
		Scope:       ScopeProject,
		Operation:   OpRead,
	},
	{
		ID:          PermUpdateIssueComment,
		DisplayName: "Update Issue Comment",
		Description: "Edit comments that have been added to issues.",
		IsGlobal:    false,
		Entity:      EntityComment,
		Scope:       ScopeProject,
		Operation:   OpUpdate,
	},
	{
		ID:           PermUpdateNotOwnIssueComment,
		DisplayName:  "Update Not Own Issue Comment",
		Description:  "Edit comments that were added to issues by other users.",
		IsGlobal:     false,
		Entity:       EntityComment,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadIssueComment},
	},

	// Issue Work Item (Project Scope)
	{
		ID:           PermCreateNotOwnWorkItem,
		DisplayName:  "Create Not Own Work Item",
		Description:  "Create work items and set another user as the work author.",
		IsGlobal:     false,
		Entity:       EntityWorkItem,
		Scope:        ScopeProject,
		Operation:    OpCreate,
		ImpliedPerms: []string{PermCreateWorkItem},
	},
	{
		ID:          PermCreateWorkItem,
		DisplayName: "Create Work Item",
		Description: "Add work items to issues. Users inherit permission to read their own work items.",
		IsGlobal:    false,
		Entity:      EntityWorkItem,
		Scope:       ScopeProject,
		Operation:   OpCreate,
	},
	{
		ID:          PermReadWorkItem,
		DisplayName: "Read Work Item",
		Description: "View the list of work items in an issue.",
		IsGlobal:    false,
		Entity:      EntityWorkItem,
		Scope:       ScopeProject,
		Operation:   OpRead,
	},
	{
		ID:           PermUpdateNotOwnWorkItem,
		DisplayName:  "Update Not Own Work Item",
		Description:  "Edit work items created by other users. Also grants permission to create work items on behalf of other users.",
		IsGlobal:     false,
		Entity:       EntityWorkItem,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadWorkItem, PermUpdateWorkItem},
	},
	{
		ID:          PermUpdateWorkItem,
		DisplayName: "Update Work Item",
		Description: "Edit work items that they have added to issues.",
		IsGlobal:    false,
		Entity:      EntityWorkItem,
		Scope:       ScopeProject,
		Operation:   OpUpdate,
	},

	// Organization (Global / Organization Scope)
	{
		ID:           PermCreateOrganization,
		DisplayName:  "Create Organization",
		Description:  "Add new organizations to the system.",
		IsGlobal:     true,
		Entity:       EntityOrganization,
		Scope:        ScopeGlobal,
		Operation:    OpCreate,
		ImpliedPerms: []string{PermReadOrganization},
	},
	{
		ID:           PermDeleteOrganization,
		DisplayName:  "Delete Organization",
		Description:  "Permanently remove organization records from the system.",
		IsGlobal:     false,
		Entity:       EntityOrganization,
		Scope:        ScopeOrganization,
		Operation:    OpDelete,
		ImpliedPerms: []string{PermReadOrganization},
	},
	{
		ID:          PermReadOrganization,
		DisplayName: "Read Organization",
		Description: "View organizations and their attributes, roles, and project assignments.",
		IsGlobal:    false,
		Entity:      EntityOrganization,
		Scope:       ScopeOrganization,
		Operation:   OpRead,
	},
	{
		ID:           PermUpdateOrganization,
		DisplayName:  "Update Organization",
		Description:  "Edit organization attributes, manage project assignments and access rights.",
		IsGlobal:     false,
		Entity:       EntityOrganization,
		Scope:        ScopeOrganization,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadOrganization},
	},

	// Project (Global / Project Scope)
	{
		ID:          PermCreateProject,
		DisplayName: "Create Project",
		Description: "Create new projects across the system.",
		IsGlobal:    true,
		Entity:      EntityProject,
		Scope:       ScopeGlobal,
		Operation:   OpCreate,
	},
	{
		ID:           PermDeleteProject,
		DisplayName:  "Delete Project",
		Description:  "Delete projects.",
		IsGlobal:     false,
		Entity:       EntityProject,
		Scope:        ScopeProject,
		Operation:    OpDelete,
		ImpliedPerms: []string{PermReadProjectFull},
	},
	{
		ID:          PermReadProjectBasic,
		DisplayName: "Read Project Basic",
		Description: "View basic project properties including name, description, logo, and project owner.",
		IsGlobal:    false,
		Entity:      EntityProject,
		Scope:       ScopeProject,
		Operation:   OpRead,
	},
	{
		ID:           PermReadProjectFull,
		DisplayName:  "Read Project Full",
		Description:  "View all project properties, roles, and permissions granted to the project team and users.",
		IsGlobal:     false,
		Entity:       EntityProject,
		Scope:        ScopeProject,
		Operation:    OpRead,
		ImpliedPerms: []string{PermReadProjectBasic},
	},
	{
		ID:           PermUpdateProject,
		DisplayName:  "Update Project",
		Description:  "Edit project properties, groups, roles, assignees, custom fields, workflows, apps, integrations.",
		IsGlobal:     false,
		Entity:       EntityProject,
		Scope:        ScopeProject,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermReadProjectFull},
	},

	// Users (Global Scope)
	{
		ID:          PermCreateUser,
		DisplayName: "Create User",
		Description: "Create new user accounts and invite users to register their own accounts.",
		IsGlobal:    true,
		Entity:      EntityUser,
		Scope:       ScopeGlobal,
		Operation:   OpCreate,
	},
	{
		ID:           PermDeleteUser,
		DisplayName:  "Delete User",
		Description:  "Delete user accounts.",
		IsGlobal:     true,
		Entity:       EntityUser,
		Scope:        ScopeGlobal,
		Operation:    OpDelete,
		ImpliedPerms: []string{PermReadUserDetails},
	},
	{
		ID:          PermReadUserBasic,
		DisplayName: "Read User Basic",
		Description: "View list of registered users and read ID, username, name, and avatar.",
		IsGlobal:    true,
		Entity:      EntityUser,
		Scope:       ScopeGlobal,
		Operation:   OpRead,
	},
	{
		ID:           PermReadUserDetails,
		DisplayName:  "Read User Details",
		Description:  "View additional profile details for all registered users.",
		IsGlobal:     true,
		Entity:       EntityUser,
		Scope:        ScopeGlobal,
		Operation:    OpRead,
		ImpliedPerms: []string{PermReadUserBasic},
	},
	{
		ID:          PermUpdateSelf,
		DisplayName: "Update Self",
		Description: "Edit own Account Security information in the user profile (2FA, credentials, tokens).",
		IsGlobal:    true,
		Entity:      EntityUser,
		Scope:       ScopeGlobal,
		Operation:   OpUpdate,
	},
	{
		ID:           PermUpdateUser,
		DisplayName:  "Update User",
		Description:  "Edit user profile data, ban, merge, and anonymize user accounts.",
		IsGlobal:     true,
		Entity:       EntityUser,
		Scope:        ScopeGlobal,
		Operation:    OpUpdate,
		ImpliedPerms: []string{PermUpdateSelf, PermReadUserDetails},
	},

	// Watch Folder / Saved Searches & Tags
	{
		ID:          PermCreateWatchFolder,
		DisplayName: "Create Tag or Saved Search",
		Description: "Create personal or shared tags, saved searches, and custom views.",
		IsGlobal:    false,
		Entity:      EntityWatchFolder,
		Scope:       ScopeProject,
		Operation:   OpCreate,
	},
	{
		ID:          PermDeleteWatchFolder,
		DisplayName: "Delete Tag or Saved Search",
		Description: "Delete tags, saved searches, and custom views.",
		IsGlobal:    false,
		Entity:      EntityWatchFolder,
		Scope:       ScopeProject,
		Operation:   OpDelete,
	},
	{
		ID:          PermUpdateWatchFolder,
		DisplayName: "Edit Tag or Saved Search",
		Description: "Edit tags, saved searches, and custom view configurations.",
		IsGlobal:    false,
		Entity:      EntityWatchFolder,
		Scope:       ScopeProject,
		Operation:   OpUpdate,
	},
	{
		ID:          PermShareWatchFolder,
		DisplayName: "Share Custom View",
		Description: "Share tags, saved searches, and custom views with other users and groups.",
		IsGlobal:    false,
		Entity:      EntityWatchFolder,
		Scope:       ScopeProject,
		Operation:   OpShare,
	},
}

// BuildDefaultCatalog generates the full permission registry and computes
// dependent permissions automatically by inverting implied relationships.
func BuildDefaultCatalog() map[string]Permission {
	catalog := make(map[string]Permission, len(rawPermissionDefinitions))
	for _, p := range rawPermissionDefinitions {
		catalog[p.ID] = p
	}

	// Compute dependent permissions
	// If A implies B, then B has dependent A
	dependentsMap := make(map[string]map[string]struct{})
	for _, p := range rawPermissionDefinitions {
		for _, impliedID := range p.ImpliedPerms {
			if dependentsMap[impliedID] == nil {
				dependentsMap[impliedID] = make(map[string]struct{})
			}
			dependentsMap[impliedID][p.ID] = struct{}{}
		}
	}

	for id, deps := range dependentsMap {
		if p, ok := catalog[id]; ok {
			depList := make([]string, 0, len(deps))
			for depID := range deps {
				depList = append(depList, depID)
			}
			sort.Strings(depList)
			p.DependentPerms = depList
			catalog[id] = p
		}
	}

	return catalog
}
