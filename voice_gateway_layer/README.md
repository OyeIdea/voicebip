# Voice Gateway Layer

## Purpose

The Voice Gateway Layer is the primary entry and exit point for all real-time voice communications within the RevoVoiceAI platform. It is responsible for interfacing with various external communication networks and protocols, translating them into a consistent internal representation for processing by the Real-Time Processing Engine and other upstream services.

This layer handles the complexities of different voice transport and signaling protocols. Go is often chosen for services in this layer due to its strong performance in high-concurrency network applications.

Supported protocols include:
*   **SIP (Session Initiation Protocol)**: For integration with traditional telephony (PSTN) via SIP trunks, and direct VoIP connections.
*   **WebRTC (Web Real-Time Communication)**: For enabling voice communication directly within web browsers and mobile applications.
*   **Telephony Provider Interfaces**: (Future) Direct integrations with specific carrier APIs.

## Services

The services within this layer manage the specifics of each protocol:

*   **`sip_gateway` (Go Service)**: Handles SIP/RTP protocols for telephony integration. Implemented in Go for performance. See the [SIP Gateway Service README](./sip_gateway/README.md) for more details.
*   **`webrtc_gateway` (Go Service)**: Enables real-time voice (and potentially video) communication directly with web browsers and mobile applications using WebRTC standards. Implemented in Go. See the [WebRTC Gateway Service README](./webrtc_gateway/README.md) for more details.
*   *(Other gateway services might be added here as they are structured.)*

## Key Responsibilities

*   **Protocol Handling**: Managing signaling and media for different voice protocols.
*   **Session Management**: Maintaining state for active calls or voice sessions.
*   **Media Bridging/Forwarding**: Receiving media streams (e.g., RTP) and forwarding them to the Real-Time Processing Engine, and vice-versa.
*   **Load Balancing**: (Potentially) Distributing incoming call load across multiple processing nodes if applicable at this layer.

This layer ensures that the rest of the RevoVoiceAI platform can focus on AI processing and business logic without needing to understand the intricacies of various voice communication technologies.
