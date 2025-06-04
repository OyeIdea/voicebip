# Streaming Data Manager

## Purpose

The Streaming Data Manager (SDM) is a service within the RevoVoiceAI Real-Time Processing Engine. It acts as a central hub for receiving audio segments from various sources (like Voice Gateways) and forwarding them to other processing services, primarily the Speech-to-Text (STT) service.

Its key functions include:
- Receiving audio segments via its gRPC `StreamIngest` service.
- Forwarding these segments to the `SpeechToTextService` via a gRPC call.

## Components

*   `manager.py`: Contains the main logic for the gRPC `StreamIngestServicer`. This servicer handles incoming audio segments and calls the STT service.
*   `audio_stream_pb2.py`: Generated Protobuf Python code for message structures (from `audio_stream.proto`).
*   `audio_stream_pb2_grpc.py`: Generated Protobuf Python code for gRPC client and server stubs (from `audio_stream.proto`).
*   `requirements.txt`: Python package dependencies (`grpcio`, `grpcio-tools`, `protobuf`).
*   `config.py`: (Placeholder) Intended for service-specific configurations.
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: StreamIngest

The Streaming Data Manager exposes a gRPC service named `StreamIngest`.

*   **Service Definition:** `StreamIngest` (defined in `audio_stream.proto`)
*   **Port:** Listens on `0.0.0.0:50051`.
*   **RPC Method:** `IngestAudioSegment(AudioSegment) returns (IngestResponse)`
    *   **`AudioSegment`**: A message containing a chunk of audio data, its format, session ID, sequence number, and other metadata.
    *   **`IngestResponse`**: A message indicating the result of the ingestion, including session ID, sequence number, and a status message.
    *   **Behavior**: Upon receiving an `AudioSegment`, the `IngestAudioSegment` method in `StreamIngestServicer`:
        1.  Logs the reception of the segment.
        2.  Establishes a gRPC connection to the `SpeechToTextService` (typically running on `localhost:50052`).
        3.  Calls the `TranscribeAudioSegment` method of the `SpeechToTextService`, forwarding the received `AudioSegment`.
        4.  Logs the `TranscriptionResponse` received from the STT service.
        5.  Returns an `IngestResponse` to the original caller, indicating that the segment was received and forwarded (or if an error occurred during forwarding).

## Interaction with Other Services

1.  **Receives from:** Voice Gateway Layer services (SIP Gateway, WebRTC Gateway). These services act as gRPC clients to the SDM's `StreamIngest` service.
2.  **Calls:** `SpeechToTextService`. The SDM acts as a gRPC client to the `SpeechToText` service's `TranscribeAudioSegment` method.

### Example Flow (gRPC based):

1.  A Voice Gateway (SIP or WebRTC) captures an audio segment from a call.
2.  The Voice Gateway calls `StreamIngest.IngestAudioSegment` on the SDM (at port `50051`), sending the `AudioSegment`.
3.  The `StreamIngestServicer` in SDM's `manager.py` receives the segment.
4.  The servicer then calls `SpeechToText.TranscribeAudioSegment` on the `SpeechToTextService` (at port `50052`), forwarding the same `AudioSegment`.
5.  The `SpeechToTextService` (currently a placeholder) returns a `TranscriptionResponse`.
6.  The SDM's `StreamIngestServicer` receives this `TranscriptionResponse`, logs it, and then returns an `IngestResponse` to the Voice Gateway.

This gRPC-based pipeline allows for efficient, structured, and real-time streaming of audio data from capture to initial processing stages.
The older methods like `register_stream`, `get_stream_data` in the `StreamingDataManager` class are conceptual placeholders from an earlier design and are not part of the current gRPC flow.
