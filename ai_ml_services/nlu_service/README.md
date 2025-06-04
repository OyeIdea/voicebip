# Natural Language Understanding (NLU) Service

## Purpose

The Natural Language Understanding (NLU) Service is a component of the RevoVoiceAI AI/ML Services Layer. Its primary function is to process textual input (typically transcribed speech from the Speech-to-Text service) to extract meaning and structure, such as intents and entities. This structured information can then be used by other services (e.g., Dialogue Management) to inform responses and actions.

## Components

*   `service.py`: Contains the gRPC `NLUServiceServicer` which implements the NLU logic.
*   `nlu_service_pb2.py`: Generated Protobuf Python code for NLU message structures.
*   `nlu_service_pb2_grpc.py`: Generated Protobuf Python code for NLU gRPC client and server stubs.
*   `requirements.txt`: Python package dependencies (`grpcio`, `grpcio-tools`, `protobuf`).
*   `config.py`: (Placeholder) Intended for service-specific configurations.
*   `models/`: (Placeholder directory) Intended to store NLU model files.
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: NLUService

The NLU Service exposes its functionality via a gRPC interface.

*   **Service Name:** `NLUService` (defined in `ai_ml_services/protos/nlu_service.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50053`.
*   **RPC Method:** `ProcessText(NLURequest) returns (NLUResponse)`
    *   **`NLURequest`**: Contains the input `text` to be processed and a `session_id` for context.
    *   **`NLUResponse`**: Returns the `session_id`, detected `intent`, a list of `entities` (with name, value, confidence), the `processed_text`, and an `intent_confidence` score.

## Current Status & Logic

The current implementation of the `NLUServiceServicer` in `service.py` provides **placeholder NLU logic**. It does not use any sophisticated NLU models yet.
The logic is primarily rule-based for demonstration:
*   It defaults to a "greeting" intent with a confidence of 0.95 and a default "User" entity.
*   If the input text contains "help" (case-insensitive), the intent is changed to "get_help", and a "topic:general" entity is added.
*   If the input text contains "weather" (case-insensitive), the intent is changed to "get_weather", and entities like "query_type:general_weather" and "date:tomorrow" (if "tomorrow" is present) are added.
*   The `session_id` from the request is passed through to the response.
*   The `processed_text` in the response is currently the same as the input text.

This placeholder implementation allows for end-to-end testing of the gRPC communication flow while the actual NLU models and logic are developed.

## Interaction in the System

1.  **Input**: Typically receives `NLURequest` from the `SpeechToTextService`, where the `text` field contains the transcript.
2.  **Processing**: The `NLUServiceServicer.ProcessText()` method applies its current (placeholder) logic.
3.  **Output**: Returns an `NLUResponse` to the caller (e.g., `SpeechToTextService`).
