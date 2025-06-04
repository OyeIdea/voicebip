# Speech-to-Text (STT) Service

## Purpose

This service is a core component of the RevoVoiceAI Real-Time Processing Engine. It uses the **Deepgram SDK** to perform real-time speech-to-text transcription. It's designed to receive audio segments (potentially as part of a continuous stream for a given session), transcribe them, and then forward the resulting transcript to the Natural Language Understanding (NLU) Service for further processing.

## Components

*   `service.py`: Contains the gRPC `SpeechToTextServicer`. This servicer handles incoming `AudioSegment` messages, interfaces with the Deepgram SDK for streaming transcription, and calls the NLU service with the transcript.
*   `config.py`: Manages configuration, primarily the `DEEPGRAM_API_KEY`.
*   `audio_stream_pb2.py`, `audio_stream_pb2_grpc.py`: Generated Protobuf/gRPC code for audio streaming (shared with `StreamingDataManager`).
*   `nlu_service_pb2.py`, `nlu_service_pb2_grpc.py`: Generated Protobuf/gRPC client stubs for calling the NLU service.
*   `requirements.txt`: Python package dependencies, including `grpcio`, `protobuf`, `deepgram-sdk`, and `python-dotenv`.
*   `__init__.py`: Makes the directory a Python package.
*   `utils.py`: (Placeholder) For any utility functions.

## Configuration

The service requires a Deepgram API key to function.
*   **Environment Variable:** The API key must be set as an environment variable named `DEEPGRAM_API_KEY`.
*   **Local Development (`.env` file):** For local development, you can create a `.env` file in the `real_time_processing_engine/speech_to_text_service/` directory and add your API key there:
    ```env
    DEEPGRAM_API_KEY=your_actual_deepgram_api_key_here
    ```
    The `config.py` uses `python-dotenv` to load this file if present. **Do not commit the `.env` file to version control.**

## gRPC Service: SpeechToText

*   **Service Definition:** `SpeechToText` (defined in `real_time_processing_engine/protos/audio_stream.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50052`.
*   **RPC Method:** `TranscribeAudioSegment(AudioSegment) returns (TranscriptionResponse)`
    *   **`AudioSegment`**: An incoming message containing a chunk of audio data (`data`), its `audio_format`, `session_id`, `sequence_number`, and an `is_final` flag.
    *   **`TranscriptionResponse`**: A message containing the `transcript`, `confidence` score, `is_final` status (indicating if this is a final transcript from Deepgram for an utterance), and the `session_id`.
    *   **Behavior**:
        1.  Upon receiving an `AudioSegment`, the `TranscribeAudioSegment` method in `SpeechToTextServicer` initializes or retrieves an active Deepgram live transcription connection associated with the `session_id`.
        2.  The `audio_data` from the segment is sent to the Deepgram streaming connection. The `audio_format` from the request is used to configure the Deepgram `LiveOptions` (e.g., encoding, sample rate).
        3.  **Streaming & Final Results:** The service uses Deepgram's interim and final results.
            *   If `AudioSegment.is_final` is `true` (signaling the end of a client-side utterance or audio stream for that session), the service attempts to wait for a final transcript from Deepgram for the audio processed so far in that session.
            *   If `AudioSegment.is_final` is `false`, the service may return an interim transcript if one is immediately available from Deepgram, or an empty transcript if not. The current implementation primarily focuses on returning a transcript when `is_final` is true.
        4.  The Deepgram connection for a `session_id` is closed when a final transcript is processed after an `is_final=true` segment, or on server shutdown.
        5.  **NLU Forwarding:** The obtained transcript (whether interim, final, or an error/timeout message) is then sent in an `NLURequest` to the `NLUService` (at `localhost:50053`) for further processing. The NLU response is logged.
        6.  The `SpeechToTextServicer` returns a `TranscriptionResponse` to its original caller (e.g., `StreamingDataManager`), containing the transcript from Deepgram.

## Current Operational Mode: Simulated STT

This service has been temporarily modified to simulate Deepgram Speech-to-Text (STT) processing. Key characteristics of this simulated mode are:

*   **No External Deepgram Connection:** It does not currently connect to the external Deepgram API.
*   **No API Key Required:** Consequently, a `DEEPGRAM_API_KEY` is not required to run the service in this simulated mode.
*   **Dummy Transcript Generation:** It generates a dummy/placeholder transcript. This transcript is typically based on metadata from the input `AudioSegment`, such as the `session_id`, `sequence_number`, and `audio_format`.
*   **Purpose of Simulation:** This simulation is primarily to enable end-to-end pipeline testing and development of the RevoVoiceAI system without a dependency on live Deepgram services or active API keys. It allows developers to verify data flow and interactions between services.
*   **NLU Forwarding Active:** The logic for forwarding the transcript to the `NLUService` remains active. The simulated (dummy) transcript will be sent to the NLU service for further processing.

## Key Dependencies
*   `deepgram-sdk`: For interacting with the Deepgram API.
*   `python-dotenv`: For managing environment variables during local development.
*   `grpcio`, `protobuf`: For gRPC communication.

## Interaction with Other Services

1.  **Receives from:** `StreamingDataManager`. The SDM calls `SpeechToText.TranscribeAudioSegment`.
2.  **Interacts with:** Deepgram's external ASR service for transcription.
3.  **Calls:** `NLUService`. After obtaining a transcript, STT calls `NLUService.ProcessText`.

## Important Notes on Current Implementation
*   **Async Bridging:** The Deepgram SDK is asynchronous. The gRPC servicer methods are synchronous. The current implementation uses `asyncio.run_coroutine_threadsafe` and a dedicated asyncio event loop running in a separate thread to manage Deepgram's async operations.
*   **Audio Format Handling:** The service attempts to map `AudioFormat` enum values to Deepgram encoding options. Proper handling of Opus (which requires decoding before sending to Deepgram live streams) and other formats is critical and may require additional processing steps not yet implemented.
*   **Error Handling:** Basic error handling for Deepgram connection and timeouts is included. More comprehensive error management would be needed for a production system.
*   **Session Management:** The service manages Deepgram connections per `session_id`. Cleanup of these connections on server shutdown is handled via `atexit`.
