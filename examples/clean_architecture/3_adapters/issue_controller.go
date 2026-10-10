package adapters

import (
	"encoding/json"
	"net/http"
	"strings"
	"youtrack/examples/clean_architecture/2_usecases"
)

type CreateIssueHTTPRequest struct {
	ID          string `json:"id"`
	ProjectID   string `json:"project_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// IssueController acts as the Primary/Driving Adapter, converting HTTP payloads
// into strongly-typed commands and invoking application use cases.
type IssueController struct {
	createUC *usecases.CreateIssueUseCase
	deleteUC *usecases.DeleteIssueUseCase
}

// NewIssueController creates a new issue controller.
func NewIssueController(createUC *usecases.CreateIssueUseCase, deleteUC *usecases.DeleteIssueUseCase) *IssueController {
	return &IssueController{
		createUC: createUC,
		deleteUC: deleteUC,
	}
}

// HandleCreateIssue parses JSON payload and invokes CreateIssueUseCase.
func (c *IssueController) HandleCreateIssue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(ContextUserID).(string)
	roles, _ := r.Context().Value(ContextActiveRoles).([]string)

	var reqBody CreateIssueHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	cmd := usecases.CreateIssueCommand{
		UserID:      userID,
		UserRoles:   roles,
		IssueID:     reqBody.ID,
		ProjectID:   reqBody.ProjectID,
		Title:       reqBody.Title,
		Description: reqBody.Description,
	}

	issue, err := c.createUC.Execute(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "authorization denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(issue)
}

// HandleDeleteIssue extracts issue ID from query parameter and invokes DeleteIssueUseCase.
func (c *IssueController) HandleDeleteIssue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, _ := r.Context().Value(ContextUserID).(string)
	roles, _ := r.Context().Value(ContextActiveRoles).([]string)

	issueID := r.URL.Query().Get("id")
	if strings.TrimSpace(issueID) == "" {
		http.Error(w, "missing issue id parameter", http.StatusBadRequest)
		return
	}

	cmd := usecases.DeleteIssueCommand{
		UserID:    userID,
		UserRoles: roles,
		IssueID:   issueID,
	}

	err := c.deleteUC.Execute(r.Context(), cmd)
	if err != nil {
		if strings.Contains(err.Error(), "authorization denied") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
