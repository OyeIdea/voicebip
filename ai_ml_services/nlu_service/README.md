# Natural Language Understanding (NLU) Service

## Purpose

The Natural Language Understanding (NLU) Service is a component of the RevoVoiceAI AI/ML Services Layer. It uses the **Google Cloud Dialogflow CX SDK** to perform Natural Language Understanding. Its primary function is to process textual input (typically transcribed speech from the Speech-to-Text service) to extract meaning and structure, such as intents and entities. This structured information is then passed to the Dialogue Management (DM) service to inform responses and actions.

## Components

*   `service.py`: Contains the gRPC `NLUServiceServicer`. This servicer interfaces with the Dialogflow CX API and then calls the Dialogue Management service.
*   `config.py`: Manages configuration, including Dialogflow CX project, agent, and location identifiers, and Google Cloud credentials.
*   `nlu_service_pb2.py`, `nlu_service_pb2_grpc.py`: Generated Protobuf/gRPC code for this NLU service.
*   `dialogue_management_service_pb2.py`, `dialogue_management_service_pb2_grpc.py`: Generated Protobuf/gRPC client stubs for calling the Dialogue Management service.
*   `requirements.txt`: Python package dependencies, including `grpcio`, `protobuf`, `google-cloud-dialogflow-cx`, and `python-dotenv`.
*   `models/`: (Placeholder directory) Not directly used by Dialogflow CX integration but kept for potential future alternative NLU models.
*   `__init__.py`: Makes the directory a Python package.

## Configuration & Prerequisites

For the NLUService to function correctly with Dialogflow CX, the following prerequisites must be met and environment variables configured:

1.  **Google Cloud Project:** A Google Cloud Platform project with the Dialogflow API enabled.
2.  **Dialogflow CX Agent:** A Dialogflow CX agent must be created and configured within your project. This agent should have defined intents, entities, and potentially flows/pages relevant to your application.
3.  **Service Account Credentials:**
    *   A service account with appropriate permissions (e.g., "Dialogflow API Client" or "Dialogflow API Reader") must be created in your Google Cloud project.
    *   Download the JSON key file for this service account.
    *   Set the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to the absolute path of this JSON key file. This is how Google Cloud client libraries automatically authenticate.
    ```bash
    export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account-key.json"
    ```

4.  **Environment Variables for NLU Service:**
    *   `GOOGLE_APPLICATION_CREDENTIALS`: (As described above) Path to your service account key file.
    *   `DIALOGFLOW_PROJECT_ID`: Your Google Cloud Project ID where the CX Agent resides.
    *   `DIALOGFLOW_AGENT_ID`: The Agent ID of your Dialogflow CX agent (found in Agent settings).
    *   `DIALOGFLOW_LOCATION_ID`: The region/location of your CX agent (e.g., "global", "us-central1"). Defaults to "global" if not set.
    *   `DIALOGFLOW_LANGUAGE_CODE`: The language code to be used for Dialogflow interactions (e.g., "en-US", "en"). Defaults to "en-US" if not set.

    For local development, these variables (except typically `GOOGLE_APPLICATION_CREDENTIALS`, which is better set in the shell) can be placed in a `.env` file within the `ai_ml_services/nlu_service/` directory. `config.py` uses `python-dotenv` to load this file.
    Example `.env` content:
    ```env
    DIALOGFLOW_PROJECT_ID=your-gcp-project-id
    DIALOGFLOW_AGENT_ID=your-cx-agent-id
    DIALOGFLOW_LOCATION_ID=us-central1 # or global
    DIALOGFLOW_LANGUAGE_CODE=en
    # GOOGLE_APPLICATION_CREDENTIALS should ideally be set in your shell environment
    ```
    **Do not commit the `.env` file or service account keys to version control.**

## gRPC Service: NLUService

