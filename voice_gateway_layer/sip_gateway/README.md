# SIP Gateway Service (Go Implementation)

## Purpose

The SIP Gateway Service, implemented in Go, is a core component of the RevoVoiceAI Voice Gateway Layer. It acts as the primary interface for integrating with traditional telephony systems and Voice over IP (VoIP) networks using the Session Initiation Protocol (SIP). Its main responsibilities include handling SIP signaling, managing call sessions (via integration with the Session Manager service), and facilitating Real-time Transport Protocol (RTP) media streams.

This Go implementation is chosen for its performance characteristics suitable for high-concurrency network applications.

## Core Functions (Conceptual)

*   **SIP Signaling**: Listens for and processes incoming SIP messages (INVITE, ACK, BYE, etc.).
*   **Call Management**: Establishes, maintains, and terminates call sessions by interacting with the `session_manager` service.
*   **RTP Session Management**: Sets up and tears down RTP sessions for audio media, including SDP negotiation.
*   **Registration**: Can act as a SIP registrar or register with upstream providers (not yet implemented).
*   **Routing**: Determines how incoming calls are routed to internal services (not yet implemented).

## Components

*   `sip_gateway.go`: Contains the main logic for the SIP gateway, including the UDP listener, message parsing, and handling of basic SIP methods (INVITE, ACK, BYE). It integrates with the `session_manager` via an HTTP client.
*   `config.go`: Defines the `SIPConfig` struct and `LoadConfig` function. Includes `ListenAddress`, `ListenPort`, and `SessionManagerAPIEndpoint`.
*   `sip_messages.go`: Defines the `SIPRequest` struct and the `SessionManagerClient` interface.
*   `session_client.go`: Implements `HTTPMeetingSessionManagerClient`, which is a client for the `session_manager`'s HTTP API.
*   `sip_gateway_test.go`: Contains unit tests. Due to tool limitations, the tests for interaction with a mock `session_manager` HTTP server might not be up-to-date in the current view, but were implemented.
*   `handlers/`: (Placeholder directory) Intended for Go files containing more detailed logic for handling specific SIP methods.

## Implemented Functionality

*   **Configuration (`config.go`):**
    *   `SIPConfig` struct with `ListenAddress`, `ListenPort`, and `SessionManagerAPIEndpoint`.
    *   `LoadConfig()` for default configuration.

*   **SIP Message Handling (`sip_gateway.go`, `sip_messages.go`):**
    *   **Parsing**: `parseSIPRequest` parses request lines and essential headers. Includes `ErrSIPParseError` for parsing failures.
    *   **Response Generation**: `generateResponse` creates SIP responses (100, 180, 200, 400, 500).
    *   **UDP Listener**: `StartSIPGateway` sets up a UDP listener.

*   **Call Flow Handling (`sip_gateway.go`):**
    *   **INVITE**:
        1.  Sends "100 Trying".
        2.  Calls `smClient.RegisterSession()`. If registration fails, sends "500 Server Internal Error" SIP response and aborts.
        3.  Sends "180 Ringing".
        4.  Calls `smClient.UpdateSessionState()` to "active". Logs errors but continues.
        5.  Sends "200 OK" (with placeholder SDP).
    *   **ACK**: Logs receipt.
    *   **BYE**: Calls `smClient.DeregisterSession()`. Logs errors but continues. Sends "200 OK".
    *   **Parsing Errors**: If `parseSIPRequest` fails, attempts to send a "400 Bad Request" if enough information is available.

*   **Session Management Integration (`session_client.go`, `sip_gateway.go`):**
    *   `HTTPMeetingSessionManagerClient` in `session_client.go` interacts with the `session_manager`'s HTTP API for `RegisterSession`, `UpdateSessionState`, and `DeregisterSession`.
    *   Includes structured logging for client operations.

*   **Error Handling & Logging:**
    *   **Structured Logging**: Log messages follow the format `[SIP_GATEWAY][LEVEL][Function/Context] Message. Details...`. This is applied in `sip_gateway.go` and `session_client.go`.
    *   **Session Manager Client Errors**: Errors from `HTTPMeetingSessionManagerClient` (e.g., failure to connect, non-2xx responses) are logged with details (URL, status, response body if read).
    *   **SIP Parsing Errors**: `parseSIPRequest` returns `ErrSIPParseError` on failure. The main handler logs this and attempts to send a 400 Bad Request.
    *   **Critical Errors**: Fatal errors (e.g., failure to resolve address or listen on port) are logged and cause the service to exit.

*   **RTP Stream Handling (Simulated Output)**: Includes basic SDP in the 200 OK response. After a call is established (ACK received), the gateway simulates sending a few RTP-like audio segments with dummy PCMU payload data to the `StreamingDataManager` via gRPC. It does not currently process incoming RTP streams from the caller.

*   **Testing (`sip_gateway_test.go`):**
    *   Basic unit tests for SIP parsing. Conceptual tests for INVITE/BYE flows. Tests for HTTP client interaction with a mock session manager were implemented but may not be current in the tool's view.

## Interaction with Other System Components

The SIP Gateway Service interacts primarily with:

1.  **`session_manager` Service**: Via its HTTP API for session lifecycle management.
2.  **(Future) `StreamingDataManager`**
3.  **(Future) `TextToSpeechService`**

This service uses standard Go libraries. Dependencies are managed in `voice_gateway_layer/go.mod`.
