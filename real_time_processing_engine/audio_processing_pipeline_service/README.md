# Audio Processing Pipeline Service

## Purpose

The Audio Processing Pipeline Service is a component of the RevoVoiceAI Real-Time Processing Engine. It is designed to enhance the quality of audio streams by applying a configurable sequence of audio processing operations. This can include tasks like noise reduction, echo cancellation, automatic gain control, and audio format conversion.

Improving audio quality is crucial for both incoming audio (to improve accuracy of Speech-to-Text) and outgoing audio (to provide a clearer experience for the user).

## Core Functions (Conceptual)

*   **Noise Reduction**: Attenuates background noise from the audio signal.
*   **Echo Cancellation**: Removes echo or feedback, especially relevant in telephony or speakerphone scenarios.
*   **Format Conversion**: Converts audio between different codecs, sample rates, or bit depths (e.g., PCM WAV to G.711 mu-law).
*   **Automatic Gain Control (AGC)**: Adjusts audio levels to maintain a consistent volume.
*   **Customizable Pipeline**: Allows for a flexible sequence of these operations to be applied based on specific needs.

## Components

*   `pipeline.py`: Contains the main `AudioProcessingPipelineService` class. Its `process_audio()` method takes an audio chunk and a list of desired operations, applying them sequentially.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as parameters for noise reduction algorithms, echo canceller settings, or preferred audio formats.
*   `__init__.py`: Makes the directory a Python package.
*   *(Subdirectories for specific modules like `noise_reduction/`, `echo_cancellation/` might be added later to house different algorithm implementations.)*

## Interaction in the Real-Time Processing Engine

The Audio Processing Pipeline Service can be invoked at various points:

*   **Input Audio Path**:
    1.  Audio is received by the `StreamingDataManager`.
    2.  Chunks can be passed to the `AudioProcessingPipelineService` for cleanup (e.g., noise reduction, echo cancellation) *before* or *after* Voice Activity Detection (VAD).
    3.  The cleaned audio is then sent to the `SpeechToTextService`.
*   **Output Audio Path**:
    1.  The `TextToSpeechService` generates synthesized audio.
    2.  This audio can be passed through the `AudioProcessingPipelineService` for final adjustments (e.g., format conversion to meet telephony requirements, volume normalization) before being sent to the user via the Voice Gateway.

The `process_audio()` method allows callers to specify which operations are needed for a particular audio stream, providing flexibility. This service will eventually integrate with audio processing libraries (e.g., WebRTC audio processing components, SpeexDSP, SoX, or custom algorithms).
