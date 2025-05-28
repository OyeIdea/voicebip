# WebRTC Gateway Service (Go Implementation)

## Purpose

The WebRTC Gateway Service, implemented in Go, is a component of the RevoVoiceAI Voice Gateway Layer. It enables real-time voice (and potentially video) communication directly with web browsers and mobile applications using WebRTC (Web Real-Time Communication) standards. This allows for rich, interactive voice experiences without requiring users to install separate telephony software.

Go is chosen for its suitability in building high-performance, concurrent network services, and `github.com/pion/webrtc/v3` is used as the WebRTC library.

## Core Functions (Conceptual)

*   **Signaling Server**: Manages the exchange of session control messages (like SDP offers/answers and ICE candidates) between WebRTC peers. This is often implemented using WebSockets.
*   **ICE/STUN/TURN Handling**: Facilitates NAT (Network Address Translation) traversal to establish direct peer-to-peer media connections where possible, using STUN for NAT discovery and TURN for relaying media if direct connection fails.
*   **Peer Connection Management**: Establishes, monitors, and tears down WebRTC peer connections.
*   **Data Channel Management**: (Optional) Can manage WebRTC data channels for sending non-media data alongside audio/video.
*   **Media Stream Handling**: Receives incoming RTP media streams from WebRTC clients and forwards them to the Real-Time Processing Engine (e.g., `StreamingDataManager`), and vice-versa for outgoing audio.

## Components

*   `webrtc_gateway.go`: Contains the `StartWebRTCGateway` function, which initializes and starts the HTTP server for WebSocket signaling.
*   `config.go`: Defines the `WebRTCConfig` struct (ListenAddress, SignalPort, StunServers) and `LoadConfig` function for service configuration.
*   `signal.go`: Implements the WebSocket signaling handler (`HandleWebSocketConnections`). This includes logic for handling "offer", "answer", and "candidate" messages, and interacting with the `pion/webrtc` PeerConnection. It also defines the `SignalMessage` struct and a conceptual `SessionManagerClient` interface with a dummy implementation.
*   `webrtc_gateway_test.go`: Contains unit tests for `SignalMessage` marshalling/unmarshalling and a basic test for the WebSocket offer/answer signaling flow using a test server.
*   `handlers/`: (Placeholder directory) Intended for Go files containing more detailed logic for WebSocket connections or specific signaling protocol messages. *(This directory is not yet used in the current basic implementation).*

## Implemented Functionality

The current implementation provides a basic WebRTC gateway with WebSocket-based signaling:

*   **Configuration (`config.go`):**
    *   `WebRTCConfig` struct:
        *   `ListenAddress` (string): The IP address for the signaling server to listen on (e.g., "0.0.0.0").
        *   `SignalPort` (int): The TCP port for the WebSocket signaling server (e.g., 8080).
        *   `StunServers` ([]string): A list of STUN server URLs (e.g., "stun:stun.l.google.com:19302").
    *   `LoadConfig()` function: Returns a default `WebRTCConfig`.

*   **Signaling (`signal.go`):**
    *   Uses `github.com/gorilla/websocket` to upgrade HTTP connections to WebSockets.
    *   `SignalMessage` struct (`{Type string, Payload string}`) is used for exchanging messages.
    *   **WebSocket Handler (`HandleWebSocketConnections`):**
        *   Establishes a WebSocket connection.
        *   Creates a new `webrtc.PeerConnection` using `github.com/pion/webrtc/v3` for each client.
        *   **Offer Handling**:
            1.  When a "offer" `SignalMessage` is received:
            2.  The offer SDP (JSON string in `Payload`) is unmarshalled.
            3.  A new `webrtc.PeerConnection` is created.
            4.  (Conceptual) Logs registration with `SessionManagerClient`.
            5.  `OnICECandidate` callback is set up: Sends any gathered ICE candidates as "candidate" `SignalMessage`s back to the client over WebSocket.
            6.  `OnTrack` callback is set up: Logs information about the received remote media track (kind, codec, SSRC, etc.). Actual media processing is a placeholder.
            7.  The client's offer is set as the remote description on the `PeerConnection`.
            8.  An answer SDP is created by the `PeerConnection`.
            9.  The answer is set as the local description.
            10. The answer SDP is marshalled to JSON and sent back to the client as an "answer" `SignalMessage`. ICE gathering is awaited before sending the answer (non-trickle ICE for answer).
        *   **ICE Candidate Handling**:
            1.  When a "candidate" `SignalMessage` is received:
            2.  The ICE candidate (JSON string in `Payload`) is unmarshalled.
            3.  The candidate is added to the `PeerConnection`.
        *   **Disconnection**:
            1.  If the WebSocket connection closes or an error occurs, the event is logged.
            2.  (Conceptual) Logs deregistration with `SessionManagerClient`.

*   **Gateway Initialization (`webrtc_gateway.go`):**
    *   `StartWebRTCGateway(cfg WebRTCConfig, smClient SessionManagerClient)`:
        *   Initializes the `webrtc.Configuration` with STUN servers from `cfg.StunServers`.
        *   Sets up an HTTP route (`/ws`) to the `HandleWebSocketConnections` function.
        *   Starts an HTTP server on the address and port specified in `cfg`.

*   **Session Management (Conceptual):**
    *   A `SessionManagerClient` interface is defined in `signal.go` (`RegisterSession`, `DeregisterSession`).
    *   A `DummyWebRTCSessionManagerClient` is used to log calls to these interface methods, simulating interaction with a session management component.

*   **Media Handling (Placeholder):**
    *   The `OnTrack` callback logs details of incoming media tracks.
    *   Actual processing, forwarding, or recording of media is not implemented. The gateway primarily sets up the communication channel.

## Interaction with Other System Components

1.  **Client Connection & Signaling**:
    *   A web or mobile client initiates a WebSocket connection to the WebRTC Gateway's `/ws` endpoint.
    *   Signaling messages (SDP offers/answers, ICE candidates) are exchanged between the client and the gateway to negotiate the WebRTC session.
2.  **Peer Connection Establishment**:
    *   The `WebRTCGatewayService` uses ICE (with STUN servers specified in its configuration) to facilitate NAT traversal and establish a media path with the client.
3.  **Media Flow**:
    *   Once the peer connection is established, encrypted RTP media (audio/video) flows between the client and the gateway.
    *   **Incoming Audio**: The gateway receives audio from the client (logged by `OnTrack`) and (Future) would forward it to the `StreamingDataManager` in the Real-Time Processing Engine.
    *   **Outgoing Audio**: (Future) Text responses from Dialogue Management would be synthesized by the `TextToSpeechService`. This audio would then be packetized by the WebRTC Gateway and sent to the client via the established peer connection.
4.  **Session Teardown**: When the call ends (WebSocket disconnects), (Conceptual) the session is logged as deregistered.

This service relies on `github.com/pion/webrtc/v3` for WebRTC functionalities and `github.com/gorilla/websocket` for WebSocket communication.
The `go.mod` and `go.sum` files in the `voice_gateway_layer` directory manage these dependencies.
