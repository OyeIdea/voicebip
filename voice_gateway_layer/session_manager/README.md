# Session Manager Service (Go Implementation)

## Purpose

The Session Manager Service, implemented in Go, is a component of the RevoVoiceAI Voice Gateway Layer. Its primary role is to provide centralized tracking and management of active voice sessions across different gateway services (e.g., SIP Gateway, WebRTC Gateway). This allows for a unified view of all ongoing interactions, facilitates routing decisions, and can be a source for real-time monitoring or billing information.

## Core Functions

*   **Session Lifecycle Management**: Creates, retrieves, updates (state, metadata), and ends sessions.
*   **Unique Session Identification**: Assigns and manages unique IDs for each session.
*   **State Tracking**: Maintains the current state of each session (e.g., "initializing", "active", "ringing", "ended").
*   **Metadata Storage**: Allows arbitrary key-value metadata to be associated with each session.
*   **Session Listing/Querying**: Provides a way to list active sessions, potentially with filters.

## Components

*   `session.go`: Defines the `Session` struct, which holds all information pertaining to a single voice session (ID, type, state, caller/callee info, timestamps, metadata, etc.).
*   `session_manager.go`: Contains the `SessionManagerService` struct and its methods. This service manages a collection of `Session` objects, providing thread-safe operations to interact with them.
*   `config.go`: (Placeholder) Intended for configurations related to session management, such as session timeout policies or parameters for external storage if session data becomes too large to hold in memory.

## Interaction with Other Services

Gateway services like `SIPGatewayService` and `WebRTCGatewayService` would interact with the `SessionManagerService` at various points in a call's lifecycle:

1.  **Session Creation**: When a new call or voice interaction is initiated (e.g., an incoming SIP INVITE or a new WebRTC peer connection), the respective gateway service calls `SessionManagerService.CreateSession()` to register the new session. The returned session ID is then used for future updates.
2.  **State Updates**: As the call progresses (e.g., ringing, answered, on hold), the gateway service informs the `SessionManagerService` by calling `UpdateSessionState()`.
3.  **Metadata Updates**: Any relevant metadata gathered during the call (e.g., specific call routing information, user identifiers) can be added using `UpdateSessionMetadata()`.
4.  **Session Querying**: Other services (e.g., monitoring tools, administrative interfaces, or even advanced routing logic) might query the `SessionManagerService` using `GetSession()` or `ListSessions()` to get information about active calls.
5.  **Session Termination**: When a call ends, the gateway service calls `EndSession()` to mark the session as ended and record its end time.

This centralized approach simplifies the individual gateway services, as they don't need to maintain comprehensive session state themselves, and provides a single source of truth for session information within the Voice Gateway Layer.
