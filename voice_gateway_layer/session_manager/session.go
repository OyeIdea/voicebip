package session_manager

import "time"

// SessionState defines the possible states of a session.
type SessionState string

const (
	// StatePending indicates the session is being set up.
	StatePending SessionState = "pending"
	// StateActive indicates the session is active and ongoing.
	StateActive SessionState = "active"
	// StateTerminated indicates the session has ended.
	StateTerminated SessionState = "terminated"
)

// Session represents a communication session (e.g., a call).
type Session struct {
	ID        string            `json:"id"` // Unique identifier for the session (e.g., Call-ID or a generated UUID)
	Type      string            `json:"type"` // Type of session, e.g., "SIP", "WebRTC"
	State     SessionState      `json:"state"`    // Current state of the session
	CreatedAt time.Time         `json:"createdAt"` // Timestamp when the session was created
	UpdatedAt time.Time         `json:"updatedAt"` // Timestamp when the session was last updated
	Details   map[string]string `json:"details,omitempty"` // Additional details about the session (e.g., From, To, custom data)
}
