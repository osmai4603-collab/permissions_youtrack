package perms

// ScopeLevel defines the level at which a permission applies in YouTrack.
type ScopeLevel string

const (
	// ScopeGlobal applies system-wide across all organizations and projects.
	ScopeGlobal ScopeLevel = "GLOBAL"
	// ScopeOrganization applies within a specific organization hierarchy.
	ScopeOrganization ScopeLevel = "ORGANIZATION"
	// ScopeProject applies within a specific project.
	ScopeProject ScopeLevel = "PROJECT"
)

// ModuleType represents the functional domain in YouTrack.
type ModuleType string

const (
	ModuleSystem          ModuleType = "SYSTEM"
	ModuleApp             ModuleType = "APP"
	ModuleArticle         ModuleType = "ARTICLE"
	ModuleArticleComment  ModuleType = "ARTICLE_COMMENT"
	ModuleIssue           ModuleType = "ISSUE"
	ModuleIssueAttachment ModuleType = "ISSUE_ATTACHMENT"
	ModuleIssueComment    ModuleType = "ISSUE_COMMENT"
	ModuleIssueWorkItem   ModuleType = "ISSUE_WORK_ITEM"
	ModuleOrganization    ModuleType = "ORGANIZATION"
	ModuleProject         ModuleType = "PROJECT"
	ModuleUser            ModuleType = "USER"
	ModuleWatchFolder     ModuleType = "WATCH_FOLDER"
)

// EntityType represents the target resource being governed.
type EntityType string

const (
	EntitySystem         EntityType = "SYSTEM"
	EntityApp            EntityType = "APP"
	EntityArticle        EntityType = "ARTICLE"
	EntityArticleComment EntityType = "ARTICLE_COMMENT"
	EntityIssue          EntityType = "ISSUE"
	EntityAttachment     EntityType = "ATTACHMENT"
	EntityComment        EntityType = "COMMENT"
	EntityWorkItem       EntityType = "WORK_ITEM"
	EntityOrganization   EntityType = "ORGANIZATION"
	EntityProject        EntityType = "PROJECT"
	EntityUser           EntityType = "USER"
	EntityWatchFolder    EntityType = "WATCH_FOLDER"
)

// OperationType defines the CRUD or operational action.
type OperationType string

const (
	OpCreate  OperationType = "CREATE"
	OpRead    OperationType = "READ"
	OpUpdate  OperationType = "UPDATE"
	OpDelete  OperationType = "DELETE"
	OpLink    OperationType = "LINK"
	OpShare   OperationType = "SHARE"
	OpAdmin   OperationType = "ADMIN"
	OpSpecial OperationType = "SPECIAL"
)

// Permission models a single authorization rule in YouTrack.
type Permission struct {
	ID             string        `json:"id"`
	DisplayName    string        `json:"display_name"`
	Description    string        `json:"description"`
	IsGlobal       bool          `json:"is_global"`
	Module         ModuleType    `json:"module"`
	Entity         EntityType    `json:"entity"`
	Scope          ScopeLevel    `json:"scope"`
	Operation      OperationType `json:"operation"`
	ImpliedPerms   []string      `json:"implied_perms,omitempty"`
	DependentPerms []string      `json:"dependent_perms,omitempty"`
}

// InherentAction represents actions that can be performed via inherent rights.
type InherentAction string

const (
	InherentReadOwnIssuePublicFields   InherentAction = "READ_OWN_ISSUE_PUBLIC_FIELDS"
	InherentUpdateOwnIssuePublicFields InherentAction = "UPDATE_OWN_ISSUE_PUBLIC_FIELDS"
	InherentLinkOwnIssue               InherentAction = "LINK_OWN_ISSUE"
	InherentModifyOwnAttachment        InherentAction = "MODIFY_OWN_ATTACHMENT"
	InherentDeleteOwnAttachment        InherentAction = "DELETE_OWN_ATTACHMENT"
	InherentRestrictOwnAttachment      InherentAction = "RESTRICT_OWN_ATTACHMENT"
	InherentReadOwnIssueComment        InherentAction = "READ_OWN_ISSUE_COMMENT"
	InherentReadOwnWorkItem            InherentAction = "READ_OWN_WORK_ITEM"
	InherentReadOwnArticleComment      InherentAction = "READ_OWN_ARTICLE_COMMENT"
	InherentUpdateOwnArticleComment    InherentAction = "UPDATE_OWN_ARTICLE_COMMENT"
	InherentDeleteOwnArticleComment    InherentAction = "DELETE_OWN_ARTICLE_COMMENT"
)
