# Session Manager Service (Go Implementation)

## Purpose

The Session Manager Service, implemented in Go, is a component of the RevoVoiceAI Voice Gateway Layer. Its primary role is to provide centralized tracking and management of active voice sessions across different gateway services (e.g., SIP Gateway, WebRTC Gateway). This allows for a unified view of all ongoing interactions, facilitates routing decisions, and can be a source for real-time monitoring or billing information.

## Implemented Functionality

### Session Data Structures (`session.go`)

*   **`SessionState` (string)**: Defines the state of a session.
    *   `StatePending` ("pending"): Session is being initialized.
    *   `StateActive` ("active"): Session is currently active.
    *   `StateTerminated` ("terminated"): Session has ended.
*   **`Session` struct**:
    ```go
    type Session struct {
        ID        string            `json:"id"`        // Unique identifier (e.g., Call-ID or UUID)
        Type      string            `json:"type"`      // "SIP", "WebRTC"
        State     SessionState      `json:"state"`     // Current state
        CreatedAt time.Time         `json:"createdAt"` // Creation timestamp
        UpdatedAt time.Time         `json:"updatedAt"` // Last update timestamp
        Details   map[string]string `json:"details,omitempty"` // Additional key-value details
    }
    ```

### In-Memory Session Store (`session_manager.go`)

*   **`SessionStore` struct**:
    *   Manages sessions in an in-memory map (`map[string]*Session`).
    *   Uses `sync.RWMutex` to ensure thread-safe access to the session map.
*   **Methods**:
    *   `NewSessionStore() *SessionStore`: Creates a new session store.
    *   `CreateSession(id string, sessionType string, details map[string]string) (*Session, error)`:
        *   Creates a new session with `StatePending`.
        *   Returns an error if a session with the given `id` already exists.
    *   `GetSession(id string) (*Session, error)`: Retrieves a session by ID. Returns error if not found.
    *   `UpdateSessionState(id string, state SessionState) (*Session, error)`: Updates the session's state and `UpdatedAt` timestamp. Returns error if not found.
    *   `UpdateSessionDetails(id string, details map[string]string) (*Session, error)`: Merges provided details with existing ones and updates `UpdatedAt`. Returns error if not found.
    *   `DeleteSession(id string) error`: Removes a session from the store. Returns error if not found.
    *   `ListSessions() []*Session`: Returns a slice of all current sessions (for debugging/admin).

### HTTP API (`api.go`)

The service exposes a RESTful HTTP API for managing sessions. It uses the standard `net/http` package.

*   **`POST /sessions`**: Creates a new session.
    *   **Request Body (JSON)**:
        ```json
        {
            "id": "call-12345",
            "type": "SIP",
            "details": {
                "caller_id": "1001",
                "callee_id": "2002"
            }
        }
        ```
    *   **Response (201 Created)**: Session JSON object.
        ```json
        {
            "id": "call-12345",
            "type": "SIP",
            "state": "pending",
            "createdAt": "2023-10-27T10:00:00Z",
            "updatedAt": "2023-10-27T10:00:00Z",
            "details": {
                "caller_id": "1001",
                "callee_id": "2002"
            }
        }
        ```
    *   **Response (409 Conflict)**: If session ID already exists (due to `ErrSessionExists`).
    *   **Response (400 Bad Request)**: If ID or Type is missing, or JSON is invalid.
    *   **Response (500 Internal Server Error)**: For other unexpected errors during creation.

*   **`GET /sessions/{id}`**: Retrieves a session by its ID.
    *   **Response (200 OK)**: Session JSON object.
    *   **Response (404 Not Found)**: If session ID does not exist (due to `ErrSessionNotFound`).
    *   **Response (500 Internal Server Error)**: For other unexpected errors.

*   **`PUT /sessions/{id}/state`**: Updates the state of an existing session.
    *   **Request Body (JSON)**:
        ```json
        {
            "state": "active"
        }
        ```
    *   **Response (200 OK)**: Updated session JSON object.
    *   **Response (404 Not Found)**: If session ID does not exist (due to `ErrSessionNotFound`).
    *   **Response (400 Bad Request)**: If state is missing or invalid (must be "pending", "active", or "terminated").
    *   **Response (500 Internal Server Error)**: For other unexpected errors.

