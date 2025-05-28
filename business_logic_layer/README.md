# Business Logic Layer

## Purpose

The Business Logic Layer (BLL) is a central part of the RevoVoiceAI platform, responsible for orchestrating application workflows, implementing core business rules, and managing the state and logic that define how the platform behaves beyond the raw processing of voice and AI.

This layer acts as an intermediary between the presentation/gateway layers and the data/AI service layers, ensuring that business processes are followed correctly. Services in this layer are currently planned to be implemented in Python for ease of integration with AI/ML services and general application development.

## Services

The services within this layer handle specific domains of business functionality:

*   **`call_routing_service`**: Determines the optimal destination for incoming calls based on various factors like NLU intent, agent skills, and queue status. See the [Call Routing Service README](./call_routing_service/README.md) for more details.
*   *(Other services like `CustomerProfileService` and `RealtimeAgentAssistantService` will be added here as they are structured.)*

## Key Responsibilities

*   **Workflow Orchestration**: Managing multi-step business processes (e.g., complex call flows, customer interaction sequences).
*   **Business Rule Execution**: Applying specific rules that define application behavior (e.g., routing logic, escalation policies, personalization strategies).
*   **Data Aggregation & Transformation**: Combining data from various sources (e.g., CRM, customer profiles, AI service outputs) to make informed decisions.
*   **Interface with Data Layer**: Managing interactions with databases and other data stores for business entities.
*   **State Management**: Handling application-level state that spans multiple interactions or services.

This layer is crucial for tailoring the RevoVoiceAI platform to specific business needs and ensuring consistent operational logic.
