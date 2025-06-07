# Design for TTS Audio Playback to Gateways

## 1. Introduction

This document outlines potential designs for delivering synthesized audio from the `TextToSpeechService` (TTS) back to the originating voice gateway (`sip_gateway` or `webrtc_gateway`) for playback to the end-user. Currently, the `DialogueManagementService` (DM) receives a text response, calls the `TextToSpeechService` which returns a (simulated) audio payload, but this audio is not yet routed back to the user.

## 2. Requirements

*   The mechanism must route audio from TTS to the correct gateway instance and session that initiated the interaction.
*   It should handle different audio formats if TTS can produce them (e.g., PCMU for SIP, potentially Opus or L16 for WebRTC).
*   Consider latency, complexity, and scalability.

## 3. Proposed Design Options

### Option 1: Direct Gateway Call from Dialogue Management Service

*   **Description:**
    1.  `DialogueManagementService` (DM) receives the `TTSResponse` (containing audio data and format) from `TextToSpeechService`.
    2.  DM needs to determine the originating gateway type (SIP/WebRTC) and potentially its address/identifier for the specific `session_id`. This information might need to be:
        *   Passed along the call chain (Gateway -> SDM -> STT -> NLU -> DM).
        *   Queried from a `SessionManager` service using the `session_id`.
    3.  DM then makes a gRPC call to a new, specific RPC method on the identified gateway (e.g., `PlayAudio(AudioPlaybackRequest)`).
    4.  The gateway receives the audio and plays it back on the appropriate call leg/track.

*   **Pros:**
    *   Relatively direct flow for simple cases.
    *   Potentially lower latency if DM directly calls the gateway.

*   **Cons:**
    *   DM becomes coupled with gateway interfaces and needs logic to differentiate/locate gateways.
    *   Requires adding new gRPC services/methods to both gateways.
    *   Session-to-gateway mapping needs to be robustly managed.
    *   If gateways are scaled independently, DM needs a discovery mechanism or a load balancer for them.

*   **Conceptual Proto Snippets (for Gateways):**
    ```protobuf
    // In a new proto, e.g., gateway_playback_service.proto
    syntax = "proto3";

    package voice_gateway_layer; // Or specific gateway package

    import "real_time_processing_engine/protos/audio_stream.proto"; // For AudioFormat

    message AudioPlaybackRequest {
      string session_id = 1; // Call-ID for SIP, or WebRTC session ID
      bytes audio_data = 2;
      real_time_processing.AudioFormat audio_format = 3;
      // Add other params like voice_activity_info if needed for interrupting playback, etc.
    }

    message PlaybackResponse {
      string session_id = 1;
      bool status = 2; // True if playback initiated successfully
      string message = 3;
    }

    service GatewayPlayback {
      rpc PlayAudio(AudioPlaybackRequest) returns (PlaybackResponse);
    }
    ```

### Option 2: Message Queue (MQ) for Audio Delivery

*   **Description:**
    1.  `TextToSpeechService` (or DM, after receiving audio from TTS) publishes the `AudioPlaybackRequest` (containing session_id, audio_data, format) to a specific message queue topic (e.g., `tts.audio.output.<gateway_type>.<gateway_id>.<session_id>` or a more general `tts.audio.output.<session_id>`).
    2.  Each gateway instance, upon establishing a call/session, subscribes to relevant topics/messages for its active sessions.
    3.  When a gateway receives an audio message for an active session, it plays it back.

*   **Pros:**
    *   Decouples TTS/DM from the gateways. Gateways only need to know about the MQ.
    *   Good for scalability and resilience. MQ can handle buffering.
    *   Flexible: Other services could potentially consume these audio outputs for logging/analysis.

*   **Cons:**
    *   Introduces a new system component (Message Queue like RabbitMQ, Kafka, NATS) with its own operational overhead.
    *   Potentially higher latency due to MQ hops.
    *   Requires careful topic design and subscription management in gateways.

### Option 3: Centralized Control via Session Orchestrator Service

*   **Description:**
    1.  The `SessionOrchestratorService` (in the Business Logic Layer) manages the overall call flow.
    2.  Gateway informs Orchestrator of new call -> Orchestrator manages flow through SDM -> STT -> NLU.
    3.  NLU result goes to Orchestrator.
    4.  Orchestrator sends text to DM (if DM is just for dialogue logic/state) or directly to TTS.
    5.  TTS returns audio (or reference to it) to Orchestrator.
    6.  Orchestrator, knowing the session and originating gateway, sends the audio data/reference to the correct gateway via a specific gRPC call (similar to Option 1's gateway RPC, but called by Orchestrator).

*   **Pros:**
    *   Centralizes complex call flow logic.
    *   Services like DM, TTS remain focused on their core tasks.
    *   Clear separation of concerns. Orchestrator handles inter-service communication details for the call.

*   **Cons:**
    *   Orchestrator becomes a highly critical component; potential bottleneck if not designed well.
    *   Increases the number of hops for some interactions.
    *   Requires the `SessionOrchestratorService` to be fully implemented and aware of all relevant services.

## 4. Recommendation

For initial implementation and simplicity, **Option 1 (Direct Gateway Call from Dialogue Management Service)**, with a supporting role from the `SessionManagerService`, is likely the quickest to get a basic end-to-end audio output working.

**Refinement for Option 1:**
*   The `SessionManagerService` (which gateways already interact with) could be enhanced. When a gateway registers a session, it provides its type (SIP/WebRTC) and a gRPC callback address (or a known service name if using service discovery).
*   When DM needs to play audio, it queries the `SessionManagerService` for the `session_id` to get the gateway's gRPC address and type.
*   DM then uses the appropriate gRPC client to call `PlayAudio` on that gateway.

This refinement keeps DM somewhat decoupled from knowing *all* gateway instances, relying on the Session Manager as a directory.

Long-term, **Option 3 (Centralized Control via Session Orchestrator Service)** offers a cleaner architecture for complex scenarios, but requires more upfront work on the orchestrator itself. Option 2 (MQ) offers good decoupling but adds infrastructure.

**Initial focus should be on defining the `GatewayPlayback` gRPC service and implementing the client call from DM, potentially with SessionManager assistance for endpoint discovery.**
