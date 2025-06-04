package sip_gateway

import "net/url"

// SIPRequest represents a parsed SIP request.
type SIPRequest struct {
	Method     string
	RequestURI *url.URL
	SIPVersion string
	Headers    map[string]string
	Body       []byte
}

// SIPResponse represents a SIP response.
// For simplicity, we'll focus on generating responses as strings directly
// rather than a full struct for responses in this initial implementation.

// Headers commonly used in SIP.
const (
	HeaderTo      = "To"
	HeaderFrom    = "From"
	HeaderCallID  = "Call-ID"
	HeaderCSeq    = "CSeq"
	HeaderContact = "Contact"
	HeaderVia     = "Via"
	HeaderMaxForwards = "Max-Forwards"
	HeaderContentType = "Content-Type"
	HeaderContentLength = "Content-Length"
)

// SessionManagerClient defines the interface for interacting with a session manager.
type SessionManagerClient interface {
	RegisterSession(callID string, sessionType string, details map[string]string) error
	UpdateSessionState(callID string, state string) error // Added state parameter
	DeregisterSession(callID string) error
}
