# WebRTC Gateway Service (Go Implementation)

## Purpose

The WebRTC Gateway Service, implemented in Go, is a component of the RevoVoiceAI Voice Gateway Layer. It enables real-time voice (and potentially video) communication directly with web browsers and mobile applications using WebRTC (Web Real-Time Communication) standards. This allows for rich, interactive voice experiences without requiring users to install separate telephony software.

Go is chosen for its suitability in building high-performance, concurrent network services, and `github.com/pion/webrtc/v3` is used as the WebRTC library.

## Core Functions (Conceptual)

*   **Signaling Server**: Manages the exchange of session control messages (like SDP offers/answers and ICE candidates) between WebRTC peers using WebSockets.
*   **ICE/STUN/TURN Handling**: Facilitates NAT traversal.
*   **Peer Connection Management**: Establishes, monitors, and tears down WebRTC peer connections.
*   **Session Lifecycle Management**: Integrates with the `session_manager` service via its HTTP API to register, update, and deregister WebRTC sessions.
*   **Media Stream Handling**: Receives incoming audio RTP media streams (typically Opus format from WebRTC clients) via the `OnTrack` callback. These Opus RTP packets are then encapsulated into `AudioSegment` messages and forwarded to the `StreamingDataManager` service via gRPC.

## Components

*   `webrtc_gateway.go`: Contains `StartWebRTCGateway` which initializes the HTTP server for WebSocket signaling and the `SessionManagerClient`.
*   `config.go`: Defines `WebRTCConfig` (ListenAddress, SignalPort, StunServers, `SessionManagerAPIEndpoint`) and `LoadConfig`. `LoadConfig` now reads from environment variables with fallbacks.
*   `signal.go`: Implements `HandleWebSocketConnections` for WebSocket signaling. It handles WebRTC negotiation (offer/answer, ICE candidates) and interacts with the `session_manager` service.
*   `session_client.go`: Implements `HTTPMeetingSessionManagerClient`, a client for the `session_manager`'s HTTP API. Defines the `SessionManagerClient` interface.
*   `webrtc_gateway_test.go`: Contains unit tests, including tests for interaction with a mock `session_manager` HTTP server.
*   `handlers/`: (Placeholder directory).

## Implemented Functionality

*   **Configuration (`config.go`):**
    *   `WebRTCConfig` struct includes `ListenAddress`, `SignalPort`, `StunServers`, and `SessionManagerAPIEndpoint`.
    *   `LoadConfig()` loads these parameters from environment variables (see "Configuration via Environment Variables" section below) with hardcoded defaults if variables are not set.

*   **Signaling & WebRTC Core (`signal.go`):**
    *   Uses `github.com/gorilla/websocket` and `github.com/pion/webrtc/v3`.
    *   `HandleWebSocketConnections`:
        *   Generates a unique session ID (UUID).
        *   Handles SDP "offer", "answer", and ICE "candidate" messages.
        *   Sets up `OnICECandidate`, `OnICEConnectionStateChange`, and `OnTrack` callbacks.

*   **Session Management Integration (`session_client.go`, `signal.go`, `webrtc_gateway.go`):**
    *   `HTTPMeetingSessionManagerClient` interacts with the `session_manager`'s HTTP API.
    *   `RegisterSession` called on new WebSocket connection. Failure results in WebSocket error and closure.
    *   `UpdateSessionState` to "active" called on `ICEConnectionStateConnected`. Failures are logged.
    *   `DeregisterSession` called deferred on WebSocket handler exit. Failures are logged.

*   **Error Handling & Logging:**
    *   **Structured Logging**: Log messages across `webrtc_gateway.go`, `signal.go`, and `session_client.go` follow the format `[WEBRTC_GATEWAY][LEVEL][Function/Context] Message. Details...`.
    *   **Session Manager Client Errors**: Errors from `HTTPMeetingSessionManagerClient` are logged with details. Critical errors like `RegisterSession` failure result in an error message sent over WebSocket and connection termination.
    *   **Pion WebRTC API Errors**: Errors from `pion/webrtc` calls are logged with context.
    *   **WebSocket Errors**: Errors during WebSocket operations are logged.
    *   **Critical Errors**: Fatal errors in `StartWebRTCGateway` are logged and cause the service to exit.

*   **Gateway Initialization (`webrtc_gateway.go`):**
    *   `StartWebRTCGateway(cfg WebRTCConfig)` initializes clients and the HTTP server with structured logging.

*   **Testing (`webrtc_gateway_test.go`):**
    *   Includes `mockWebRTCSessionManagerServer`.
    *   `TestWebSocketSignaling_OfferAnswer_WithSessionManager` tests interactions with the mock session manager.

### Opus Audio Handling
The WebRTC gateway receives Opus-encoded RTP packets from the client via the `OnTrack` callback. The current implementation includes **placeholders and comments** in `signal.go` that outline where Opus decoding to PCM (specifically Linear16, e.g., at 16kHz, 16-bit signed) would be implemented, potentially using a Go Opus library like `github.com/pion/opus`.

Currently, for pipeline testing and development **without actual Opus decoding implemented yet**, the raw Opus packet data is still what's packaged into the `AudioSegment`'s data field. However, to facilitate downstream service configuration and testing, the `AudioSegment`'s `audio_format` field is now explicitly set to `LINEAR16` (e.g., assuming 16kHz, 16-bit signed PCM). This **simulates** the audio format that *would* be produced after the conceptual decoding step.

This approach allows services like the `SpeechToTextService` to be configured for the target `LINEAR16` PCM format. For a production system that needs to connect to STT engines expecting uncompressed PCM, **implementing the actual Opus decoding at the marked placeholders in `signal.go` is a crucial next step.**

## Configuration via Environment Variables

The service can be configured using the following environment variables. If an environment variable is not set or is empty, a default value will be used.

*   **`WEBRTC_LISTEN_ADDRESS`**: The IP address for the WebRTC gateway's signaling server to listen on.
    *   Default: `"0.0.0.0"`
*   **`WEBRTC_SIGNAL_PORT`**: The TCP port for the WebSocket signaling server.
    *   Default: `8080`
*   **`WEBRTC_STUN_SERVERS`**: A comma-separated list of STUN server URLs (e.g., `"stun:stun.l.google.com:19302,stun:another.stun.server:19302"`).
    *   Default: `"stun:stun.l.google.com:19302"`
*   **`WEBRTC_SESSION_MANAGER_API_ENDPOINT`**: The HTTP endpoint for the `session_manager` service.
    *   Default: `"http://localhost:8000"`

## Interaction with Other System Components

1.  **`session_manager` Service**: Via its HTTP API for session lifecycle management.
2.  **Web Clients**: Connect via WebSocket for signaling and establish WebRTC peer connections for media.
3.  **(Future) `StreamingDataManager`**: Would receive media from `OnTrack`.

The `go.mod` and `go.sum` files in `voice_gateway_layer` manage dependencies.
