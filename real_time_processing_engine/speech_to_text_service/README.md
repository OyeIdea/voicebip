# Speech-to-Text (STT) Service

## Purpose

This service is a core component of the RevoVoiceAI Real-Time Processing Engine. Its primary responsibility is to receive audio data (primarily via gRPC as audio segments) and transcribe it into text.

## Components

*   `service.py`: Contains the main logic for the `SpeechToTextService` class and the gRPC `SpeechToTextServicer`.
*   `audio_stream_pb2.py`: Generated Protobuf Python code for message structures (shared with other services).
*   `audio_stream_pb2_grpc.py`: Generated Protobuf Python code for gRPC client and server stubs (shared).
*   `requirements.txt`: Python package dependencies.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as model paths, language settings, or performance parameters.
*   `utils.py`: (Placeholder) Intended for any utility functions that support the service's operations.
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: SpeechToText

The Speech-to-Text service exposes a gRPC service for receiving audio segments and returning transcriptions.

*   **Service Definition:** `SpeechToText`
*   **RPC Method:** `TranscribeAudioSegment(AudioSegment) returns (TranscriptionResponse)`
    *   `AudioSegment`: A message containing a chunk of audio data, its format, session ID, sequence number, and other metadata. Defined in `real_time_processing_engine/protos/audio_stream.proto`.
    *   `TranscriptionResponse`: A message containing the transcription for the segment, session ID, sequence number, a flag indicating if the transcript is final, and a confidence score.
*   **Port:** The gRPC server listens on `0.0.0.0:50052`.

Currently, the `TranscribeAudioSegment` method receives audio segments, logs their reception, and returns a **placeholder** `TranscriptionResponse`. The actual transcription logic using STT models (e.g., Whisper, wav2vec2) is planned for future development.

## Interaction with Other Services

The `SpeechToTextService` is designed to be called by other services within the Real-Time Processing Engine, primarily the `StreamingDataManager`.

### Example Flow:

1.  The `StreamingDataManager` receives an `AudioSegment` via its own gRPC service.
2.  (Future Implementation) The `StreamingDataManager` then acts as a gRPC client to the `SpeechToTextService`.
3.  It calls the `TranscribeAudioSegment` method, sending the `AudioSegment`.
4.  The `SpeechToTextServicer` in `service.py` receives the segment.
5.  Currently, it generates a placeholder transcript (e.g., "Placeholder transcript...").
6.  The servicer returns a `TranscriptionResponse` to the `StreamingDataManager`.
7.  (Future) The `StreamingDataManager` might then forward this transcription to other services or make it available via another interface.

The older methods like `transcribe_audio_stream` and `process_audio_chunk` in the `SpeechToTextService` class are conceptual placeholders and may be integrated with or replaced by the gRPC service logic.