*   **`PUT /sessions/{id}/details`**: Updates (merges) the details of an existing session.
    *   **Request Body (JSON)**:
        ```json
        {
            "details": {
                "custom_info": "updated_value",
                "another_key": "new_info"
            }
        }
        ```
    *   **Response (200 OK)**: Updated session JSON object.
    *   **Response (404 Not Found)**: If session ID does not exist (due to `ErrSessionNotFound`).
    *   **Response (400 Bad Request)**: If details field is missing or JSON is invalid.
    *   **Response (500 Internal Server Error)**: For other unexpected errors.

*   **`DELETE /sessions/{id}`**: Deletes a session by its ID.
    *   **Response (204 No Content)**: On successful deletion.
    *   **Response (404 Not Found)**: If session ID does not exist (due to `ErrSessionNotFound`).
    *   **Response (500 Internal Server Error)**: For other unexpected errors.

*   **API Server Initialization (`StartAPIServer`)**:
    *   Sets up HTTP routes using `http.NewServeMux()`.
    *   Starts the HTTP server on the configured address and port.

### Error Handling and Logging

*   **Specific Errors**: The session store (`session_manager.go`) now returns specific error types:
    *   `ErrSessionNotFound`: When a requested session ID is not found.
    *   `ErrSessionExists`: When attempting to create a session with an ID that already exists.
*   **HTTP Status Codes**: The API handlers (`api.go`) use `errors.Is()` to check for these specific errors and map them to appropriate HTTP status codes:
    *   `ErrSessionNotFound` maps to `HTTP 404 Not Found`.
    *   `ErrSessionExists` maps to `HTTP 409 Conflict`.
    *   Other unexpected errors generally map to `HTTP 500 Internal Server Error`.
*   **Structured Logging**: All log messages follow a structured format: `[SERVICE_NAME][LEVEL][Function_Name] Message. Details...`. For this service, `SERVICE_NAME` is `[SESSION_MANAGER]`. This applies to both store logic and API request handling, providing better context for debugging and monitoring.

### Configuration (`config.go`)

*   **`SessionManagerConfig` struct**:
    *   `ListenAddress` (string): The IP address for the API server to listen on (e.g., "0.0.0.0").
    *   `ApiPort` (int): The TCP port for the API server (e.g., 8000).
*   **`LoadConfig()` function**: Returns a default `SessionManagerConfig` ("0.0.0.0", port 8000).

### Testing (`session_manager_test.go`)

*   Unit tests for `SessionStore` methods covering creation, retrieval, state updates, detail updates, and deletion.
*   Basic HTTP API tests using `net/http/httptest` for:
    *   Creating a session (`POST /sessions`).
    *   Getting a session (`GET /sessions/{id}`).
    *   Updating session state (`PUT /sessions/{id}/state`).
    *   Updating session details (`PUT /sessions/{id}/details`).
    *   Deleting a session (`DELETE /sessions/{id}`).

## Interaction with Other Services

Gateway services like `SIPGatewayService` and `WebRTCGatewayService` would use the Session Manager's HTTP API to:

1.  **Session Creation**: `POST /sessions` when a new call starts.
2.  **State Updates**: `PUT /sessions/{id}/state` as the call progresses (e.g., "pending" -> "active", "active" -> "terminated").
3.  **Metadata Updates**: `PUT /sessions/{id}/details` to add or modify call-specific information.
4.  **Session Querying**: `GET /sessions/{id}` if they need to retrieve session information.
5.  **Session Termination**: `DELETE /sessions/{id}` when a call ends (or `PUT /sessions/{id}/state` with state "terminated").

This centralized approach simplifies the individual gateway services, as they don't need to maintain comprehensive session state themselves, and provides a single source of truth for session information within the Voice Gateway Layer. The `SessionManagerClient` interface previously conceptualized in `sip_gateway` and `webrtc_gateway` would be implemented by an HTTP client making requests to this Session Manager API.
No external dependencies beyond the standard Go library were added for this service. The `go.mod` file in `voice_gateway_layer/` remains unchanged by this specific service's implementation.
