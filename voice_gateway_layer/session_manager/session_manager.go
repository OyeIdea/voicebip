package session_manager

import (
	"fmt"
	"sync"
	"time"
)

// SessionStore holds active sessions and provides thread-safe access.
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionStore creates and returns a new SessionStore.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// CreateSession creates a new session and adds it to the store.
// It sets the initial state to StatePending.
func (ss *SessionStore) CreateSession(id string, sessionType string, details map[string]string) (*Session, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, exists := ss.sessions[id]; exists {
		return nil, fmt.Errorf("session with ID '%s' already exists", id)
	}

	now := time.Now()
	session := &Session{
		ID:        id,
		Type:      sessionType,
		State:     StatePending, // Default to pending, can be updated to active immediately if needed
		CreatedAt: now,
		UpdatedAt: now,
		Details:   details,
	}

	ss.sessions[id] = session
	return session, nil
}

// GetSession retrieves a session by its ID.
func (ss *SessionStore) GetSession(id string) (*Session, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	session, exists := ss.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session with ID '%s' not found", id)
	}
	return session, nil
}

// UpdateSessionState updates the state of an existing session.
func (ss *SessionStore) UpdateSessionState(id string, state SessionState) (*Session, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session, exists := ss.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session with ID '%s' not found", id)
	}

	// Validate state transition if necessary (not implemented here for simplicity)
	session.State = state
	session.UpdatedAt = time.Now()
	ss.sessions[id] = session
	return session, nil
}

// UpdateSessionDetails updates the details of an existing session.
// This merges the new details with existing ones. For overwriting specific fields
// or removing them, more specific logic would be needed.
func (ss *SessionStore) UpdateSessionDetails(id string, newDetails map[string]string) (*Session, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session, exists := ss.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session with ID '%s' not found", id)
	}

	if session.Details == nil {
		session.Details = make(map[string]string)
	}
	for key, value := range newDetails {
		session.Details[key] = value
	}
	session.UpdatedAt = time.Now()
	ss.sessions[id] = session
	return session, nil
}

// DeleteSession removes a session from the store.
func (ss *SessionStore) DeleteSession(id string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, exists := ss.sessions[id]; !exists {
		return nil, fmt.Errorf("session with ID '%s' not found", id)
	}
	delete(ss.sessions, id)
	return nil
}

// ListSessions returns a slice of all current sessions.
// Useful for debugging or admin interfaces, but use with caution on very large stores.
func (ss *SessionStore) ListSessions() []*Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	sessions := make([]*Session, 0, len(ss.sessions))
	for _, session := range ss.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// Main function to start the session manager service (includes API server).
// func main() {
// 	cfg := LoadConfig()
// 	store := NewSessionStore()
// 	log.Printf("Session Manager attempting to start API server on %s:%d\n", cfg.ListenAddress, cfg.ApiPort)
// 	StartAPIServer(cfg, store)
// }
