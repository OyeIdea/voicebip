# Dialogue Management (DM) Service

## Purpose

The Dialogue Management (DM) Service is a component of the RevoVoiceAI AI/ML Services Layer. It acts as the conversational "brain," taking the structured output from the Natural Language Understanding (NLU) Service and determining the system's next response or action to guide the conversation.

## Components

*   `service.py`: Contains the gRPC `DialogueManagementServicer` which implements the core dialogue logic.
*   `dialogue_management_service_pb2.py`: Generated Protobuf Python code for DM message structures.
*   `dialogue_management_service_pb2_grpc.py`: Generated Protobuf Python code for DM gRPC client and server stubs.
*   `nlu_service_pb2.py`: Copied from NLU service; contains definitions for `NLUResponse` which is part of `DialogueRequest`.
*   `requirements.txt`: Python package dependencies (`grpcio`, `grpcio-tools`, `protobuf`).
*   `config.py`: (Placeholder) For service-specific configurations.
*   `rules/`: (Placeholder directory) For potential dialogue rules.
*   `state_trackers/`: (Placeholder directory) For dialogue state tracking mechanisms.
*   `__init__.py`: Makes the directory a Python package.

## gRPC Service: DialogueManagementService

The DM Service exposes its functionality via a gRPC interface.

*   **Service Name:** `DialogueManagementService` (defined in `ai_ml_services/protos/dialogue_management_service.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50054`.
*   **RPC Method:** `ManageTurn(DialogueRequest) returns (DialogueResponse)`
    *   **`DialogueRequest`**: Contains the `session_id` and the `nlu_result` (which is an `NLUResponse` message from the NLU service).
    *   **`DialogueResponse`**: Returns the `session_id` and a `text_response` to be conveyed to the user.

## Current Status & Logic

The current implementation of the `DialogueManagementServicer` in `service.py` provides **placeholder dialogue logic**. It does not employ sophisticated state tracking or policy management yet.
The logic is primarily rule-based for demonstration:
*   It inspects the `intent` field from the received `nlu_result`.
*   Based on the intent, it formulates a simple text response:
    *   **"greeting" intent:** "Hello there! How can I help you today?"
    *   **"get_help" intent:** "I understand you need help. I'll do my best to assist you."
    *   **"get_weather" intent:** Attempts to use a "location" entity if present (e.g., "I'm sorry, I can't fetch the actual weather for [location] right now..."). If no location entity, defaults to "your area".
    *   **Unknown/empty intent:** "I'm sorry, I didn't quite understand that..." or "I'm not sure what you mean..."
*   The `session_id` from the request is passed through to the response.
*   Future enhancements would involve more complex state management, policy decisions, and potentially integration with backend systems.

## Interaction in the System

1.  **Input**: Receives a `DialogueRequest` from the `NLUService`. The core of this request is the `NLUResponse` (containing intent, entities, etc.).
2.  **Processing**: The `DialogueManagementServicer.ManageTurn()` method applies its current placeholder logic based on the NLU intent.
3.  **Output**: Returns a `DialogueResponse` containing the system's text response to the `NLUService`. The `NLUService` currently just logs this response. In a more complete pipeline, this response might flow back to a component that handles Text-to-Speech (TTS).
