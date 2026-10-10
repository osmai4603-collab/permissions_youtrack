package rbac

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// SessionManager oversees runtime security contexts, dynamic role activation,
// JIT expiration enforcement, and Dynamic Separation of Duties (DSD).
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	graph    *RoleGraph
	sodMgr   *SoDManager
}

// NewSessionManager creates a new SessionManager.
func NewSessionManager(graph *RoleGraph, sodMgr *SoDManager) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		graph:    graph,
		sodMgr:   sodMgr,
	}
}

// CreateSession initializes a new session for the specified user with an optional initial role set and TTL.
func (sm *SessionManager) CreateSession(userID string, initialRoles []RoleID, ttl time.Duration) (*Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("%w: user ID must not be empty", ErrInvalidInput)
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour // Default 24h session window
	}

	now := time.Now()
	expiresAt := now.Add(ttl)

	// Pre-validate initial roles against DSD
	if sm.sodMgr != nil && len(initialRoles) > 1 {
		var validated []RoleID
		for _, r := range initialRoles {
			if err := sm.sodMgr.ValidateSessionActivation(validated, r); err != nil {
				return nil, fmt.Errorf("%w: initial roles breach DSD: %v", err, initialRoles)
			}
			validated = append(validated, r)
		}
	}

	sessionID := generateSessionID()

	activeRoles := make(map[RoleID]time.Time, len(initialRoles))
	for _, r := range initialRoles {
		activeRoles[r] = expiresAt
	}

	sess := &Session{
		ID:          sessionID,
		UserID:      userID,
		ActiveRoles: activeRoles,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		Metadata:    make(map[string]string),
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[sessionID] = sess

	return cloneSession(sess), nil
}

// GetSession retrieves the session by ID, verifying validity and purging expired JIT roles.
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	now := time.Now()
	if !sess.IsActive(now) {
		delete(sm.sessions, sessionID)
		return nil, fmt.Errorf("%w: session %s has expired", ErrSessionExpired, sessionID)
	}

	// Purge expired active roles
	sm.purgeExpiredRolesLocked(sess, now)

	return cloneSession(sess), nil
}

// ActivateRole activates a role inside an active session.
// If ttl > 0, the role is activated with a temporary JIT lease.
// It strictly validates DSD constraints before activation.
func (sm *SessionManager) ActivateRole(sessionID string, role RoleID, ttl time.Duration) error {
	if role == "" {
		return fmt.Errorf("%w: role ID must not be empty", ErrInvalidInput)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	now := time.Now()
	if !sess.IsActive(now) {
		delete(sm.sessions, sessionID)
		return fmt.Errorf("%w: session %s has expired", ErrSessionExpired, sessionID)
	}

	// Purge expired roles first to reflect current active state
	sm.purgeExpiredRolesLocked(sess, now)

	// Collect currently active roles
	activeList := make([]RoleID, 0, len(sess.ActiveRoles))
	for r := range sess.ActiveRoles {
		activeList = append(activeList, r)
	}

	// Dynamic Separation of Duties (DSD) verification
	if sm.sodMgr != nil {
		if err := sm.sodMgr.ValidateSessionActivation(activeList, role); err != nil {
			return err
		}
	}

	// Determine role expiry
	roleExpiry := sess.ExpiresAt
	if ttl > 0 {
		candidate := now.Add(ttl)
		if candidate.Before(sess.ExpiresAt) {
			roleExpiry = candidate
		}
	}

	sess.ActiveRoles[role] = roleExpiry
	return nil
}

// DeactivateRole removes a role from the active roles of a session.
func (sm *SessionManager) DeactivateRole(sessionID string, role RoleID) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	delete(sess.ActiveRoles, role)
	return nil
}

// TerminateSession destroys an active session.
func (sm *SessionManager) TerminateSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.sessions[sessionID]; !exists {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}
	delete(sm.sessions, sessionID)
	return nil
}

// GetActiveRoles returns the list of non-expired roles active in the session.
func (sm *SessionManager) GetActiveRoles(sessionID string) ([]RoleID, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sess, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
	}

	now := time.Now()
	if !sess.IsActive(now) {
		delete(sm.sessions, sessionID)
		return nil, fmt.Errorf("%w: session %s has expired", ErrSessionExpired, sessionID)
	}

	sm.purgeExpiredRolesLocked(sess, now)

	roles := make([]RoleID, 0, len(sess.ActiveRoles))
	for r := range sess.ActiveRoles {
		roles = append(roles, r)
	}
	sort.Strings(roles)
	return roles, nil
}

// GetActivePermissions aggregates the effective permissions for all non-expired roles
// currently active in the session, resolved transitively via the role hierarchy graph.
func (sm *SessionManager) GetActivePermissions(sessionID string) ([]string, error) {
	activeRoles, err := sm.GetActiveRoles(sessionID)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{})
	for _, r := range activeRoles {
		if sm.graph != nil {
			effective, err := sm.graph.GetEffectivePermissions(r)
			if err == nil {
				for _, p := range effective {
					permSet[p] = struct{}{}
				}
			}
		}
	}

	result := make([]string, 0, len(permSet))
	for p := range permSet {
		result = append(result, p)
	}
	sort.Strings(result)
	return result, nil
}

// ListUserSessions returns all active sessions for a given user.
func (sm *SessionManager) ListUserSessions(userID string) []*Session {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var res []*Session
	for id, s := range sm.sessions {
		if s.UserID == userID {
			if s.IsActive(now) {
				sm.purgeExpiredRolesLocked(s, now)
				res = append(res, cloneSession(s))
			} else {
				delete(sm.sessions, id)
			}
		}
	}
	return res
}

// purgeExpiredRolesLocked removes roles whose activation TTL has passed.
func (sm *SessionManager) purgeExpiredRolesLocked(s *Session, now time.Time) {
	for r, expiry := range s.ActiveRoles {
		if !expiry.IsZero() && now.After(expiry) {
			delete(s.ActiveRoles, r)
		}
	}
}

// generateSessionID generates a cryptographically random session token.
func generateSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "sess_" + hex.EncodeToString(b)
}

// cloneSession creates a deep copy of a Session.
func cloneSession(s *Session) *Session {
	if s == nil {
		return nil
	}
	cp := *s
	if s.ActiveRoles != nil {
		cp.ActiveRoles = make(map[RoleID]time.Time, len(s.ActiveRoles))
		for k, v := range s.ActiveRoles {
			cp.ActiveRoles[k] = v
		}
	}
	if s.Metadata != nil {
		cp.Metadata = make(map[string]string, len(s.Metadata))
		for k, v := range s.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}
