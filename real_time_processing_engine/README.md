# RevoVoiceAI - Real-Time Processing Engine

This directory contains the services that form the core of the Real-Time Processing Engine for RevoVoiceAI. These services are responsible for handling live audio input, performing initial processing, and orchestrating interactions with the AI/ML layer.

## Services

Currently, the following services are outlined:

*   **`speech_to_text_service`**: Responsible for transcribing audio streams into text.
*   **`streaming_data_manager`**: Manages incoming data streams (e.g., audio from calls) and routes them to appropriate processing services like the Speech-to-Text service.

## Conceptual Data Flow: Audio Transcription

A typical flow for transcribing an audio stream (e.g., from a phone call) would be:

1.  **Stream Initiation**: An external component (like a Voice Gateway, not yet implemented here) signals the start of a new audio stream.
2.  **Stream Registration**: This external component, or a higher-level orchestrator, instructs the `StreamingDataManager` to register the new audio stream. The `StreamingDataManager` now knows how to access this audio (e.g., via a URI or by receiving chunks directly).
3.  **Routing to STT**: The `StreamingDataManager` is instructed to route this stream to the `SpeechToTextService`.
4.  **Data Fetching & Forwarding**:
    *   The `StreamingDataManager` continuously fetches chunks of audio data from the registered stream source.
    *   It forwards these audio chunks to the `SpeechToTextService`'s `process_audio_chunk()` method (or uses `transcribe_audio_stream()` if it's a complete file).
5.  **Transcription**: The `SpeechToTextService` processes the audio chunks and produces transcribed text.
6.  **Output**: The transcribed text is then made available, typically passed on to other services like Natural Language Understanding (NLU) or Dialogue Management for further processing. (The exact mechanism for this output is TBD).

This modular design allows for flexibility:
- The `SpeechToTextService` focuses solely on transcription and doesn't need to know the origin or specifics of the audio stream.
- The `StreamingDataManager` handles the complexities of stream management and can be extended to support various stream sources or types.
