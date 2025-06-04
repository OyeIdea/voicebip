# Natural Language Understanding (NLU) Service

## Purpose

The Natural Language Understanding (NLU) Service is a component of the RevoVoiceAI AI/ML Services Layer. Its primary function is to process textual input (typically transcribed speech from the Speech-to-Text service) to extract meaning and structure, such as intents and entities. This structured information is then passed to the Dialogue Management (DM) service to inform responses and actions.

## Components

*   `service.py`: Contains the gRPC `NLUServiceServicer`. This servicer implements the placeholder NLU logic and then calls the Dialogue Management service.
*   `nlu_service_pb2.py`, `nlu_service_pb2_grpc.py`: Generated Protobuf/gRPC code for this NLU service.
*   `dialogue_management_service_pb2.py`, `dialogue_management_service_pb2_grpc.py`: Generated Protobuf/gRPC client stubs for calling the Dialogue Management service.
*   `requirements.txt`: Python package dependencies (`grpcio`, `grpcio-tools`, `protobuf`).
*   `config.py`: (Placeholder) For service-specific configurations.
*   `models/`: (Placeholder directory) For NLU model files.
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: NLUService

The NLU Service exposes its functionality via a gRPC interface.

*   **Service Name:** `NLUService` (defined in `ai_ml_services/protos/nlu_service.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50053`.
*   **RPC Method:** `ProcessText(NLURequest) returns (NLUResponse)`
    *   **`NLURequest`**: Contains the input `text` and `session_id`.
    *   **`NLUResponse`**: Returns the `session_id`, detected `intent`, `entities`, `processed_text`, and `intent_confidence`.
    *   **Behavior**:
        1.  Receives an `NLURequest` (typically from `SpeechToTextService`).
        2.  Applies placeholder NLU logic (keyword-based intent/entity detection for "greeting", "get_help", "get_weather").
        3.  Generates an `NLUResponse` with these placeholder results.
        4.  **Calls Dialogue Management Service**: It then creates a `DialogueRequest` (embedding the `NLUResponse` it just generated) and calls the `ManageTurn` RPC of the `DialogueManagementService` (running on port `50054`).
        5.  The response from `DialogueManagementService` is logged.
        6.  Finally, the `NLUService` returns its *own* generated `NLUResponse` to its original caller (e.g., `SpeechToTextService`).

## Current Status & Logic

The NLU processing logic in `service.py` is currently a **placeholder**. It uses simple keyword matching:
*   Defaults to "greeting" intent.
*   Detects "get_help" if "help" is in the text.
*   Detects "get_weather" if "weather" is in the text, and can extract "tomorrow" as a date entity.
The actual NLU models for robust intent recognition and entity extraction will be integrated in future development.

## Interaction in the System

1.  **Receives from:** `SpeechToTextService`. The STT service calls `NLUService.ProcessText` with the transcribed text.
2.  **Processing**:
    *   The `NLUServiceServicer` applies its placeholder NLU logic.
    *   It then calls `DialogueManagementService.ManageTurn` with the NLU results.
3.  **Calls:** `DialogueManagementService`.
4.  **Output**: Returns its own `NLUResponse` to the `SpeechToTextService`. The result from the `DialogueManagementService` is currently only logged by NLU.
