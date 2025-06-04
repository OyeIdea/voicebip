# Protocol Buffer Definitions for Real-Time Processing Engine

This directory contains the Protocol Buffer (`.proto`) files that define the data structures and service interfaces used across the Real-Time Processing Engine.

## `audio_stream.proto`

This file contains the core definitions for exchanging audio data and interacting with related gRPC services within the real-time audio processing pipeline.

**Package:** `real_time_processing`

**Go Package:** `revovoiceai/voice_gateway_layer/internal/protos/real_time_processing` (Used by Voice Gateway Go components)

### Enum: `AudioFormat`

Defines the different audio encoding formats that can be specified in an `AudioSegment`.

*   `AUDIO_FORMAT_UNSPECIFIED = 0`: Default, unspecified format. Should ideally not be used in active streams.
*   `PCMU = 1`: Pulse Code Modulation, G.711 µ-law. Common in telephony.
*   `PCMA = 2`: Pulse Code Modulation, G.711 A-law. Common in telephony.
*   `OPUS = 3`: Opus codec. Common in WebRTC and modern VoIP.

### Message: `AudioSegment`

Represents a segment or chunk of audio data being streamed. This message is fundamental to the audio processing pipeline.

*   `string session_id = 1;`
    *   A unique identifier for the overall audio session or call.
*   `int64 timestamp = 2;`
    *   Unix timestamp in milliseconds indicating when the segment was captured or generated.
*   `AudioFormat audio_format = 3;`
    *   The encoding format of the audio data in this segment (see `AudioFormat` enum).
*   `uint32 sequence_number = 4;`
    *   A sequential number for segments within the same `session_id`, used for ordering.
*   `bytes data = 5;`
    *   The raw audio data payload.
*   `bool is_final = 6;`
    *   A boolean flag indicating if this is the final segment for a given utterance or logical block of speech. Useful for VAD (Voice Activity Detection) or signaling the end of a complete thought.

### Service: `StreamIngest`

A gRPC service for ingesting audio segments into the system.
*   **Implemented by:** `StreamingDataManager` (Python service in `real_time_processing_engine`).
*   **RPC: `IngestAudioSegment (AudioSegment) returns (IngestResponse)`**
    *   **Role:** Receives individual `AudioSegment` messages from clients (e.g., Voice Gateways). It then forwards these segments to the `SpeechToText` service.
    *   `AudioSegment`: The audio data being sent.
    *   `IngestResponse`: A response message indicating the outcome of the ingestion.
*   **Message: `IngestResponse`**
    *   `string session_id = 1;`: The session ID from the request.
    *   `uint32 sequence_number = 2;`: The sequence number from the request.
    *   `string status_message = 3;`: A message describing the result (e.g., "Segment received and forwarded to STT").

### Service: `SpeechToText`

A gRPC service for performing speech-to-text transcription on audio segments.
*   **Implemented by:** `SpeechToTextService` (Python service in `real_time_processing_engine`).
*   **RPC: `TranscribeAudioSegment (AudioSegment) returns (TranscriptionResponse)`**
    *   **Role:** Processes an `AudioSegment`, generates a (currently placeholder) transcript, and then calls the `NLUService` with this transcript.
    *   `AudioSegment`: The audio data to be transcribed.
    *   `TranscriptionResponse`: The result of the transcription (currently a placeholder).
*   **Message: `TranscriptionResponse`**
    *   `string session_id = 1;`: The session ID from the request.
    *   `uint32 sequence_number = 2;`: The sequence number from the request.
    *   `string transcript = 3;`: The transcribed text.
    *   `bool is_final = 4;`: Indicates if this is the final transcript for the segment or a complete utterance.
    *   `float confidence = 5;`: The confidence score of the transcription (0.0 to 1.0).

---

## `tts_service.proto`

This file defines the interface for the Text-to-Speech (TTS) Service.

**Package:** `real_time_processing_engine.tts`

**Go Package:** `revovoiceai/real_time_processing_engine/protos/tts_service` (For potential Go clients/servers)

### Message: `TTSRequest`

Represents a request to the TTS service to synthesize speech from text.

*   `string text_to_synthesize = 1;`
    *   The text that needs to be converted into audible speech.
*   `string session_id = 2;`
    *   A unique identifier for the session, used for context and logging.
*   `string voice_config_id = 3;`
    *   An optional identifier for a specific voice configuration or model to be used for synthesis (e.g., "male_us_english_1", "female_uk_english_premium").

### Message: `TTSResponse`

Represents the response from the TTS service after processing a synthesis request. For the current placeholder implementation, this primarily contains status information.

*   `string session_id = 1;`
    *   The session identifier, passed through from the `TTSRequest`.
*   `string status_message = 2;`
    *   A message describing the result or status of the synthesis request (e.g., "Synthesis started", "Text received by TTS. Placeholder synthesis.").
*   `// string audio_chunk_id = 3;`
    *   (Commented out) Optional: Identifier for the first audio chunk if the audio is streamed.
*   `// bytes audio_data = 4;`
    *   (Commented out) Optional: For returning small audio data directly in the response.
*   `// string error_message = 5;`
    *   (Commented out) Optional: If an error occurred during synthesis.

### Service: `TextToSpeechService`

Defines the gRPC service contract for text-to-speech synthesis.

*   **RPC: `SynthesizeText (TTSRequest) returns (TTSResponse)`**
    *   **Role:** This is the primary method for the TTS service. It takes a `TTSRequest` (containing the text to synthesize), and is intended to produce synthesized speech. Currently, it's a placeholder that acknowledges the request.
    *   This service is implemented by the Python `TextToSpeechServicer` in the `real_time_processing_engine/text_to_speech_service` directory.
