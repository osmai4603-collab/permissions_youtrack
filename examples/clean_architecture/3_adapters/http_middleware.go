package adapters

import (
	"context"
	"net/http"
	"strings"
	"youtrack/internal/services/rbac"
)

type contextKey string

const (
	ContextUserID      contextKey = "auth_user_id"
	ContextActiveRoles contextKey = "auth_active_roles"
	ContextSessionID   contextKey = "auth_session_id"
)

// AuthMiddleware represents the Edge Policy Enforcement Point (PEP).
// It verifies session tokens, extracts active roles, and performs fast-fail validation.
type AuthMiddleware struct {
	rbacService rbac.IRBACService
}

// NewAuthMiddleware creates a new edge PEP middleware.
func NewAuthMiddleware(rbacService rbac.IRBACService) *AuthMiddleware {
	return &AuthMiddleware{
		rbacService: rbacService,
	}
}

// AuthenticateAndExtractSession intercepts HTTP requests to validate the caller's session token.
func (m *AuthMiddleware) AuthenticateAndExtractSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		token := strings.TrimPrefix(authHeader, "Bearer ")
		token = strings.TrimSpace(token)

		if token == "" {
			http.Error(w, "missing or empty authorization token", http.StatusUnauthorized)
			return
		}

		// Validate session token with RBAC service
		sess, err := m.rbacService.GetSession(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid or expired session: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Collect active non-expired roles in this session
		var activeRoles []string
		for r := range sess.ActiveRoles {
			activeRoles = append(activeRoles, r)
		}

		// Inject security context into request
		ctx := context.WithValue(r.Context(), ContextUserID, sess.UserID)
		ctx = context.WithValue(ctx, ContextActiveRoles, activeRoles)
		ctx = context.WithValue(ctx, ContextSessionID, sess.ID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireCoarsePermission enforces an edge-level permission check (Fast-Fail)
// before executing the downstream controller or use case.
func (m *AuthMiddleware) RequireCoarsePermission(permission string, next http.HandlerFunc) http.HandlerFunc {
	return m.AuthenticateAndExtractSession(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(ContextUserID).(string)
		roles, _ := r.Context().Value(ContextActiveRoles).([]string)

		// Coarse-grained check on route entry
		err := m.rbacService.AuthorizePermission(r.Context(), userID, roles, permission)
		if err != nil {
			http.Error(w, "forbidden: "+err.Error(), http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
