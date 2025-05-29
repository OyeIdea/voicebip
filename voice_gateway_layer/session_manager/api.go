package session_manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// StartAPIServer initializes and starts the HTTP API server for the Session Manager.
func StartAPIServer(cfg SessionManagerConfig, store *SessionStore) {
	mux := http.NewServeMux()

	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s[API][Request] Path: %s, Method: %s, From: %s", LogPrefix, r.URL.Path, r.Method, r.RemoteAddr)
		switch r.Method {
		case http.MethodPost:
			handleCreateSession(w, r, store)
		default:
			log.Printf("%s[API][WARN] MethodNotAllowed: Path=%s, Method=%s", LogPrefix, r.URL.Path, r.Method)
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s[API][Request] Path: %s, Method: %s, From: %s", LogPrefix, r.URL.Path, r.Method, r.RemoteAddr)
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/sessions/"), "/")
		sessionID := parts[0]
		if sessionID == "" {
			log.Printf("%s[API][WARN] BadRequest: Session ID is required. Path=%s", LogPrefix, r.URL.Path)
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
				log.Printf("%s[API][WARN] MethodNotAllowed: Path=%s, Method=%s", LogPrefix, r.URL.Path, r.Method)
				http.Error(w, "Method Not Allowed for /sessions/{id}", http.StatusMethodNotAllowed)
			}
		} else if len(parts) == 2 { // /sessions/{id}/state or /sessions/{id}/details
			subResource := parts[1]
			switch subResource {
			case "state":
				if r.Method == http.MethodPut {
					handleUpdateSessionState(w, r, store, sessionID)
				} else {
					log.Printf("%s[API][WARN] MethodNotAllowed: Path=%s, Method=%s", LogPrefix, r.URL.Path, r.Method)
					http.Error(w, "Method Not Allowed for /sessions/{id}/state", http.StatusMethodNotAllowed)
				}
			case "details":
				if r.Method == http.MethodPut {
					handleUpdateSessionDetails(w, r, store, sessionID)
				} else {
					log.Printf("%s[API][WARN] MethodNotAllowed: Path=%s, Method=%s", LogPrefix, r.URL.Path, r.Method)
					http.Error(w, "Method Not Allowed for /sessions/{id}/details", http.StatusMethodNotAllowed)
				}
			default:
				log.Printf("%s[API][WARN] NotFound: Path=%s", LogPrefix, r.URL.Path)
				http.NotFound(w, r)
			}
		} else {
			log.Printf("%s[API][WARN] NotFound: Path=%s", LogPrefix, r.URL.Path)
			http.NotFound(w, r)
		}
	})

	listenAddr := fmt.Sprintf("%s:%d", cfg.ListenAddress, cfg.ApiPort)
	log.Printf("%s[API][INFO] Session Manager API server starting on http://%s", LogPrefix, listenAddr)
	if err := http.ListenAndServe(listenAddr, mux); err != nil {
		log.Fatalf("%s[API][FATAL] Failed to start Session Manager API server: %v", LogPrefix, err)
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
		errMsg := fmt.Sprintf("Invalid request body: %v", err)
		log.Printf("%s[API][ERROR][handleCreateSession] %s", LogPrefix, errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.ID == "" || req.Type == "" {
		errMsg := "Session ID and Type are required"
		log.Printf("%s[API][WARN][handleCreateSession] %s. RequestID: %s, RequestType: %s", LogPrefix, errMsg, req.ID, req.Type)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	session, err := store.CreateSession(req.ID, req.Type, req.Details)
	if err != nil {
		if errors.Is(err, ErrSessionExists) {
			log.Printf("%s[API][WARN][handleCreateSession] Session already exists: ID=%s. Error: %v", LogPrefix, req.ID, err)
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			log.Printf("%s[API][ERROR][handleCreateSession] Failed to create session: ID=%s. Error: %v", LogPrefix, req.ID, err)
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("%s[API][INFO][handleCreateSession] Session created: ID=%s, Type=%s", LogPrefix, session.ID, session.Type)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func handleGetSession(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	session, err := store.GetSession(sessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			log.Printf("%s[API][WARN][handleGetSession] Session not found: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("%s[API][ERROR][handleGetSession] Failed to get session: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, "Failed to get session", http.StatusInternalServerError)
		}
		return
	}
	log.Printf("%s[API][INFO][handleGetSession] Session retrieved: ID=%s", LogPrefix, session.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

type UpdateStateRequest struct {
	State SessionState `json:"state"`
}

func handleUpdateSessionState(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	var req UpdateStateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg := fmt.Sprintf("Invalid request body: %v", err)
		log.Printf("%s[API][ERROR][handleUpdateSessionState] %s. SessionID: %s", LogPrefix, errMsg, sessionID)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.State == "" {
		errMsg := "State field is required"
		log.Printf("%s[API][WARN][handleUpdateSessionState] %s. SessionID: %s", LogPrefix, errMsg, sessionID)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	switch req.State {
	case StatePending, StateActive, StateTerminated:
		// valid
	default:
		errMsg := fmt.Sprintf("Invalid state value: %s. Must be one of '%s', '%s', '%s'", req.State, StatePending, StateActive, StateTerminated)
		log.Printf("%s[API][WARN][handleUpdateSessionState] %s. SessionID: %s", LogPrefix, errMsg, sessionID)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	session, err := store.UpdateSessionState(sessionID, req.State)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			log.Printf("%s[API][WARN][handleUpdateSessionState] Session not found: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("%s[API][ERROR][handleUpdateSessionState] Failed to update session state: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, "Failed to update session state", http.StatusInternalServerError)
		}
		return
	}
	log.Printf("%s[API][INFO][handleUpdateSessionState] Session state updated: ID=%s, NewState=%s", LogPrefix, session.ID, session.State)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

type UpdateDetailsRequest struct {
	Details map[string]string `json:"details"`
}

func handleUpdateSessionDetails(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	var req UpdateDetailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errMsg := fmt.Sprintf("Invalid request body: %v", err)
		log.Printf("%s[API][ERROR][handleUpdateSessionDetails] %s. SessionID: %s", LogPrefix, errMsg, sessionID)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Details == nil { // Note: empty map is valid, but null is not for this design.
		errMsg := "Details field is required and cannot be null"
		log.Printf("%s[API][WARN][handleUpdateSessionDetails] %s. SessionID: %s", LogPrefix, errMsg, sessionID)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	session, err := store.UpdateSessionDetails(sessionID, req.Details)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			log.Printf("%s[API][WARN][handleUpdateSessionDetails] Session not found: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("%s[API][ERROR][handleUpdateSessionDetails] Failed to update session details: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, "Failed to update session details", http.StatusInternalServerError)
		}
		return
	}
	log.Printf("%s[API][INFO][handleUpdateSessionDetails] Session details updated: ID=%s", LogPrefix, session.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func handleDeleteSession(w http.ResponseWriter, r *http.Request, store *SessionStore, sessionID string) {
	err := store.DeleteSession(sessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			log.Printf("%s[API][WARN][handleDeleteSession] Session not found: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			log.Printf("%s[API][ERROR][handleDeleteSession] Failed to delete session: ID=%s. Error: %v", LogPrefix, sessionID, err)
			http.Error(w, "Failed to delete session", http.StatusInternalServerError)
		}
		return
	}
	log.Printf("%s[API][INFO][handleDeleteSession] Session deleted: ID=%s", LogPrefix, sessionID)
	w.WriteHeader(http.StatusNoContent)
}
