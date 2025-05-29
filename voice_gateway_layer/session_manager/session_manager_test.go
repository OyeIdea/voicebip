package session_manager

import (
	"bytes"
	"encoding/json"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// --- Test SessionStore ---

func TestNewSessionStore(t *testing.T) {
	store := NewSessionStore()
	if store.sessions == nil {
		t.Error("NewSessionStore sessions map should not be nil")
	}
}

func TestCreateSession(t *testing.T) {
	store := NewSessionStore()
	details := map[string]string{"key": "value"}
	session, err := store.CreateSession("sess1", "SIP", details)

	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	if session.ID != "sess1" {
		t.Errorf("Expected ID sess1, got %s", session.ID)
	}
	if session.Type != "SIP" {
		t.Errorf("Expected Type SIP, got %s", session.Type)
	}
	if session.State != StatePending {
		t.Errorf("Expected StatePending, got %s", session.State)
	}
	if !reflect.DeepEqual(session.Details, details) {
		t.Errorf("Expected Details %v, got %v", details, session.Details)
	}

	// Test creating a session with an existing ID
	_, err = store.CreateSession("sess1", "WebRTC", nil)
	if !errors.Is(err, ErrSessionExists) { // Check for specific error
		t.Errorf("Expected ErrSessionExists when creating session with duplicate ID, got %v", err)
	}
}

func TestGetSession(t *testing.T) {
	store := NewSessionStore()
	store.CreateSession("sess1", "SIP", nil)

	session, err := store.GetSession("sess1")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if session.ID != "sess1" {
		t.Errorf("Expected ID sess1, got %s", session.ID)
	}

	_, err = store.GetSession("nonexistent")
	if !errors.Is(err, ErrSessionNotFound) { // Check for specific error
		t.Errorf("Expected ErrSessionNotFound when getting non-existent session, got %v", err)
	}
}

func TestUpdateSessionState(t *testing.T) {
	store := NewSessionStore()
	store.CreateSession("sess1", "SIP", nil)

	updatedSession, err := store.UpdateSessionState("sess1", StateActive)
	if err != nil {
		t.Fatalf("UpdateSessionState failed: %v", err)
	}
	if updatedSession.State != StateActive {
		t.Errorf("Expected StateActive, got %s", updatedSession.State)
	}
	if updatedSession.UpdatedAt.IsZero() || updatedSession.UpdatedAt == updatedSession.CreatedAt {
		t.Error("UpdatedAt should be modified")
	}

	_, err = store.UpdateSessionState("nonexistent", StateActive)
	if !errors.Is(err, ErrSessionNotFound) { // Check for specific error
		t.Errorf("Expected ErrSessionNotFound when updating non-existent session, got %v", err)
	}
}

func TestUpdateSessionDetails(t *testing.T) {
	store := NewSessionStore()
	initialDetails := map[string]string{"initialKey": "initialValue"}
	store.CreateSession("sess1", "WebRTC", initialDetails)

	newDetails := map[string]string{"newKey": "newValue", "initialKey": "updatedValue"}
	updatedSession, err := store.UpdateSessionDetails("sess1", newDetails)
	if err != nil {
		t.Fatalf("UpdateSessionDetails failed: %v", err)
	}

	expectedDetails := map[string]string{"initialKey": "updatedValue", "newKey": "newValue"}
	if !reflect.DeepEqual(updatedSession.Details, expectedDetails) {
		t.Errorf("Expected Details %v, got %v", expectedDetails, updatedSession.Details)
	}
	if updatedSession.UpdatedAt.IsZero() || updatedSession.UpdatedAt == updatedSession.CreatedAt {
		t.Error("UpdatedAt should be modified")
	}

	// Test updating details for a session that initially had nil details
	store.CreateSession("sess2", "SIP", nil)
	detailsForSess2 := map[string]string{"k": "v"}
	updatedSess2, err := store.UpdateSessionDetails("sess2", detailsForSess2)
	if err != nil {
		t.Fatalf("UpdateSessionDetails for sess2 failed: %v", err)
	}
	if !reflect.DeepEqual(updatedSess2.Details, detailsForSess2) {
		t.Errorf("Expected Details %v for sess2, got %v", detailsForSess2, updatedSess2.Details)
	}

	_, err = store.UpdateSessionDetails("nonexistent", newDetails)
	if !errors.Is(err, ErrSessionNotFound) { // Check for specific error
		t.Errorf("Expected ErrSessionNotFound when updating details of non-existent session, got %v", err)
	}
}

func TestDeleteSession(t *testing.T) {
	store := NewSessionStore()
	store.CreateSession("sess1", "SIP", nil)

	err := store.DeleteSession("sess1")
	if err != nil {
		t.Fatalf("DeleteSession failed: %v", err)
	}

	_, err = store.GetSession("sess1")
	if !errors.Is(err, ErrSessionNotFound) { // Check for specific error
		t.Errorf("Expected ErrSessionNotFound when getting deleted session, got %v", err)
	}

	err = store.DeleteSession("nonexistent")
	if !errors.Is(err, ErrSessionNotFound) { // Check for specific error
		t.Errorf("Expected ErrSessionNotFound when deleting non-existent session, got %v", err)
	}
}

func TestListSessions(t *testing.T) {
	store := NewSessionStore()
	store.CreateSession("sess1", "SIP", nil)
	store.CreateSession("sess2", "WebRTC", map[string]string{"user": "alice"})

	sessions := store.ListSessions()
	if len(sessions) != 2 {
		t.Fatalf("Expected 2 sessions, got %d", len(sessions))
	}
	// Could add more checks to verify content, but length is a good start
}

// --- Test HTTP API Handlers ---

var testStore *SessionStore
var testServer *httptest.Server

func setupAPITests() {
	testStore = NewSessionStore()
	// Populate with some initial data for GET, PUT, DELETE tests
	testStore.CreateSession("existingSess1", "SIP", map[string]string{"caller": "123", "callee": "456"})
	testStore.CreateSession("existingSess2", "WebRTC", map[string]string{"user": "testuser"})

	// Re-initialize mux for each test suite run if tests modify global state (not the case here with testStore re-init)
	mux := http.NewServeMux()
	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			handleCreateSession(w, r, testStore)
		} else { http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed) }
	})
	mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/sessions/"), "/")
		sessionID := parts[0]
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet: handleGetSession(w, r, testStore, sessionID)
			case http.MethodDelete: handleDeleteSession(w, r, testStore, sessionID)
			default: http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		} else if len(parts) == 2 {
			subResource := parts[1]
			switch subResource {
			case "state": if r.Method == http.MethodPut { handleUpdateSessionState(w, r, testStore, sessionID) } else { http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed) }
			case "details": if r.Method == http.MethodPut { handleUpdateSessionDetails(w, r, testStore, sessionID) } else { http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed) }
			default: http.NotFound(w, r)
			}
		} else { http.NotFound(w, r) }
	})
	testServer = httptest.NewServer(mux)
}

