# Call Routing Service

## Purpose

The Call Routing Service is a core component of the RevoVoiceAI Business Logic Layer. Its primary responsibility is to determine the optimal destination or next step for an incoming call or interaction based on a variety of factors. This ensures that callers are efficiently connected to the most appropriate resource, whether it's a specific agent, a queue, an IVR menu, or another automated service.

## Core Functions (Conceptual)

*   **Intent-Based Routing**: Uses NLU (Natural Language Understanding) output (intent, entities) to guide routing decisions.
*   **Skill-Based Routing**: Matches callers with agents possessing specific skills required to handle their query.
*   **Queue Management Integration**: Interacts with queueing systems to place calls in appropriate queues if agents are unavailable.
*   **Priority Handling**: Manages call prioritization based on caller status, urgency, or other business rules.
*   **Data-Driven Decisions**: Can use data from CRM, Customer Profile Service, or other sources to inform routing.
*   **Rule Engine Integration**: May utilize a rule engine for complex routing logic.

## Components

*   `service.py`: Contains the main `CallRoutingService` class. Its `get_routing_decision()` method takes call context information and returns a routing decision.
*   `config.py`: (Placeholder) Intended for service-specific configurations, such as default routing rules, queue IDs, skill definitions, or connection details for dependent services.
*   `__init__.py`: Makes the directory a Python package.

## Interaction with Other Services

The Call Routing Service is typically invoked after initial call processing and NLU:

1.  **Input**: Receives call context data. This usually includes:
    *   Output from the `NLUService` (intent, entities).
    *   Caller information (e.g., `caller_id`, ANI).
    *   Session information from the `SessionManagerService`.
    *   Potentially, data from the `CustomerProfileService`.
2.  **Processing**: The `get_routing_decision()` method evaluates this context against its configured rules and logic. This might involve:
    *   Checking agent availability and skills (requiring integration with an agent management system, not yet structured).
    *   Querying queue statistics.
3.  **Output**: Returns a structured routing decision, specifying:
    *   `route_type`: (e.g., AGENT, QUEUE, IVR, VOICEMAIL, HANGUP).
    *   `target_id`: The specific agent, queue, or IVR menu chosen.
    *   Other parameters like priority.
4.  **Next Step**: An orchestrator or a session management component (perhaps in the Voice Gateway Layer or a higher-level Session Orchestrator in this Business Logic Layer) would then act on this decision to direct the call.

This service is central to providing an efficient and intelligent contact center experience.
