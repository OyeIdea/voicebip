package sip_gateway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	SipGatewayLogPrefix = "[SIP_GATEWAY]"
)

// CreateSessionRequest mirrors the structure expected by the session_manager API.
type CreateSessionRequest struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Details map[string]string `json:"details,omitempty"`
}

// UpdateStateRequest mirrors the structure expected by the session_manager API.
type UpdateStateRequest struct {
	State string `json:"state"` // Using string directly as SessionState type is in another package
}

// HTTPMeetingSessionManagerClient implements SessionManagerClient for HTTP interaction.
type HTTPMeetingSessionManagerClient struct {
	Endpoint string
	Client   *http.Client
}

// NewHTTPMeetingSessionManagerClient creates a new client for the session manager API.
func NewHTTPMeetingSessionManagerClient(apiEndpoint string) *HTTPMeetingSessionManagerClient {
	log.Printf("%s[INFO][NewHTTPMeetingSessionManagerClient] Initializing Session Manager client for endpoint: %s", SipGatewayLogPrefix, apiEndpoint)
	return &HTTPMeetingSessionManagerClient{
		Endpoint: apiEndpoint,
		Client:   &http.Client{Timeout: 5 * time.Second},
	}
}

// RegisterSession registers a new session with the session_manager.
func (c *HTTPMeetingSessionManagerClient) RegisterSession(callID string, sessionType string, details map[string]string) error {
	log.Printf("%s[INFO][RegisterSession] Attempting to register session: CallID=%s, Type=%s", SipGatewayLogPrefix, callID, sessionType)
	requestBody := CreateSessionRequest{
		ID:      callID,
		Type:    sessionType,
		Details: details,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("%s[ERROR][RegisterSession] Failed to marshal request: CallID=%s, Error: %v", SipGatewayLogPrefix, callID, err)
		return fmt.Errorf("failed to marshal register session request: %w", err)
	}

	url := fmt.Sprintf("%s/sessions", c.Endpoint)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("%s[ERROR][RegisterSession] Failed to create POST request: URL=%s, CallID=%s, Error: %v", SipGatewayLogPrefix, url, callID, err)
		return fmt.Errorf("failed to create POST request to %s: %w", url, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Printf("%s[ERROR][RegisterSession] Failed to execute POST request: URL=%s, CallID=%s, Error: %v", SipGatewayLogPrefix, url, callID, err)
		return fmt.Errorf("failed to POST to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body) // Read body for more context
		log.Printf("%s[ERROR][RegisterSession] Failed to register session: URL=%s, CallID=%s, Status=%s, Body=%s", SipGatewayLogPrefix, url, callID, resp.Status, string(bodyBytes))
		return fmt.Errorf("failed to register session, status: %s, url: %s, response: %s", resp.Status, url, string(bodyBytes))
	}
	log.Printf("%s[INFO][RegisterSession] Session registered successfully: CallID=%s", SipGatewayLogPrefix, callID)
	return nil
}

// UpdateSessionState updates the state of an existing session.
func (c *HTTPMeetingSessionManagerClient) UpdateSessionState(callID string, state string) error {
	log.Printf("%s[INFO][UpdateSessionState] Attempting to update session state: CallID=%s, State=%s", SipGatewayLogPrefix, callID, state)
	requestBody := UpdateStateRequest{State: state}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("%s[ERROR][UpdateSessionState] Failed to marshal request: CallID=%s, Error: %v", SipGatewayLogPrefix, callID, err)
		return fmt.Errorf("failed to marshal update state request: %w", err)
	}

	url := fmt.Sprintf("%s/sessions/%s/state", c.Endpoint, callID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("%s[ERROR][UpdateSessionState] Failed to create PUT request: URL=%s, CallID=%s, Error: %v", SipGatewayLogPrefix, url, callID, err)
		return fmt.Errorf("failed to create PUT request to %s: %w", url, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Printf("%s[ERROR][UpdateSessionState] Failed to execute PUT request: URL=%s, CallID=%s, Error: %v", SipGatewayLogPrefix, url, callID, err)
		return fmt.Errorf("failed to PUT to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("%s[ERROR][UpdateSessionState] Failed to update session state: URL=%s, CallID=%s, Status=%s, Body=%s", SipGatewayLogPrefix, url, callID, resp.Status, string(bodyBytes))
		return fmt.Errorf("failed to update session state, status: %s, url: %s, response: %s", resp.Status, url, string(bodyBytes))
	}
	log.Printf("%s[INFO][UpdateSessionState] Session state updated successfully: CallID=%s, State=%s", SipGatewayLogPrefix, callID, state)
	return nil
}

// DeregisterSession deregisters a session from the session_manager.
func (c *HTTPMeetingSessionManagerClient) DeregisterSession(callID string) error {
	log.Printf("%s[INFO][DeregisterSession] Attempting to deregister session: CallID=%s", SipGatewayLogPrefix, callID)
	url := fmt.Sprintf("%s/sessions/%s", c.Endpoint, callID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		log.Printf("%s[ERROR][DeregisterSession] Failed to create DELETE request: URL=%s, CallID=%s, Error: %v", SipGatewayLogPrefix, url, callID, err)
		return fmt.Errorf("failed to create DELETE request to %s: %w", url, err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Printf("%s[ERROR][DeregisterSession] Failed to execute DELETE request: URL=%s, CallID=%s, Error: %v", SipGatewayLogPrefix, url, callID, err)
		return fmt.Errorf("failed to DELETE to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		if resp.StatusCode == http.StatusNotFound {
			log.Printf("%s[WARN][DeregisterSession] Session not found during deregister, assuming already deleted: CallID=%s, URL=%s", SipGatewayLogPrefix, callID, url)
			return nil // Considered acceptable
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("%s[ERROR][DeregisterSession] Failed to deregister session: URL=%s, CallID=%s, Status=%s, Body=%s", SipGatewayLogPrefix, url, callID, resp.Status, string(bodyBytes))
		return fmt.Errorf("failed to deregister session, status: %s, url: %s, response: %s", resp.Status, url, string(bodyBytes))
	}
	log.Printf("%s[INFO][DeregisterSession] Session deregistered successfully: CallID=%s", SipGatewayLogPrefix, callID)
	return nil
}
