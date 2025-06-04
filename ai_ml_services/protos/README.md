# Protocol Buffer Definitions for AI/ML Services

This directory contains the Protocol Buffer (`.proto`) files that define the data structures and service interfaces used by services within the `ai_ml_services` layer of the RevoVoiceAI platform.

## `nlu_service.proto`

This file defines the interface for the Natural Language Understanding (NLU) Service.

**Package:** `ai_ml_services.nlu`

**Go Package:** `revovoiceai/ai_ml_services/protos/nlu_service`
(This `go_package` option is specified in the proto file for Go code generation, should it be used for Go-based NLU clients or servers in the future.)

### Message: `NLURequest`

Represents a request to the NLU service for processing a piece of text.

*   `string text = 1;`
    *   The input text that needs to be processed by the NLU model (e.g., a user's utterance, a transcribed sentence).
*   `string session_id = 2;`
    *   A unique identifier for the ongoing session or conversation. This can be used by the NLU service to maintain context across multiple turns.
*   `// map<string, string> context = 3;`
    *   (Currently commented out in the `.proto` file) An optional field that could be used to explicitly pass conversational context (e.g., previous intents, slots filled) if needed by the NLU model.

### Message: `Entity`

Represents a single piece of structured information (an entity) extracted from the input text.

*   `string name = 1;`
    *   The name or type of the entity (e.g., "location", "time", "person_name", "product").
*   `string value = 2;`
    *   The actual value of the extracted entity (e.g., "New York", "tomorrow at 3 PM", "Alice", "SuperWidget X").
*   `float confidence = 3;`
    *   A score between 0.0 and 1.0 indicating the NLU model's confidence in the correctness of the extracted entity.

### Message: `NLUResponse`

Represents the structured output from the NLU service after processing an `NLURequest`.

*   `string session_id = 1;`
    *   The session identifier, passed through from the `NLURequest`.
*   `string intent = 2;`
    *   The primary intent detected in the input text (e.g., "book_flight", "query_weather", "place_order").
*   `repeated Entity entities = 3;`
    *   A list of `Entity` messages, representing all the entities extracted from the text.
*   `string processed_text = 4;`
    *   The version of the text that was actually used for NLU processing. This might be the same as the input text or could be a cleaned/normalized version (e.g., after removing punctuation or correcting typos).
*   `float intent_confidence = 5;`
    *   A score between 0.0 and 1.0 indicating the NLU model's confidence in the detected `intent`.

### Service: `NLUService`

Defines the gRPC service contract for NLU processing.

*   **RPC: `ProcessText (NLURequest) returns (NLUResponse)`**
    *   **Role:** This is the primary method for the NLU service. It takes an `NLURequest` (containing the text to analyze), performs NLU processing, and returns an `NLUResponse` (containing the extracted intent, entities, and other relevant information).
    *   This service is currently implemented by the Python `NLUServiceServicer` in the `ai_ml_services/nlu_service` directory.

---

## `dialogue_management_service.proto`

This file defines the interface for the Dialogue Management (DM) Service.

**Package:** `ai_ml_services.dialogue_management`

**Go Package:** `revovoiceai/ai_ml_services/protos/dialogue_management_service`

**Imports:** `nlu_service.proto` (from the same directory)

### Message: `DialogueRequest`

Represents a request to the DM service to manage a conversational turn.

*   `string session_id = 1;`
    *   A unique identifier for the ongoing session or conversation.
*   `ai_ml_services.nlu.NLUResponse nlu_result = 2;`
    *   The structured output from the `NLUService` for the current user utterance. This provides the DM service with the detected intent and entities.
*   `// string user_id = 3;`
    *   (Commented out) Optional field for user-specific personalization.
*   `// map<string, string> conversation_context = 4;`
    *   (Commented out) Optional field for passing explicit conversation state/context if not fully managed by the DM service itself or if additional external context is needed.

### Message: `DialogueResponse`

Represents the response from the DM service for a given turn.

*   `string session_id = 1;`
    *   The session identifier, passed through from the `DialogueRequest`.
*   `string text_response = 2;`
    *   The text response that the system should convey to the user. This could be a direct answer, a clarifying question, or a statement to guide the conversation.
*   `// string action_code = 3;`
    *   (Commented out) Optional field to specify a non-verbal action the system should take (e.g., "ROUTE_TO_AGENT", "QUERY_DATABASE").
*   `// map<string, string> updated_conversation_context = 4;`
    *   (Commented out) Optional field for the DM service to return the updated conversation state explicitly.

### Service: `DialogueManagementService`

Defines the gRPC service contract for managing dialogue.

*   **RPC: `ManageTurn (DialogueRequest) returns (DialogueResponse)`**
    *   **Role:** This is the primary method for the DM service. It takes a `DialogueRequest` (which includes the NLU results for the last user turn), processes it according to its internal dialogue logic and state, and returns a `DialogueResponse` indicating the system's next step (primarily the text response).
    *   This service is currently implemented by the Python `DialogueManagementServicer` in the `ai_ml_services/dialogue_management_service` directory.
