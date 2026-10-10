package clean_architecture_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"youtrack/examples/clean_architecture/1_domain"
	"youtrack/examples/clean_architecture/2_usecases"
	"youtrack/examples/clean_architecture/3_adapters"
	"youtrack/internal/services/rbac"
)

// setupCleanArchitectureSystem wires all four layers via Dependency Injection.
func setupCleanArchitectureSystem(t *testing.T) (http.Handler, rbac.IRBACService, domain.IssueRepository) {
	ctx := context.Background()

	// 1. Initialize RBAC Security Subsystem (Domain & Application Service)
	rbacSvc := rbac.NewService(nil, nil)

	// Configure Role Hierarchy:
	// DEVELOPER -> has CREATE_ISSUE, READ_ISSUE, and DELETE_ISSUE
	// PROJECT_ADMIN -> inherits DEVELOPER
	devRole := &rbac.Role{
		ID:          "DEVELOPER",
		Name:        "Software Developer",
		Permissions: []string{"CREATE_ISSUE", "READ_ISSUE", "DELETE_ISSUE"},
	}
	adminRole := &rbac.Role{
		ID:          "PROJECT_ADMIN",
		Name:        "Project Administrator",
		Permissions: []string{},
		Parents:     []string{"DEVELOPER"},
	}

	if _, err := rbacSvc.CreateRole(ctx, devRole); err != nil {
		t.Fatalf("failed to create dev role: %v", err)
	}
	if _, err := rbacSvc.CreateRole(ctx, adminRole); err != nil {
		t.Fatalf("failed to create admin role: %v", err)
	}

	// Configure Hybrid ABAC Context Evaluators:
	// Global: Scope matching (ensures Project scope matches resource)
	// Specific to DELETE_ISSUE: Either Administrative role OR Inherent Ownership!
	cm := rbacSvc.GetContextManager()
	cm.AddGlobalRule(rbac.ScopeMatchEvaluator())
	cm.AddPermissionRule("DELETE_ISSUE", rbac.AnyOf(
		// Allow if user is admin role (which has authority across the project)
		func(ctx context.Context, req rbac.AccessRequest) (bool, string) {
			for _, r := range req.RoleIDs {
				if r == "PROJECT_ADMIN" {
					return true, "permitted via admin role"
				}
			}
			return false, "not admin"
		},
		// OR allow if user is inherent author!
		rbac.InherentOwnershipEvaluator(),
	))

	// 2. Outbound Repository Adapter (Layer 3 & 4)
	issueRepo := adapters.NewInMemoryIssueRepository()

	// 3. Application Use Cases (Layer 2)
	createUC := usecases.NewCreateIssueUseCase(issueRepo, rbacSvc)
	deleteUC := usecases.NewDeleteIssueUseCase(issueRepo, rbacSvc)

	// 4. Inbound HTTP Adapters & Middlewares (Layer 3)
	controller := adapters.NewIssueController(createUC, deleteUC)
	middleware := adapters.NewAuthMiddleware(rbacSvc)

	// 5. Framework & Drivers Router (Layer 4)
	mux := http.NewServeMux()
	mux.HandleFunc("/issues", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			middleware.AuthenticateAndExtractSession(controller.HandleCreateIssue)(w, r)
		case http.MethodDelete:
			middleware.AuthenticateAndExtractSession(controller.HandleDeleteIssue)(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux, rbacSvc, issueRepo
}

func TestCleanArchitecture_EndToEndRBACFlow(t *testing.T) {
	ctx := context.Background()
	handler, rbacSvc, repo := setupCleanArchitectureSystem(t)

	// Create Sessions:
	// Alice: DEVELOPER in project ALPHA
	aliceSess, err := rbacSvc.CreateSession(ctx, "alice", []string{"DEVELOPER"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Bob: DEVELOPER in project ALPHA
	bobSess, err := rbacSvc.CreateSession(ctx, "bob", []string{"DEVELOPER"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Charlie: PROJECT_ADMIN in project ALPHA
	charlieSess, err := rbacSvc.CreateSession(ctx, "charlie", []string{"PROJECT_ADMIN"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Unauthenticated Request -> 401 Unauthorized
	reqUnauth := httptest.NewRequest(http.MethodPost, "/issues", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, reqUnauth)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unauthenticated request, got %d", rr.Code)
	}

	// 2. Alice creates issue in Project ALPHA -> 201 Created
	payload := adapters.CreateIssueHTTPRequest{
		ID:          "issue-101",
		ProjectID:   "ALPHA",
		Title:       "Fix login redirect bug",
		Description: "Users are redirected to blank page",
	}
	body, _ := json.Marshal(payload)
	reqCreate := httptest.NewRequest(http.MethodPost, "/issues", bytes.NewReader(body))
	reqCreate.Header.Set("Authorization", "Bearer "+aliceSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, reqCreate)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for Alice, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify issue was persisted in domain repository
	savedIssue, err := repo.FindByID(ctx, "issue-101")
	if err != nil || savedIssue.AuthorID != "alice" {
		t.Fatalf("expected issue-101 with author alice in repository, got: %v", savedIssue)
	}

	// 3. Bob attempts to delete Alice's issue -> 403 Forbidden
	// (Bob has DEVELOPER role, but does NOT have DELETE_ISSUE and is NOT the author!)
	reqDeleteBob := httptest.NewRequest(http.MethodDelete, "/issues?id=issue-101", nil)
	reqDeleteBob.Header.Set("Authorization", "Bearer "+bobSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, reqDeleteBob)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when Bob tries to delete Alice's issue, got %d", rr.Code)
	}

	// 4. Charlie (PROJECT_ADMIN) deletes Alice's issue -> 204 No Content
	// (Permitted via role inheritance: PROJECT_ADMIN inherits DELETE_ISSUE)
	reqDeleteCharlie := httptest.NewRequest(http.MethodDelete, "/issues?id=issue-101", nil)
	reqDeleteCharlie.Header.Set("Authorization", "Bearer "+charlieSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, reqDeleteCharlie)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for Project Admin Charlie, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify issue deleted from repository
	if _, err := repo.FindByID(ctx, "issue-101"); err != domain.ErrIssueNotFound {
		t.Errorf("expected issue-101 to be deleted from repository")
	}

	// 5. Inherent Ownership: Bob creates his own issue, and deletes his OWN issue
	payloadBob := adapters.CreateIssueHTTPRequest{
		ID:          "issue-102",
		ProjectID:   "ALPHA",
		Title:       "Bob's personal task",
		Description: "Task by Bob",
	}
	bodyBob, _ := json.Marshal(payloadBob)
	reqCreateBob := httptest.NewRequest(http.MethodPost, "/issues", bytes.NewReader(bodyBob))
	reqCreateBob.Header.Set("Authorization", "Bearer "+bobSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, reqCreateBob)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created for Bob, got %d", rr.Code)
	}

	// Bob deletes his OWN issue -> 204 No Content (Permitted via Inherent Ownership!)
	reqDeleteBobOwn := httptest.NewRequest(http.MethodDelete, "/issues?id=issue-102", nil)
	reqDeleteBobOwn.Header.Set("Authorization", "Bearer "+bobSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, reqDeleteBobOwn)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for Bob deleting his own issue, got %d: %s", rr.Code, rr.Body.String())
	}

	// Verify issue-102 was deleted from repository
	if _, err := repo.FindByID(ctx, "issue-102"); err != domain.ErrIssueNotFound {
		t.Errorf("expected issue-102 to be deleted from repository")
	}
}

func TestCleanArchitecture_JITTemporaryElevation(t *testing.T) {
	ctx := context.Background()
	handler, rbacSvc, repo := setupCleanArchitectureSystem(t)

	// Dave has a session initially with only DEVELOPER role
	daveSess, err := rbacSvc.CreateSession(ctx, "dave", []string{"DEVELOPER"}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	// Pre-seed an issue
	issue, _ := domain.NewIssue("issue-999", "ALPHA", "Critical Bug", "Bug", "other_user")
	_ = repo.Save(ctx, issue)

	// Dave tries to delete -> 403 Forbidden
	req1 := httptest.NewRequest(http.MethodDelete, "/issues?id=issue-999", nil)
	req1.Header.Set("Authorization", "Bearer "+daveSess.ID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req1)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for Dave without elevation, got %d", rr.Code)
	}

	// Elevate Dave with JIT access to PROJECT_ADMIN for 40 milliseconds
	err = rbacSvc.ActivateRoleInSession(ctx, daveSess.ID, "PROJECT_ADMIN", 40*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to activate JIT role: %v", err)
	}

	// Immediate retry during JIT window -> 204 No Content (Permitted!)
	req2 := httptest.NewRequest(http.MethodDelete, "/issues?id=issue-999", nil)
	req2.Header.Set("Authorization", "Bearer "+daveSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req2)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content during JIT window, got %d: %s", rr.Code, rr.Body.String())
	}

	// Wait for JIT window to expire
	time.Sleep(50 * time.Millisecond)

	// Re-seed another issue
	issue2, _ := domain.NewIssue("issue-888", "ALPHA", "Another Bug", "Bug", "other_user")
	_ = repo.Save(ctx, issue2)

	// Subsequent delete after JIT expiration -> 403 Forbidden (Auto-evicted!)
	req3 := httptest.NewRequest(http.MethodDelete, "/issues?id=issue-888", nil)
	req3.Header.Set("Authorization", "Bearer "+daveSess.ID)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req3)
	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden after JIT expiration, got %d", rr.Code)
	}
}