*   **Service Name:** `NLUService` (defined in `ai_ml_services/protos/nlu_service.proto`)
*   **Port:** The gRPC server listens on `0.0.0.0:50053`.
*   **RPC Method:** `ProcessText(NLURequest) returns (NLUResponse)`
    *   **`NLURequest`**: Contains the input `text` and `session_id`.
    *   **`NLUResponse`**: Returns the `session_id`, detected `intent`, `entities`, `processed_text`, and `intent_confidence`.
    *   **Behavior**:
        1.  Receives an `NLURequest` (typically from `SpeechToTextService`).
        2.  If the Dialogflow CX client (`self.sessions_client`) was not initialized successfully (due to missing configuration), it returns an error NLUResponse.
        3.  Constructs a Dialogflow CX session path using the project, location, agent, and the `session_id` from the request.
        4.  Creates a `QueryInput` with the request text and language code.
        5.  Calls the `detect_intent` method of the Dialogflow CX `SessionsClient`.
        6.  **Maps Dialogflow CX `QueryResult` to `NLUResponse`**:
            *   `intent`: From `query_result.intent.display_name`. If no intent is matched, defaults to "no_intent_matched".
            *   `intent_confidence`: From `query_result.intent_detection_confidence`.
            *   `entities`: Dialogflow `query_result.parameters` are converted to `Entity` messages. Simple types (string, number, boolean) are converted to their string representation. Complex types (structs, lists) are serialized to a JSON string and stored in the `Entity.value` field. A default confidence of 1.0 is assigned to entities derived from parameters.
            *   `processed_text`: From `query_result.text` (the text Dialogflow used for processing).
        7.  Handles exceptions during the Dialogflow API call, returning an error NLUResponse if an issue occurs.
        8.  **Calls Dialogue Management Service**: After obtaining the `NLUResponse` from Dialogflow CX (or an error response), it creates a `DialogueRequest` and calls the `ManageTurn` RPC of the `DialogueManagementService` (running on port `50054`).
        9.  The response from `DialogueManagementService` is logged.
        10. Finally, the `NLUService` returns the `NLUResponse` (from Dialogflow CX or error) to its original caller.

### Fallback Behavior / Simulated Pipeline Note

It's important to understand how the `NLUService` behaves if the Dialogflow CX client (`self.sessions_client`) is not properly initialized or if there's an issue calling the Dialogflow CX API:

*   **Initialization Failure:** If the Dialogflow CX client fails to initialize during the service startup (e.g., due to missing or invalid `GOOGLE_APPLICATION_CREDENTIALS`, `DIALOGFLOW_PROJECT_ID`, or `DIALOGFLOW_AGENT_ID`), the `ProcessText` method will not attempt to call the external Dialogflow CX API.
*   **Error Response Generation:** In such cases (initialization failure or errors during an attempted API call), the service returns a predefined error-like `NLUResponse`. This response will typically include:
    *   An `intent` like `"error_no_dialogflow_client"` or `"error_calling_dialogflow"`.
    *   A low `intent_confidence` (e.g., 0.0).
    *   An empty list of `entities`.
*   **Pipeline Testing:** This fallback behavior is crucial for testing the broader RevoVoiceAI pipeline, especially when actual NLU processing via Dialogflow CX is not required for a specific test, or if the full Dialogflow CX setup is pending. It allows downstream services to be tested with a known error-state NLU output.
*   **Continued Processing:** The `NLUService` will still proceed to call the `DialogueManagementService` with this fallback/error `NLUResponse`. This ensures that the subsequent services in the pipeline are exercised even if NLU processing itself could not be completed.

## Key Dependencies
*   `google-cloud-dialogflow-cx`: The Google Cloud client library for Dialogflow CX.
*   `python-dotenv`: For managing environment variables during local development.
*   `grpcio`, `protobuf`: For gRPC communication.

## Interaction in the System

1.  **Receives from:** `SpeechToTextService`. The STT service calls `NLUService.ProcessText` with the transcribed text.
2.  **Interacts with:** Google Cloud Dialogflow CX for NLU processing.
3.  **Calls:** `DialogueManagementService` with the NLU results.
4.  **Output**: Returns its `NLUResponse` (derived from Dialogflow CX) to the `SpeechToTextService`.
