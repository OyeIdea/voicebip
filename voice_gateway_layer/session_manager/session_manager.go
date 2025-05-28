// package session_manager
package session_manager

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid" // For generating session IDs
)

// SessionManagerService manages active communication sessions.
type SessionManagerService struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
}

// NewSessionManagerService creates and initializes a new SessionManagerService instance.
func NewSessionManagerService() *SessionManagerService {
	return &SessionManagerService{
		sessions: make(map[string]*Session),
		// Mutex is zero-valued and ready to use.
	}
}

// CreateSession generates a new unique session ID and creates a new Session instance.
func (sm *SessionManagerService) CreateSession(sessionType string, callerInfo string, calleeInfo string, gatewayID string) (*Session, error) {
	sessionID := uuid.New().String()
	session := &Session{
		ID:         sessionID,
		Type:       sessionType,
		State:      "initializing", // Initial state
		StartTime:  time.Now(),
		CallerInfo: callerInfo,
		CalleeInfo: calleeInfo,
		GatewayID:  gatewayID,
		Metadata:   make(map[string]string), // Initialize metadata map
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.sessions[sessionID] = session

	fmt.Printf("SessionManager: Created session %s of type %s\n", sessionID, sessionType)
	return session, nil
}

// GetSession retrieves a session by its ID.
func (sm *SessionManagerService) GetSession(sessionID string) (*Session, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// UpdateSessionState updates the state of an existing session.
func (sm *SessionManagerService) UpdateSessionState(sessionID string, newState string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	session.State = newState
	fmt.Printf("SessionManager: Updated state of session %s to %s\n", sessionID, newState)
	return nil
}

// UpdateSessionMetadata updates or adds key-value pairs to the session's metadata.
func (sm *SessionManagerService) UpdateSessionMetadata(sessionID string, metadata map[string]string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	if session.Metadata == nil { // Ensure metadata map is initialized if it wasn't
		session.Metadata = make(map[string]string)
	}
	for key, value := range metadata {
		session.Metadata[key] = value
	}
	fmt.Printf("SessionManager: Updated metadata for session %s\n", sessionID)
	return nil
}

// EndSession marks a session as ended and sets its end time.
func (sm *SessionManagerService) EndSession(sessionID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	session.State = "ended"
	session.EndTime = time.Now()
	fmt.Printf("SessionManager: Ended session %s\n", sessionID)
	// For now, we mark as ended. Deletion strategy can be added later (e.g., based on EndTime).
	return nil
}

// ListSessions returns a slice of active sessions, optionally filtered by type.
func (sm *SessionManagerService) ListSessions(sessionTypeFilter string) []*Session {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var filteredSessions []*Session
	for _, session := range sm.sessions {
		if sessionTypeFilter == "" || session.Type == sessionTypeFilter {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

// Example Usage (conceptual):
// func main() {
//     sm := NewSessionManagerService()
//     session, _ := sm.CreateSession("SIP", "caller123", "callee456", "sip-gw-1")
//     fmt.Printf("Created session: %+v\n", session)
//
//     metadata := map[string]string{"key1": "value1", "language": "en"}
//     sm.UpdateSessionMetadata(session.ID, metadata)
//
//     retrievedSession, _ := sm.GetSession(session.ID)
//     fmt.Printf("Retrieved session: %+v\n", retrievedSession)
//
//     sm.UpdateSessionState(session.ID, "active")
//     fmt.Printf("Session state updated. Current: %+v\n", retrievedSession)
//
//     sm.EndSession(session.ID)
//     fmt.Printf("Session ended. Current: %+v\n", retrievedSession)
//
//     fmt.Printf("All sessions: %+v\n", sm.ListSessions(""))
//     fmt.Printf("SIP sessions: %+v\n", sm.ListSessions("SIP"))
// }
