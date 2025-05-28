# RevoVoiceAI - AI/ML Services Layer

This directory houses the services that provide the core intelligence and decision-making capabilities for the RevoVoiceAI platform. These services process data (primarily text from speech transcription) to understand user intent, extract information, manage dialogue, and generate appropriate responses.

## Services

Currently, the following services are outlined:

*   **`nlu_service` (Natural Language Understanding Service)**: Responsible for processing input text to identify user intents, extract key entities, and manage conversational context. It may also coordinate with the Sentiment Analysis Engine. See the [NLU Service README](./nlu_service/README.md) for more details.
*   **`dialogue_management_service` (Dialogue Management Service)**: Manages the conversation flow, tracks dialogue state, and determines the system's next actions or responses based on NLU output (including sentiment) and predefined policies. See the [Dialogue Management Service README](./dialogue_management_service/README.md) for more details.
*   **`sentiment_analysis_engine` (Sentiment Analysis Engine Service)**: Dedicated to analyzing text to determine the emotional tone or sentiment. Its output can be used by NLU, DM, or for analytics. See the [Sentiment Analysis Engine README](./sentiment_analysis_engine/README.md) for more details.

## Conceptual Conversational AI Pipeline

The services within this layer, along with those from other layers like the Real-Time Processing Engine, typically operate in a pipeline to handle user interactions. A common flow is:

1.  **Voice Input**: User speaks.
2.  **Speech-to-Text (STT)**: The audio is transcribed into plain text by a service in the Real-Time Processing Engine (e.g., `SpeechToTextService`).
3.  **Natural Language Understanding (NLU) & Sentiment Analysis**:
    *   The transcribed text is passed to the `NLUService`.
    *   The `NLUService` analyzes the text to recognize intents, extract entities, and update dialogue context.
    *   Concurrently or as part of NLU's process, the `SentimentAnalysisEngineService` analyzes the text for emotional tone. The NLU output is then enriched with this sentiment information.
4.  **Dialogue Management (DM)**: The structured output from NLU (intents, entities, context, sentiment) is then passed to the `DialogueManagementService`. The DM service:
    *   Tracks the overall conversation state.
    *   Applies dialogue policies/rules to decide on the next system action (e.g., reply, ask for clarification, query a database, end call).
    *   Determines the content for any response to the user.
5.  **Response Generation / Text-to-Speech (TTS)**:
    *   If the DM decides on a verbal response, the content is finalized.
    *   This response text is then sent to a Text-to-Speech (TTS) service (likely in the Real-Time Processing Engine) to synthesize voice.
6.  **Voice Output**: The synthesized voice is played back to the user.

This modular architecture allows each service to specialize in its specific task, promoting maintainability and scalability. The `NLUService`, `SentimentAnalysisEngineService`, and `DialogueManagementService` are key components in the "understanding" and "decision-making" phases of this pipeline.

## Specialized AI Services

*   **`VoiceCloningService`**: Responsible for voice model training and custom voice synthesis. ([Details](./voice_cloning_service/README.md))
*   **`PredictiveAnalyticsEngine`**: Handles call volume forecasting and customer behavior prediction. ([Details](./predictive_analytics_engine/README.md))
*   **`PersonalizationEngine`**: Manages customer profile analysis and interaction customization. ([Details](./personalization_engine/README.md))
