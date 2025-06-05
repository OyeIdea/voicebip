# Testing Guide

This guide provides instructions for testing specific components and flows within the RevoVoiceAI platform, particularly those involving external services or complex interactions.

## Testing SpeechToTextService with Live Deepgram API

This section outlines how to test the `SpeechToTextService` with a real Deepgram API key to ensure it can connect and perform live transcription.

### Prerequisites

1.  **Deepgram API Key**: You must have a valid Deepgram API Key.
2.  **Environment Variable**: The `SpeechToTextService` expects the API key to be available in an environment variable named `DEEPGRAM_API_KEY`. Export this variable in your shell before running the service:
    ```bash
    export DEEPGRAM_API_KEY="your_actual_deepgram_api_key_here"
    ```
    Ensure this key is active and has transcription minutes available on your Deepgram account.
    Alternatively, for local development, you can place it in a `.env` file within the `real_time_processing_engine/speech_to_text_service/` directory (ensure `.env` is in `.gitignore`).

### Running the Pipeline for STT Test

To test the STT service, you'll need to run the following services:

1.  **`SpeechToTextService`**:
    ```bash
    python real_time_processing_engine/speech_to_text_service/service.py
    ```
    This service listens on `localhost:50052`.

2.  **`StreamingDataManager`**:
    ```bash
    python real_time_processing_engine/streaming_data_manager/manager.py
    ```
    This service listens on `localhost:50051` and forwards requests to the STT service.

3.  **A Voice Gateway**: You'll need one of the voice gateways to send audio to the `StreamingDataManager`.

    *   **Using `sip_gateway` (Sends PCMU Audio)**:
        *   This gateway simulates sending PCMU audio after a mock call setup.
        *   Run it (ensure Go environment is set up):
            ```bash
            # From the voice_gateway_layer directory
            go run ./sip_gateway/ # Or however you build/run your Go services
            ```
        *   To trigger the `sip_gateway` to send audio, you would typically use a SIP client (e.g., Linphone, MicroSIP, or a SIP testing tool) to place a call to `sip:test@<sip_gateway_ip>:<sip_gateway_port>` (default port is usually 5060 if not specified otherwise in its config). Once the call is answered (200 OK by the gateway), the gateway will start sending its dummy PCMU audio stream.

    *   **Using `webrtc_gateway` (Sends LINEAR16 PCM Audio - Conceptually Decoded from Opus)**:
        *   This gateway forwards audio received from a WebRTC client. The code has placeholders for Opus decoding and now marks the audio sent to `StreamingDataManager` as `LINEAR16` (assumed to be 16kHz).
        *   Run it:
            ```bash
            # From the voice_gateway_layer directory
            go run ./webrtc_gateway/ # Or however you build/run your Go services
            ```
        *   To test, you'd need a WebRTC client (e.g., a browser application) to connect to its WebSocket endpoint (`ws://localhost:8080/ws` by default), establish a peer connection, and stream audio.
        *   Note that for this path to work with the real Deepgram STT service, the conceptual Opus decoding (planned in `signal.go`) must be fully implemented to provide actual Linear16 PCM data.

### Expected Behavior and Logs

When the pipeline is running and audio is sent from a gateway:

1.  **Gateway Logs**:
    *   `sip_gateway` or `webrtc_gateway` should log that it's sending `AudioSegment`s to the `StreamingDataManager` and should log the `IngestResponse` status.

2.  **`StreamingDataManager` Logs**:
    *   `StreamingDataManager: Received AudioSegment...` (detailing session, sequence, format like PCMU or LINEAR16).
    *   `StreamingDataManager: Received transcription from STT...` (showing the transcript received back).

3.  **`SpeechToTextService` Logs (Crucial for Deepgram Test)**:
    *   **If `DEEPGRAM_API_KEY` is missing or invalid**: You will likely see a `CRITICAL: DEEPGRAM_API_KEY is not set...` message on startup, or errors like `Deepgram error for session_id: Exception('auth')` or `Error starting Deepgram connection...` with authentication-related messages when an audio segment arrives.
    *   `Attempting to start Deepgram connection for session_id with format AUDIO_FORMAT_PCMU` (or `AUDIO_FORMAT_LINEAR16`).
    *   `Deepgram connection started for session_id with encoding mulaw, sample rate 8000` (for PCMU) or `encoding linear16, sample rate 16000` (for LINEAR16).
    *   `Deepgram final transcript for session_id: '...' (Confidence: ...)` - **This indicates successful transcription by Deepgram.**
    *   `Deepgram interim transcript for session_id: '...'` (if interim results are enabled and occur).
    *   `SpeechToTextService: Calling NLUService...` with the real transcript.

### Audio Format Notes:

*   The `sip_gateway` sends audio marked as `PCMU`. The `SpeechToTextService` configures Deepgram for "mulaw" at 8000 Hz for this.
*   The `webrtc_gateway` now sends audio marked as `LINEAR16` (conceptually decoded from Opus, assumed to be 16kHz PCM). The `SpeechToTextService` configures Deepgram for "linear16" at 16000 Hz. Ensure your Deepgram plan/model supports these encodings and sample rates.

This test will verify that the `SpeechToTextService` can correctly authenticate with Deepgram, open a live transcription stream, send audio data, and receive transcription results.
