# Speech-to-Text (STT) Service

## Purpose

This service is a core component of the RevoVoiceAI Real-Time Processing Engine. Its primary responsibility is to receive audio data (either as a complete stream or in chunks) and transcribe it into text in real-time.

## Components

*   `service.py`: Contains the main logic for the STT service, including the `SpeechToTextService` class with methods to handle audio input and perform transcription.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as model paths, language settings, or performance parameters.
*   `utils.py`: (Placeholder) Intended for any utility functions that support the service's operations.
*   `__init__.py`: Makes the directory a Python package.

## Usage (Conceptual)

The `SpeechToTextService` is designed to be instantiated and used by other components within the Real-Time Processing Engine, such as a Streaming Data Manager or a direct API gateway.

-   The `transcribe_audio_stream(audio_stream_uri, audio_format)` method can be used for processing complete audio files or streams.
-   The `process_audio_chunk(audio_chunk)` method is intended for scenarios where audio is processed in smaller, sequential chunks for lower latency transcription.

The service will eventually integrate with specific speech recognition models (e.g., Whisper, wav2vec2) to perform the actual transcription.