func teardownAPITests() {
	testServer.Close()
}

func TestAPI_CreateSession(t *testing.T) {
	setupAPITests()
	defer teardownAPITests()

	payload := CreateSessionRequest{ID: "apiSess1", Type: "WebRTC", Details: map[string]string{"client": "browser"}}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post(testServer.URL+"/sessions", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("HTTP POST failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201 Created, got %d", resp.StatusCode)
	}

	var createdSession Session
	if err := json.NewDecoder(resp.Body).Decode(&createdSession); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if createdSession.ID != "apiSess1" {
		t.Errorf("Expected created session ID apiSess1, got %s", createdSession.ID)
	}
	if createdSession.Details["client"] != "browser" {
		t.Errorf("Expected detail client=browser, got %v", createdSession.Details)
	}
	
	// Test duplicate creation (expecting 409 Conflict)
	resp2, err := http.Post(testServer.URL+"/sessions", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		t.Fatalf("HTTP POST (duplicate) failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusConflict { // Check for 409
		t.Errorf("Expected status 409 Conflict for duplicate, got %d", resp2.StatusCode)
	}
}

func TestAPI_GetSession(t *testing.T) {
	setupAPITests()
	defer teardownAPITests()

	resp, err := http.Get(testServer.URL + "/sessions/existingSess1")
	if err != nil {
		t.Fatalf("HTTP GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var session Session
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if session.ID != "existingSess1" {
		t.Errorf("Expected session ID existingSess1, got %s", session.ID)
	}

	// Test Not Found (expecting 404)
	respNotFound, err := http.Get(testServer.URL + "/sessions/nonexistentSession")
	if err != nil {
		t.Fatalf("HTTP GET (not found) failed: %v", err)
	}
	defer respNotFound.Body.Close()
	if respNotFound.StatusCode != http.StatusNotFound { // Check for 404
		t.Errorf("Expected status 404 Not Found, got %d", respNotFound.StatusCode)
	}
}

func TestAPI_UpdateSessionState(t *testing.T) {
	setupAPITests()
	defer teardownAPITests()

	payload := UpdateStateRequest{State: StateActive}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, testServer.URL+"/sessions/existingSess1/state", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP PUT /state failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}

	var updatedSession Session
	if err := json.NewDecoder(resp.Body).Decode(&updatedSession); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if updatedSession.State != StateActive {
		t.Errorf("Expected state %s, got %s", StateActive, updatedSession.State)
	}

	// Test Not Found for update
	reqNotFound, _ := http.NewRequest(http.MethodPut, testServer.URL+"/sessions/nonexistentSession/state", bytes.NewBuffer(payloadBytes))
	reqNotFound.Header.Set("Content-Type", "application/json")
	respNotFound, err := http.DefaultClient.Do(reqNotFound)
	if err != nil {
		t.Fatalf("HTTP PUT /state (not found) failed: %v", err)
	}
	defer respNotFound.Body.Close()
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found for update on non-existent session, got %d", respNotFound.StatusCode)
	}
}

func TestAPI_UpdateSessionDetails(t *testing.T) {
	setupAPITests()
	defer teardownAPITests()

	payload := UpdateDetailsRequest{Details: map[string]string{"caller": "999", "newDetail": "added"}}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPut, testServer.URL+"/sessions/existingSess1/details", bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP PUT /details failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", resp.StatusCode)
	}
	var updatedSession Session
	if err := json.NewDecoder(resp.Body).Decode(&updatedSession); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	if updatedSession.Details["caller"] != "999" {
		t.Errorf("Expected detail caller=999, got %s", updatedSession.Details["caller"])
	}
	if updatedSession.Details["newDetail"] != "added" {
		t.Errorf("Expected detail newDetail=added, got %s", updatedSession.Details["newDetail"])
	}

	// Test Not Found for update
	reqNotFound, _ := http.NewRequest(http.MethodPut, testServer.URL+"/sessions/nonexistentSession/details", bytes.NewBuffer(payloadBytes))
	reqNotFound.Header.Set("Content-Type", "application/json")
	respNotFound, err := http.DefaultClient.Do(reqNotFound)
	if err != nil {
		t.Fatalf("HTTP PUT /details (not found) failed: %v", err)
	}
	defer respNotFound.Body.Close()
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found for update on non-existent session, got %d", respNotFound.StatusCode)
	}
}


func TestAPI_DeleteSession(t *testing.T) {
	setupAPITests()
	defer teardownAPITests()

	req, _ := http.NewRequest(http.MethodDelete, testServer.URL+"/sessions/existingSess2", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP DELETE failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204 No Content, got %d", resp.StatusCode)
	}

	// Verify it's deleted
	getResp, _ := http.Get(testServer.URL + "/sessions/existingSess2")
	if getResp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found after delete, got %d", getResp.StatusCode)
	}
	getResp.Body.Close()

	// Test Not Found for delete
	reqNotFound, _ := http.NewRequest(http.MethodDelete, testServer.URL+"/sessions/nonexistentSession", nil)
	respNotFound, err := http.DefaultClient.Do(reqNotFound)
	if err != nil {
		t.Fatalf("HTTP DELETE (not found) failed: %v", err)
	}
	defer respNotFound.Body.Close()
	if respNotFound.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 Not Found for delete on non-existent session, got %d", respNotFound.StatusCode)
	}
}
