package policy

import (
	perms "youtrack/internal/services/permissions_services"
	roles "youtrack/internal/services/roles_service"
)

// IssueContext captures metadata required for evaluating issue visibility and access.
type IssueContext struct {
	ID         string   `json:"id"`
	ProjectID  string   `json:"project_id"`
	ReporterID string   `json:"reporter_id"`
	VisibleTo  []string `json:"visible_to,omitempty"` // User IDs or Group IDs with visibility grants
}

// ArticleContext captures metadata for knowledge base articles and hierarchical inheritance.
type ArticleContext struct {
	ID            string          `json:"id"`
	ProjectID     string          `json:"project_id"`
	AuthorID      string          `json:"author_id"`
	VisibleTo     []string        `json:"visible_to,omitempty"`     // Explicit visibility list
	ParentArticle *ArticleContext `json:"parent_article,omitempty"` // Parent article for inherited visibility
}

// IPolicyService defines contextual, resource-level, and inherent access policies.
type IPolicyService interface {
	CanViewIssue(userID string, issue IssueContext) (bool, error)
	CanViewArticle(userID string, article ArticleContext) (bool, error)
	CanAccessIssuePrivateField(userID string, projectID string, isUpdate bool) (bool, error)
	CheckInherentAccess(action perms.InherentAction, userID string, authorOrReporterID string, projectID string) (bool, error)
}

// Service implements IPolicyService using IRolesService and IPermissionService.
type Service struct {
	rolesSvc   roles.IRolesService
	permSvc    perms.IPermissionService
	idProvider roles.PrincipalHierarchyProvider
}

// NewService creates a new contextual policy service.
func NewService(rolesSvc roles.IRolesService, permSvc perms.IPermissionService, idProvider roles.PrincipalHierarchyProvider) *Service {
	if permSvc == nil {
		permSvc = perms.NewService()
	}
	return &Service{
		rolesSvc:   rolesSvc,
		permSvc:    permSvc,
		idProvider: idProvider,
	}
}

// CanViewIssue checks whether a user can view a given issue:
// 1. User must possess READ_ISSUE in the project scope.
// 2. If visibility is unrestricted (VisibleTo is empty), access is granted.
// 3. If user is the issue reporter, inherent access grants visibility to public fields.
// 4. If user possesses PermOverrideVisibility (READ_HIDDEN_STUFF), restrictions are bypassed.
// 5. If user or any of their groups is included in VisibleTo, access is granted.
func (svc *Service) CanViewIssue(userID string, issue IssueContext) (bool, error) {
	projectScope := roles.ScopedProject(issue.ProjectID)
	hasReadIssue, err := svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, perms.PermReadIssue)
	if err != nil {
		return false, err
	}
	if !hasReadIssue {
		return false, nil
	}

	// Unrestricted visibility
	if len(issue.VisibleTo) == 0 {
		return true, nil
	}

	// Reporter inherent right
	if userID == issue.ReporterID {
		return true, nil
	}

	// Admin override visibility
	hasOverride, err := svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, perms.PermOverrideVisibility)
	if err != nil {
		return false, err
	}
	if hasOverride {
		return true, nil
	}

	// Check if user ID or user groups match the VisibleTo list
	var identities []string
	if svc.idProvider != nil {
		var idErr error
		identities, idErr = svc.idProvider.ResolveIdentities(roles.AssignUser, userID)
		if idErr != nil {
			return false, idErr
		}
	} else {
		identities = []string{userID, roles.AllUsersGroupID}
	}

	identityMap := make(map[string]struct{}, len(identities))
	for _, id := range identities {
		identityMap[id] = struct{}{}
	}

	for _, target := range issue.VisibleTo {
		if _, ok := identityMap[target]; ok {
			return true, nil
		}
	}

	return false, nil
}

// CanViewArticle checks if a user can view an article according to YouTrack rules:
// - User must possess READ_ARTICLE in the project scope.
// - Sub-articles inherit visibility restrictions from their parent article(s).
// - User must have access to both parent article(s) and the current article.
func (svc *Service) CanViewArticle(userID string, article ArticleContext) (bool, error) {
	projectScope := roles.ScopedProject(article.ProjectID)
	userAssign := roles.UserAssign(userID)
	hasReadArticle, err := svc.rolesSvc.HasPermission(userAssign, projectScope, perms.PermReadArticle)
	if err != nil {
		return false, err
	}
	if !hasReadArticle {
		return false, nil
	}

	// Check author inherent rights
	if userID == article.AuthorID {
		return true, nil
	}

	// Check admin override visibility
	hasOverride, err := svc.rolesSvc.HasPermission(userAssign, projectScope, perms.PermOverrideVisibility)
	if err != nil {
		return false, err
	}
	if hasOverride {
		return true, nil
	}

	// 1. If there is a parent article, user must be able to view the parent article first
	if article.ParentArticle != nil {
		canViewParent, err := svc.CanViewArticle(userID, *article.ParentArticle)
		if err != nil {
			return false, err
		}
		if !canViewParent {
			return false, nil
		}
	}

	// 2. Check local visibility restrictions on this article
	if len(article.VisibleTo) == 0 {
		return true, nil
	}

	var identities []string
	if svc.idProvider != nil {
		var idErr error
		identities, idErr = svc.idProvider.ResolveIdentities(roles.AssignUser, userID)
		if idErr != nil {
			return false, idErr
		}
	} else {
		identities = []string{userID, roles.AllUsersGroupID}
	}

	identityMap := make(map[string]struct{}, len(identities))
	for _, id := range identities {
		identityMap[id] = struct{}{}
	}

	for _, target := range article.VisibleTo {
		if _, ok := identityMap[target]; ok {
			return true, nil
		}
	}

	return false, nil
}

// CanAccessIssuePrivateField checks if the user has permission to read or update private custom fields.
// - Reading private fields requires READ_ISSUE and PRIVATE_READ_ISSUE.
// - Updating private fields requires UPDATE_ISSUE and PRIVATE_UPDATE_ISSUE.
func (svc *Service) CanAccessIssuePrivateField(userID string, projectID string, isUpdate bool) (bool, error) {
	projectScope := roles.ScopedProject(projectID)
	if isUpdate {
		hasUpdate, err := svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, perms.PermUpdateIssue)
		if err != nil || !hasUpdate {
			return false, err
		}
		return svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, perms.PermUpdateIssuePrivateFields)
	}

	hasRead, err := svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, perms.PermReadIssue)
	if err != nil || !hasRead {
		return false, err
	}
	return svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, perms.PermReadIssuePrivateFields)
}

// CheckInherentAccess evaluates inherent rights granted to creators/reporters without explicit permissions,
// dynamically bound to the user's project permissions.
func (svc *Service) CheckInherentAccess(action perms.InherentAction, userID string, authorOrReporterID string, projectID string) (bool, error) {
	projectScope := roles.ScopedProject(projectID)
	isAuthor := (userID == authorOrReporterID)
	hasPermFn := func(requiredPerm string) bool {
		has, err := svc.rolesSvc.HasPermission(roles.UserAssign(userID), projectScope, requiredPerm)
		return err == nil && has
	}
	return svc.permSvc.CheckInherentAccess(action, isAuthor, hasPermFn), nil
}
