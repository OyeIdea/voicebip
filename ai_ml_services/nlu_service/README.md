# Natural Language Understanding (NLU) Service

## Purpose

The Natural Language Understanding (NLU) Service is a key component of the RevoVoiceAI AI/ML Services Layer. Its primary function is to process textual input (typically transcribed speech from the Speech-to-Text service) to extract meaning and structure. This structured information is then used by other services, such as Dialogue Management, to inform responses and actions.

## Core Capabilities (Conceptual)

The NLU service aims to provide the following capabilities:

*   **Intent Recognition**: Identifying the user's goal or purpose behind their utterance (e.g., `request_balance`, `book_appointment`).
*   **Entity Extraction**: Identifying and extracting key pieces of information (entities) from the text, such as dates, names, locations, or product names (e.g., extracting "tomorrow" as a date entity).
*   **Context Management**: Maintaining and updating the conversational context, which helps in understanding follow-up questions and resolving ambiguities.
*   **Sentiment Analysis**: (Future capability) Assessing the emotional tone of the user's text (e.g., positive, negative, neutral).

## Components

*   `service.py`: Contains the main `NLUService` class. This class orchestrates the NLU pipeline by calling internal methods for intent recognition, entity extraction, and context management. Its main method is `process_text()`.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as paths to NLU models, language settings, or thresholds for confidence scores.
*   `utils.py`: (Placeholder) Intended for any utility functions supporting NLU operations.
*   `models/`: (Placeholder directory) Intended to store NLU model files, vocabularies, or related assets.
*   `__init__.py`: Makes the directory a Python package.

## Interaction in the Conversational AI Pipeline

The NLU Service typically fits into the pipeline as follows:

1.  **Input**: Receives plain text, usually from the Speech-to-Text (STT) service.
2.  **Processing**: The `NLUService.process_text()` method is called. Internally, this involves:
    *   Recognizing intent(s).
    *   Extracting relevant entities.
    *   Updating the dialogue context.
3.  **Output**: Produces a structured output (e.g., a JSON object or Python dictionary) containing the recognized intents, extracted entities, updated context, and potentially sentiment scores.
4.  **Next Step**: This structured output is then typically passed to a Dialogue Management service, which decides on the appropriate next action or response.

This service will eventually integrate with specific NLU libraries or pre-trained models (e.g., from Hugging Face Transformers, spaCy, or custom models) to perform these tasks.
