# SIP Gateway Service (Go Implementation)

## Purpose

The SIP Gateway Service, implemented in Go, is a core component of the RevoVoiceAI Voice Gateway Layer. It acts as the primary interface for integrating with traditional telephony systems and Voice over IP (VoIP) networks using the Session Initiation Protocol (SIP). Its main responsibilities include handling SIP signaling, managing call sessions, and facilitating Real-time Transport Protocol (RTP) media streams.

This Go implementation is chosen for its performance characteristics suitable for high-concurrency network applications.

## Core Functions (Conceptual)

*   **SIP Signaling**: Listens for and processes incoming SIP messages (INVITE, ACK, BYE, REGISTER, OPTIONS, etc.).
*   **Call Management**: Establishes, maintains, and terminates call sessions.
*   **RTP Session Management**: Sets up and tears down RTP sessions for audio media, including SDP negotiation.
*   **Registration**: Can act as a SIP registrar or register with upstream providers.
*   **Routing**: Determines how incoming calls are routed to internal services (e.g., Real-Time Processing Engine).

## Components

*   `sip_gateway.go`: Contains the main logic for the SIP gateway, including the UDP listener, message parsing, and handling of basic SIP methods (INVITE, ACK, BYE).
*   `config.go`: Defines the `SIPConfig` struct and `LoadConfig` function for service configuration. Currently supports `ListenAddress` and `ListenPort`.
*   `sip_messages.go`: Defines the `SIPRequest` struct for parsed SIP messages and constants for common SIP headers. It also includes the conceptual `SessionManagerClient` interface.
*   `sip_gateway_test.go`: Contains unit tests for parsing SIP messages and conceptual handling of INVITE and BYE methods.
*   `handlers/`: (Placeholder directory) Intended for Go files containing more detailed logic for handling specific SIP methods (e.g., `invite_handler.go`). *(This directory is not yet used in the current basic implementation).*

## Implemented Functionality

The current implementation provides a basic, functional SIP gateway capable of handling a simple call flow:

*   **Configuration (`config.go`):**
    *   `SIPConfig` struct:
        *   `ListenAddress` (string): The IP address for the gateway to listen on (e.g., "0.0.0.0").
        *   `ListenPort` (int): The UDP port for SIP signaling (e.g., 5060).
    *   `LoadConfig()` function: Returns a default `SIPConfig` with "0.0.0.0" and port 5060.

*   **SIP Message Handling (`sip_gateway.go`, `sip_messages.go`):**
    *   **Parsing**: `parseSIPRequest` can parse the request line (Method, Request-URI, SIP Version) and essential headers (`To`, `From`, `Call-ID`, `CSeq`, `Via`, `Max-Forwards`, `Content-Type`, `Content-Length`).
    *   **Response Generation**: `generateResponse` can create string representations for SIP responses, including "100 Trying", "180 Ringing", and "200 OK". For "200 OK" to an INVITE, a basic placeholder SDP body is included.
    *   **UDP Listener**: `StartSIPGateway` sets up a UDP listener on the configured address and port.

*   **Call Flow Handling (`sip_gateway.go`):**
    *   **INVITE**:
        1.  Logs the receipt of the INVITE message.
        2.  Sends a "100 Trying" response to the sender.
        3.  Logs a message indicating conceptual registration with the `SessionManagerClient` using the `Call-ID`.
        4.  Sends a "180 Ringing" response.
        5.  After a short simulated delay, sends a "200 OK" response, including a basic SDP.
        6.  Logs a placeholder message that an RTP stream should now be established.
    *   **ACK**:
        1.  Logs the receipt of the ACK message. (No response is sent for ACK in this context).
    *   **BYE**:
        1.  Logs the receipt of the BYE message.
        2.  Logs a message indicating conceptual deregistration with the `SessionManagerClient` using the `Call-ID`.
        3.  Sends a "200 OK" response to the sender.

*   **Session Management (Conceptual):**
    *   A `SessionManagerClient` interface is defined (`RegisterSession`, `DeregisterSession`).
    *   A `DummySessionManagerClient` is used to log calls to these interface methods, simulating interaction with a session management component.

*   **RTP Stream Handling (Placeholder):**
    *   The gateway logs messages indicating when RTP streams should be established or terminated.
    *   A basic SDP (Session Description Protocol) body is included in the "200 OK" response to an INVITE, which is necessary for RTP session negotiation. However, actual RTP stream setup, media handling, and SDP negotiation logic are not yet implemented.

## Interaction with Other System Components

The SIP Gateway Service is a critical entry and exit point for voice:

1.  **Incoming Call**:
    *   Receives SIP INVITE.
    *   Parses INVITE, negotiates SDP (currently basic placeholder), establishes RTP session (placeholder).
    *   Conceptually signals `session_manager` (from the same layer) to register the session.
    *   (Future) Would signal `StreamingDataManager` (Real-Time Processing Engine) about the new audio stream.
2.  **Outgoing Call / Response**:
    *   (Future) Would receive instructions/text from Dialogue Management.
    *   (Future) If TTS response, audio is generated by `TextToSpeechService` and streamed via RTP through the session managed by this SIP Gateway.
    *   Handles SIP signaling for call maintenance or termination (currently BYE).

This service will integrate with Go-based SIP libraries (e.g., `github.com/ghettovoice/gosip` or others) and potentially RTP libraries to handle the underlying protocol complexities in future, more advanced iterations. For now, it uses standard Go libraries for basic UDP networking and string manipulation.
