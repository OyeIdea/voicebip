# Voice Activity Detection (VAD) Service

## Purpose

The Voice Activity Detection (VAD) Service is a component of the RevoVoiceAI Real-Time Processing Engine. Its primary function is to analyze an audio stream and distinguish between periods of human speech and silence (or non-speech noise). This helps in optimizing subsequent audio processing tasks, such as Speech-to-Text (STT), by focusing only on relevant audio segments.

## Core Functions (Conceptual)

*   **Audio Segmentation**: Processes incoming audio chunks (or a continuous stream) to identify boundaries between speech and non-speech.
*   **Event Generation**: Can emit events or return segment information indicating the start and end times of detected speech/silence periods.

## Components

*   `service.py`: Contains the main `VADService` class. Its `detect_speech_segments()` method takes an audio chunk and returns a list of identified speech and silence segments.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as VAD model sensitivity, frame size, padding durations, or energy thresholds.
*   `__init__.py`: Makes the directory a Python package.

## Interaction in the Real-Time Processing Engine

The VAD Service can be used in several ways within the engine:

*   **Preprocessing for STT**:
    1.  The `StreamingDataManager` (or another orchestrator) receives an audio stream.
    2.  Audio chunks are passed to the `VADService`.
    3.  The `VADService` identifies speech segments.
    4.  Only these speech segments are then forwarded to the `SpeechToTextService` for transcription. This can save computational resources and reduce errors from transcribing noise.
*   **Silence Detection**: Can be used to detect prolonged silences, which might trigger specific actions in a dialogue system (e.g., prompting the user if they are still there).
*   **Speaker Diarization Preprocessing**: VAD is often a first step in speaker diarization (identifying who spoke when).

This service will eventually integrate with specific VAD libraries or models (e.g., WebRTC VAD, Silero VAD, or custom models).
