# RevoVoiceAI - AI/ML Services Layer

This directory houses the services that provide the core intelligence and decision-making capabilities for the RevoVoiceAI platform. These services process data (primarily text from speech transcription) to understand user intent, extract information, manage dialogue, and generate appropriate responses.

## Services

Currently, the following services are outlined:

*   **`nlu_service` (Natural Language Understanding Service)**: Responsible for processing input text to identify user intents, extract key entities, and manage conversational context.

## Conceptual Conversational AI Pipeline

The services within this layer, along with those from other layers like the Real-Time Processing Engine, typically operate in a pipeline to handle user interactions. A common flow is:

1.  **Voice Input**: User speaks.
2.  **Speech-to-Text (STT)**: The audio is transcribed into plain text by a service in the Real-Time Processing Engine (e.g., `SpeechToTextService`).
3.  **Natural Language Understanding (NLU)**: The transcribed text is passed to the `NLUService` within this AI/ML layer. The NLU service analyzes the text to:
    *   Recognize the user's intent(s).
    *   Extract relevant entities.
    *   Update and maintain the dialogue context.
    *   (Optionally) Analyze sentiment.
4.  **Dialogue Management (DM)**: The structured output from the NLU service (intents, entities, context) is then passed to a Dialogue Management service (to be developed). The DM service:
    *   Tracks the conversation state.
    *   Decides on the next action or response based on the NLU output and business logic.
    *   May involve querying business systems or knowledge bases.
5.  **Response Generation / Text-to-Speech (TTS)**:
    *   The Dialogue Manager determines the content of the response.
    *   This response text is then sent to a Text-to-Speech (TTS) service (likely in the Real-Time Processing Engine) to synthesize voice.
6.  **Voice Output**: The synthesized voice is played back to the user.

This modular architecture allows each service to specialize in its specific task, promoting maintainability and scalability. The `NLUService` is a crucial first step in the "understanding" phase of this pipeline.
