# RevoVoiceAI Platform - Initial Structure

## Introduction

Welcome to the RevoVoiceAI platform! This repository contains the initial structural outline for a sophisticated, microservices-based, cloud-native architecture designed to handle voice interactions, AI processing, and business system integrations.

This document provides a high-level overview of the components structured so far. For more detailed information on specific layers or services, please refer to their respective README files linked below.

## Architectural Layers & Services

So far, we have begun to structure the following key architectural layers and their core services:

### Voice Gateway Layer

The entry point for all voice communications, handling various signaling and media protocols.
See detailed information in the [Voice Gateway Layer README](./voice_gateway_layer/README.md).

**Services Outlined:**
*   **`sip_gateway` (Go Service)**: Handles SIP/RTP protocols for telephony integration.
*   **`webrtc_gateway` (Go Service)**: Enables WebRTC-based real-time communication from browsers/mobile.

### Real-Time Processing Engine

Handles core voice processing and real-time data streaming.
See detailed information in the [Real-Time Processing Engine README](./real_time_processing_engine/README.md).

**Services Outlined:**
*   **`StreamingDataManager`**: Manages incoming data streams (e.g., audio from calls) and routes them to appropriate processing services.
*   **`SpeechToTextService` (STT)**: Transcribes audio streams into text.
*   **`TextToSpeechService` (TTS)**: Converts textual responses into audible speech.
*   *(Note: VAD and Audio Processing Pipeline services are also structured here, but not listed in this top-level summary for brevity. See layer README.)*


### AI/ML Services Layer

Provides the intelligence and decision-making capabilities for the platform.
See detailed information in the [AI/ML Services Layer README](./ai_ml_services/README.md).

**Services Outlined:**
*   **`NLUService` (Natural Language Understanding)**: Processes input text to identify user intents, extract key entities, and manage conversational context.
*   **`DialogueManagementService` (DM)**: Manages the conversation flow, tracks dialogue state, and determines the system's next actions or responses based on NLU output.

## High-Level Interaction Flow

The core conversational AI pipeline, involving services from these layers, generally follows this sequence:

1.  **Voice Input (via `VoiceGatewayLayer`)**: User speaks; call is established (e.g., via `sip_gateway` (Go Service) or `webrtc_gateway` (Go Service)).
2.  **Stream Management (`StreamingDataManager`)**: Captures and manages the incoming audio stream.
3.  **Speech-to-Text (`SpeechToTextService`)**: Transcribes the audio to text. (VAD and Audio Processing Pipeline may be used here).
4.  **Natural Language Understanding (`NLUService`)**: Interprets the text to understand intent and extract entities.
5.  **Dialogue Management (`DialogueManagementService`)**: Decides the next action or response based on NLU output and conversation state.
6.  **Text-to-Speech (`TextToSpeechService`)**: Converts the system's text response back into audio.
7.  **Voice Output (via `VoiceGatewayLayer`)**: The synthesized audio is played back to the user (e.g., via `sip_gateway` (Go Service) or `webrtc_gateway` (Go Service) and RTP/SRTP).

This flow enables a full, bidirectional voice-based interaction.

## Future Work & Next Steps

This repository currently represents only the initial structural skeleton of the RevoVoiceAI platform. Many other components outlined in the full technical architecture document are yet to be structured, including but not limited to:

*   Other services in the Voice Gateway Layer.
*   Remaining services in the Real-Time Processing Engine (e.g., explicitly listing VAD, Audio Processing Pipeline in more detail).
*   Specialized AI Services (e.g., Voice Cloning, Sentiment Analysis as a dedicated service, Predictive Analytics, Personalization Engine).
*   **Business Logic Layer** (Call Management, Customer Services, Agent Support).
*   **Integration Layer** (CRM/ERP connectors, Omnichannel Hub).
*   **Data Management Layer** (Databases, Data Processing pipelines).
*   **Security & Compliance Layer**.
*   **Management & Monitoring Layer**.

Further development will involve implementing the logic within these structured services, integrating actual AI/ML models, setting up communication protocols between services (e.g., gRPC, REST APIs, message queues), and defining the infrastructure for deployment (e.g., Kubernetes).
