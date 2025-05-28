package session_manager

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// StartAPIServer initializes and starts the HTTP API server for the Session Manager.
func StartAPIServer(cfg SessionManagerConfig, store *SessionStore) {
	mux := http.NewServeMux()

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreateSession(w, r, store)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		// Path: /sessions/{id} or /sessions/{id}/state or /sessions/{id}/details
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/sessions/"), "/")
		sessionID := parts[0]
		if sessionID == "" {
			http.Error(w, "Session ID is required", http.StatusBadRequest)
			return
		}

		if len(parts) == 1 { // /sessions/{id}
			switch r.Method {
			case http.MethodGet:
				handleGetSession(w, r, store, sessionID)
			case http.MethodDelete:
				handleDeleteSession(w, r, store, sessionID)
			default:
				http.Error(w, "Method Not Allowed for /sessions/{id}", http.StatusMethodNotAllowed)
			}
		} else if len(parts) == 2 { // /sessions/{id}/state or /sessions/{id}/details
			subResource := parts[1]
			switch subResource {
			case "state":
				if r.Method == http.MethodPut {
					handleUpdateSessionState(w, r, store, sessionID)
				} else {
					http.Error(w, "Method Not Allowed for /sessions/{id}/state", http.StatusMethodNotAllowed)
				}
			case "details":
				if r.Method == http.MethodPut {
					handleUpdateSessionDetails(w, r, store, sessionID)
				} else {
					http.Error(w, "Method Not Allowed for /sessions/{id}/details", http.StatusMethodNotAllowed)
				}
			default:
				http.NotFound(w, r)
			}
		} else {
			http.NotFound(w, r)
		}
	})

	listenAddr := fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ApiPort)
	log.Printf("Session Manager API server starting on http://%s\n", listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("Failed to start Session Manager API server: %v", err)
	}
}

type CreateSessionRequest struct {
	ID      string            `json:"id"`
	Type    string            `json:"type"`
	Details map[string]string `json:"details,omitempty"`
}

func handleCreateSession(w http.ResponseWriter, r *http.Request, store *SessionStore) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.ID == "" || req.Type == "" {
		http.Error(w, "Session ID and Type are required", http.StatusBadRequest)
		return
	}

	session, err := store.CreateSession(req.ID, req.Type, req.Details)
	if err != nil {
		// Check if it's a duplicate ID error
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
		} else {
			http.Error(w, "Failed to create session: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func handleGetSession(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	session, err := store.GetSession(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

type UpdateStateRequest struct {
	State SessionState `json:"state"`
}

func handleUpdateSessionState(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	var req UpdateStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.State == "" {
		http.Error(w, "State field is required", http.StatusBadRequest)
		return
	}
	// Basic validation for state value
	switch req.State {
	case StatePending, StateActive, StateTerminated:
		// valid
	default:
		http.Error(w, fmt.Sprintf("Invalid state value: %s. Must be one of 'pending', 'active', 'terminated'", req.State), http.StatusBadRequest)
		return
	}


	session, err := store.UpdateSessionState(sessionID, req.State)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

type UpdateDetailsRequest struct {
	Details map[string]string `json:"details"`
}

func handleUpdateSessionDetails(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	var req UpdateDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Details == nil {
		http.Error(w, "Details field is required and cannot be null", http.StatusBadRequest)
		return
	}


	session, err := store.UpdateSessionDetails(sessionID, req.Details)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func handleDeleteSession(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	err := store.DeleteSession(sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
