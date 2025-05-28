// package session_manager
package session_manager

import "time"

// Session represents an active communication session.
// It holds information about the session's state, participants, and associated metadata.
type Session struct {
	ID         string    // Unique identifier for the session
	Type       string    // Type of the session, e.g., "SIP", "WebRTC"
	State      string    // Current state of the session, e.g., "initializing", "active", "ringing", "ended"
	StartTime  time.Time // Timestamp when the session was initiated
	EndTime    time.Time // Timestamp when the session ended; zero value if active
	CallerInfo string    // Information about the caller
	CalleeInfo string    // Information about the callee
	GatewayID  string    // ID of the specific gateway instance (SIP or WebRTC) handling the session
	Metadata   map[string]string // Arbitrary key-value pairs for additional session data
}
