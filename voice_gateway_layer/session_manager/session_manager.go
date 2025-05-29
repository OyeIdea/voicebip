package session_manager

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	LogPrefix = "[SESSION_MANAGER]"
)

// Specific error types
var (
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionExists   = errors.New("session already exists")
)

// SessionStore holds active sessions and provides thread-safe access.
type SessionStore struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionStore creates and returns a new SessionStore.
func NewSessionStore() *SessionStore {
	log.Printf("%s[INFO][NewSessionStore] Initializing new session store", LogPrefix)
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
		log.Printf("%s[WARN][CreateSession] Attempted to create session with existing ID: %s", LogPrefix, id)
		return nil, fmt.Errorf("session ID '%s': %w", id, ErrSessionExists)
	}

	now := time.Now()
	session := &Session{
		ID:        id,
		Type:      sessionType,
		State:     StatePending,
		CreatedAt: now,
		UpdatedAt: now,
		Details:   details,
	}

	ss.sessions[id] = session
	log.Printf("%s[INFO][CreateSession] Session created successfully: ID=%s, Type=%s, Details=%v", LogPrefix, id, sessionType, details)
	return session, nil
}

// GetSession retrieves a session by its ID.
func (ss *SessionStore) GetSession(id string) (*Session, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	session, exists := ss.sessions[id]
	if !exists {
		log.Printf("%s[WARN][GetSession] Session not found: ID=%s", LogPrefix, id)
		return nil, fmt.Errorf("session ID '%s': %w", id, ErrSessionNotFound)
	}
	log.Printf("%s[INFO][GetSession] Session retrieved: ID=%s", LogPrefix, id)
	return session, nil
}

// UpdateSessionState updates the state of an existing session.
func (ss *SessionStore) UpdateSessionState(id string, state SessionState) (*Session, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session, exists := ss.sessions[id]
	if !exists {
		log.Printf("%s[WARN][UpdateSessionState] Session not found for state update: ID=%s", LogPrefix, id)
		return nil, fmt.Errorf("session ID '%s': %w", id, ErrSessionNotFound)
	}

	// TODO: Validate state transition if necessary
	previousState := session.State
	session.State = state
	session.UpdatedAt = time.Now()
	ss.sessions[id] = session // Not strictly necessary as it's a pointer, but good for clarity
	log.Printf("%s[INFO][UpdateSessionState] Session state updated: ID=%s, OldState=%s, NewState=%s", LogPrefix, id, previousState, state)
	return session, nil
}

// UpdateSessionDetails updates the details of an existing session.
func (ss *SessionStore) UpdateSessionDetails(id string, newDetails map[string]string) (*Session, error) {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	session, exists := ss.sessions[id]
	if !exists {
		log.Printf("%s[WARN][UpdateSessionDetails] Session not found for details update: ID=%s", LogPrefix, id)
		return nil, fmt.Errorf("session ID '%s': %w", id, ErrSessionNotFound)
	}

	if session.Details == nil {
		session.Details = make(map[string]string)
	}
	for key, value := range newDetails {
		session.Details[key] = value
	}
	session.UpdatedAt = time.Now()
	ss.sessions[id] = session // Not strictly necessary
	log.Printf("%s[INFO][UpdateSessionDetails] Session details updated: ID=%s, NewDetails=%v", LogPrefix, id, newDetails)
	return session, nil
}

// DeleteSession removes a session from the store.
func (ss *SessionStore) DeleteSession(id string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, exists := ss.sessions[id]; !exists {
		log.Printf("%s[WARN][DeleteSession] Session not found for deletion: ID=%s", LogPrefix, id)
		return fmt.Errorf("session ID '%s': %w", id, ErrSessionNotFound)
	}
	delete(ss.sessions, id)
	log.Printf("%s[INFO][DeleteSession] Session deleted successfully: ID=%s", LogPrefix, id)
	return nil
}

// ListSessions returns a slice of all current sessions.
func (ss *SessionStore) ListSessions() []*Session {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	sessions := make([]*Session, 0, len(ss.sessions))
	for _, session := range ss.sessions {
		sessions = append(sessions, session)
	}
	log.Printf("%s[INFO][ListSessions] Listed %d sessions", LogPrefix, len(sessions))
	return sessions
}

// Main function to start the session manager service (includes API server).
// func main() {
// 	log.Printf("%s[INFO][main] Starting Session Manager service...", LogPrefix)
// 	cfg := LoadConfig()
// 	store := NewSessionStore()
// 	log.Printf("%s[INFO][main] Session Manager API server attempting to start on %s:%d", LogPrefix, cfg.ListenAddress, cfg.ApiPort)
// 	StartAPIServer(cfg, store)
// }
