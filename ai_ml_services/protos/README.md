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
    *   This service is currently implemented by the Python `NLUServiceServicer` in the `ai_ml_services/nlu_service` directory, which provides placeholder logic.
