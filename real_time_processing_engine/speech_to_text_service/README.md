# Speech-to-Text (STT) Service

## Purpose

This service is a core component of the RevoVoiceAI Real-Time Processing Engine. Its primary responsibility is to receive audio data (primarily via gRPC as audio segments), transcribe it into text, and then forward this text to the Natural Language Understanding (NLU) Service for further processing.

## Components

*   `service.py`: Contains the main logic for the gRPC `SpeechToTextServicer`. This servicer handles incoming audio segments, generates a placeholder transcript, and then calls the NLU service.
*   `audio_stream_pb2.py`, `audio_stream_pb2_grpc.py`: Generated Protobuf/gRPC code for audio streaming (shared).
*   `nlu_service_pb2.py`, `nlu_service_pb2_grpc.py`: Generated Protobuf/gRPC code for NLU service interaction (client stubs for NLUService).
*   `requirements.txt`: Python package dependencies.
*   `config.py`: (Placeholder) Intended for service-specific configurations.
*   `utils.py`: (Placeholder) Intended for any utility functions.
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: SpeechToText

The Speech-to-Text service exposes a gRPC service for receiving audio segments.

*   **Service Definition:** `SpeechToText` (defined in `real_time_processing_engine/protos/audio_stream.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50052`.
*   **RPC Method:** `TranscribeAudioSegment(AudioSegment) returns (TranscriptionResponse)`
    *   **`AudioSegment`**: An incoming message containing a chunk of audio data, its format, session ID, etc.
    *   **`TranscriptionResponse`**: A message containing the (currently placeholder) transcription for the segment, session ID, sequence number, etc.
    *   **Behavior**:
        1.  Upon receiving an `AudioSegment`, the `TranscribeAudioSegment` method logs its reception.
        2.  It generates a **placeholder transcript** (e.g., "Placeholder transcript for segment X..."). Actual STT model integration is a future step.
        3.  It then acts as a gRPC client to the `NLUService`. It creates an `NLURequest` containing the placeholder transcript and the original `session_id`.
        4.  It calls the `ProcessText` method of the `NLUService` (typically running on `localhost:50053`).
        5.  The response from `NLUService` (containing intent and entities) is logged.
        6.  Finally, it returns the placeholder `TranscriptionResponse` to its original caller (e.g., `StreamingDataManager`).

## Interaction with Other Services

1.  **Receives from:** `StreamingDataManager`. The SDM calls the STT service's `TranscribeAudioSegment` method.
2.  **Calls:** `NLUService`. After generating a transcript, the STT service calls the `NLUService`'s `ProcessText` method.

### Example Flow:

1.  The `StreamingDataManager` (SDM) receives an `AudioSegment` and calls `SpeechToText.TranscribeAudioSegment` on this STT service (at port `50052`).
2.  The `SpeechToTextServicer` in `service.py` receives the segment.
3.  It generates a placeholder transcript (e.g., "Placeholder transcript for segment 1 of session abc.").
4.  It then sends this transcript in an `NLURequest` to `NLUService.ProcessText` (at port `50053`).
5.  The `NLUService` (currently a placeholder) processes the text and returns an `NLUResponse` (e.g., intent="greeting").
6.  The `SpeechToTextServicer` logs this `NLUResponse`.
7.  The `SpeechToTextServicer` returns its original placeholder `TranscriptionResponse` to the SDM.

The older `SpeechToTextService` class (business logic) is currently not directly integrated into the gRPC servicer's placeholder flow but can be used for future real STT model integration.
The NLU results are not yet incorporated into the `TranscriptionResponse` sent back by STT; this could be a future enhancement or handled by a downstream aggregator service.
