# WebRTC Gateway Service (Go Implementation)

## Purpose

The WebRTC Gateway Service, implemented in Go, is a component of the RevoVoiceAI Voice Gateway Layer. It enables real-time voice (and potentially video) communication directly with web browsers and mobile applications using WebRTC (Web Real-Time Communication) standards. This allows for rich, interactive voice experiences without requiring users to install separate telephony software.

Go is chosen for its suitability in building high-performance, concurrent network services.

## Core Functions (Conceptual)

*   **Signaling Server**: Manages the exchange of session control messages (like SDP offers/answers and ICE candidates) between WebRTC peers. This is often implemented using WebSockets.
*   **ICE/STUN/TURN Handling**: Facilitates NAT (Network Address Translation) traversal to establish direct peer-to-peer media connections where possible, using STUN for NAT discovery and TURN for relaying media if direct connection fails.
*   **Peer Connection Management**: Establishes, monitors, and tears down WebRTC peer connections.
*   **Data Channel Management**: (Optional) Can manage WebRTC data channels for sending non-media data alongside audio/video.
*   **Media Stream Handling**: Receives incoming RTP media streams from WebRTC clients and forwards them to the Real-Time Processing Engine (e.g., `StreamingDataManager`), and vice-versa for outgoing audio.

## Components

*   `webrtc_gateway.go`: Contains the main `WebRTCGatewayService` struct and its methods for initializing the signaling server, managing WebRTC peer connections, handling ICE events, and processing signaling messages.
*   `config.go`: (Placeholder) Intended for service-specific configurations, such as signaling server address/port, STUN/TURN server URIs and credentials, and supported codecs.
*   `handlers/`: (Placeholder directory) Intended for Go files containing more detailed logic for handling WebSocket connections or specific signaling protocol messages.
*   *(Other Go package files might be added as development progresses.)*

## Interaction with Other System Components

1.  **Client Connection & Signaling**:
    *   A web or mobile client initiates a connection to the WebRTC Gateway's signaling endpoint (e.g., a WebSocket server).
    *   Signaling messages (SDP offers/answers, ICE candidates) are exchanged between the client and the gateway to negotiate the WebRTC session.
2.  **Peer Connection Establishment**:
    *   The `WebRTCGatewayService` uses ICE (with STUN/TURN servers specified in its configuration) to establish a media path with the client.
3.  **Media Flow**:
    *   Once the peer connection is established, encrypted RTP media (audio/video) flows between the client and the gateway.
    *   **Incoming Audio**: The gateway receives audio from the client and forwards it to the `StreamingDataManager` in the Real-Time Processing Engine for STT, NLU, etc.
    *   **Outgoing Audio**: Text responses from Dialogue Management are synthesized by the `TextToSpeechService`. This audio is then packetized by the WebRTC Gateway and sent to the client via the established peer connection.
4.  **Session Teardown**: When the call ends, signaling messages are exchanged to close the peer connection and related resources.

This service will integrate with Go-based WebRTC libraries (e.g., `pion/webrtc`) to handle the low-level details of the WebRTC protocol stack.
