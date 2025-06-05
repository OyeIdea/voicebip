# Dialogue Management (DM) Service

## Purpose

The Dialogue Management (DM) Service is a component of the RevoVoiceAI AI/ML Services Layer. It acts as the conversational "brain," taking the structured output from the Natural Language Understanding (NLU) Service, determining the system's next response or action, and then forwarding this textual response to the Text-to-Speech (TTS) Service for synthesis.

## Components

*   `service.py`: Contains the gRPC `DialogueManagementServicer`. This servicer implements placeholder dialogue logic and then calls the TTS service.
*   `dialogue_management_service_pb2.py`, `dialogue_management_service_pb2_grpc.py`: Generated Protobuf/gRPC code for this DM service.
*   `nlu_service_pb2.py`: Copied from NLU service; contains definitions for `NLUResponse`.
*   `tts_service_pb2.py`, `tts_service_pb2_grpc.py`: Generated Protobuf/gRPC client stubs for calling the Text-to-Speech service.
*   `requirements.txt`: Python package dependencies (`grpcio`, `grpcio-tools`, `protobuf`).
*   `config.py`: (Placeholder) For service-specific configurations.
*   `rules/`, `state_trackers/`: (Placeholder directories).
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: DialogueManagementService

The DM Service exposes its functionality via a gRPC interface.

*   **Service Name:** `DialogueManagementService` (defined in `ai_ml_services/protos/dialogue_management_service.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50054`.
*   **RPC Method:** `ManageTurn(DialogueRequest) returns (DialogueResponse)`
    *   **`DialogueRequest`**: Contains `session_id` and `nlu_result` (from NLU service).
    *   **`DialogueResponse`**: Returns `session_id` and the determined `text_response`.
    *   **Behavior**:
        1.  Receives a `DialogueRequest` from `NLUService`.
        2.  Applies placeholder dialogue logic based on the `nlu_result.intent` to formulate a `text_response`.
        3.  **Calls Text-to-Speech Service**: It then creates a `TTSRequest` containing this `text_response` and the `session_id`.
        4.  It calls the `SynthesizeText` RPC of the `TextToSpeechService` (running on port `50055`).
        5.  The response from `TextToSpeechService` (currently a status message) is logged.
        6.  Finally, the DM service returns its own `DialogueResponse` (containing the `text_response`) to its original caller (`NLUService`).

## Current Status & Logic

The dialogue logic in `service.py` is currently a **placeholder**, but has been enhanced with basic state management:
*   It uses a simple in-memory session state (per `session_id`) to handle basic multi-turn interactions. It formulates text responses based on NLU intents like "greeting", "get_help", "get_weather", and a new "goodbye" intent.
*   For the "get_weather" intent, if a location entity is missing, it will now ask the user for a location and remember that it's waiting for this information in the next turn. It then attempts to use the provided location in its (still placeholder) weather response.
*   A new `order_drink` intent has been introduced. If this intent is detected, the service will attempt to fill slots for `drink_name` and `drink_size`. If either is missing, it will ask the user for the missing information in a subsequent turn, using the session state to remember the context of the order. Once all information is gathered, it provides a confirmation response.
*   A "goodbye" intent is now handled, which will result in a polite farewell and clearing of the session state for that user.
*   It now includes specific handling for NLU error intents (such as `error_no_dialogflow_client` or `error_calling_dialogflow`), responding with a message indicating that it's having trouble understanding and suggesting the user try again later. This allows for more graceful behavior when integrated into a simulated pipeline where the NLU service might be providing fallback responses.
*   The actual dialogue state tracking and policy management for complex conversations will be developed in the future.

## Interaction in the System

1.  **Receives from:** `NLUService`. The NLU service calls `DialogueManagementService.ManageTurn` with the NLU results.
2.  **Processing**:
    *   The `DialogueManagementServicer` applies its placeholder dialogue logic to generate a text response.
    *   It then calls `TextToSpeechService.SynthesizeText` with this text response.
3.  **Calls:** `TextToSpeechService`.
4.  **Output**: Returns its `DialogueResponse` (containing the generated text) to the `NLUService`. The result from the `TextToSpeechService` is currently only logged by DM. In a full pipeline, the audio from TTS would be the final output to the user.
