# Protocol Buffer Definitions for Real-Time Processing Engine

This directory contains the Protocol Buffer (`.proto`) files that define the data structures and service interfaces used across the Real-Time Processing Engine.

## `audio_stream.proto`

This file contains the core definitions for exchanging audio data and interacting with related gRPC services.

**Package:** `real_time_processing`

**Go Package:** `revovoiceai/voice_gateway_layer/internal/protos/real_time_processing`

### Enum: `AudioFormat`

Defines the different audio encoding formats that can be specified in an `AudioSegment`.

*   `AUDIO_FORMAT_UNSPECIFIED = 0`: Default, unspecified format. Should ideally not be used in active streams.
*   `PCMU = 1`: Pulse Code Modulation, G.711 µ-law. Common in telephony.
*   `PCMA = 2`: Pulse Code Modulation, G.711 A-law. Common in telephony.
*   `OPUS = 3`: Opus codec. Common in WebRTC and modern VoIP.

### Message: `AudioSegment`

Represents a segment or chunk of audio data being streamed.

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
*   **Implemented by:** `StreamingDataManager` (Python service)
*   **RPC: `IngestAudioSegment (AudioSegment) returns (IngestResponse)`**
    *   **Role:** Receives individual `AudioSegment` messages from clients (e.g., Voice Gateways).
    *   `AudioSegment`: The audio data being sent.
    *   `IngestResponse`: A response message indicating the outcome of the ingestion.
*   **Message: `IngestResponse`**
    *   `string session_id = 1;`: The session ID from the request.
    *   `uint32 sequence_number = 2;`: The sequence number from the request.
    *   `string status_message = 3;`: A message describing the result (e.g., "Segment received", "Segment received and forwarded to STT", or error details).

### Service: `SpeechToText`

A gRPC service for performing speech-to-text transcription on audio segments.
*   **Implemented by:** `SpeechToTextService` (Python service)
*   **RPC: `TranscribeAudioSegment (AudioSegment) returns (TranscriptionResponse)`**
    *   **Role:** Processes an `AudioSegment` and returns a transcription.
    *   `AudioSegment`: The audio data to be transcribed.
    *   `TranscriptionResponse`: The result of the transcription.
*   **Message: `TranscriptionResponse`**
    *   `string session_id = 1;`: The session ID from the request.
    *   `uint32 sequence_number = 2;`: The sequence number from the request.
    *   `string transcript = 3;`: The transcribed text. (Currently a placeholder).
    *   `bool is_final = 4;`: Indicates if this is the final transcript for the segment or a complete utterance.
    *   `float confidence = 5;`: The confidence score of the transcription (0.0 to 1.0).
