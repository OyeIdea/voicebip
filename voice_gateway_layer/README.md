# Voice Gateway Layer

## Purpose

The Voice Gateway Layer is the primary entry and exit point for all real-time voice communications within the RevoVoiceAI platform. It is responsible for interfacing with various external communication networks and protocols, translating them into a consistent internal representation for processing by the Real-Time Processing Engine and other upstream services.

This layer handles the complexities of different voice transport and signaling protocols, such as:
*   **SIP (Session Initiation Protocol)**: For integration with traditional telephony (PSTN) via SIP trunks, and direct VoIP connections.
*   **WebRTC (Web Real-Time Communication)**: For enabling voice communication directly within web browsers and mobile applications.
*   **Telephony Provider Interfaces**: Direct integrations with specific carrier APIs (e.g., Twilio, Vonage) if needed, though often these are also exposed via SIP.

## Services

The services within this layer manage the specifics of each protocol:

*   **`sip_gateway_service`**: Handles SIP/RTP protocols for telephony integration. See the [SIP Gateway Service README](./sip_gateway_service/README.md) for more details.
*   *(Other services like `webrtc_gateway_service` will be added here as they are structured.)*

## Key Responsibilities

*   **Protocol Handling**: Managing signaling and media for different voice protocols.
*   **Session Management**: Maintaining state for active calls or voice sessions.
*   **Media Bridging/Forwarding**: Receiving media streams (e.g., RTP) and forwarding them to the Real-Time Processing Engine, and vice-versa.
*   **Load Balancing**: (Potentially) Distributing incoming call load across multiple processing nodes if applicable at this layer.

This layer ensures that the rest of the RevoVoiceAI platform can focus on AI processing and business logic without needing to understand the intricacies of various voice communication technologies.
