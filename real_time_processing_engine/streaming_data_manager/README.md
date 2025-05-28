# Streaming Data Manager

## Purpose

The Streaming Data Manager is a service within the RevoVoiceAI Real-Time Processing Engine. It is responsible for managing real-time data streams, primarily audio, that need to be processed by other services within the engine, such as the Speech-to-Text (STT) service.

Its key functions include:
- Registering and unregistering data streams from various sources.
- Fetching data from these streams.
- Routing stream data to appropriate processing services.

## Components

*   `manager.py`: Contains the main logic for the `StreamingDataManager` class, including methods to manage and route streams.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as stream buffer sizes, connection timeouts, or supported stream protocols.
*   `__init__.py`: Makes the directory a Python package.

## Interaction with Other Services

A primary role of the Streaming Data Manager is to act as an intermediary between raw data sources and processing services.

### Example Flow (with Speech-to-Text Service):

1.  An external system (e.g., Voice Gateway) notifies the Streaming Data Manager about a new audio stream (e.g., a live call).
2.  The `StreamingDataManager` registers this stream using `register_stream(stream_id, source_info)`.
3.  To begin transcription, the `route_stream_to_stt(stream_id)` method is called.
4.  The `StreamingDataManager` would then continuously:
    a.  Fetch audio data from the source using `get_stream_data(stream_id, chunk_size)`.
    b.  Send the fetched audio chunks to the `SpeechToTextService` (e.g., by calling its `process_audio_chunk()` method).
5.  Once the stream ends or is no longer needed, it is unregistered using `unregister_stream(stream_id)`.

This decouples the STT service (and other future AI services) from the specifics of how individual data streams are accessed and managed.
