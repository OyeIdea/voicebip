# business_logic_layer/call_routing_service/service.py

class CallRoutingService:
    """
    Service responsible for making decisions on how to route incoming calls
    based on various contextual factors.
    """

    def __init__(self, config=None):
        """
        Initializes the CallRoutingService.

        Args:
            config: Configuration object or dictionary.
                    (Placeholder for future configuration loading from config.py,
                     e.g., routing rules, queue definitions, agent skill mappings)
        """
        self.config = config
        # Future initialization for loading routing rules, connecting to agent presence services, etc.
        print("CallRoutingService initialized.")

    def get_routing_decision(self, call_context: dict) -> dict:
        """
        Determines the routing decision for a call based on its context (NLU, caller info, etc.).

        Args:
            call_context (dict): A dictionary containing information relevant to routing the call.
                                 This might include:
                                 - "nlu_data": dict (output from NLUService, e.g., intent, entities)
                                 - "caller_id": str
                                 - "session_id": str
                                 - "media_type": str (e.g., "voice", "chat")
                                 - "current_queue_time": int (optional)

        Returns:
            dict: A dictionary specifying the routing decision.
        """
        print(f"\nCall Routing: Determining route for call_context: {call_context}")

        if not call_context:
            return {"route_type": "ERROR", "reason": "Missing call_context"}

        # Placeholder logic based on NLU intent
        # Future enhancements:
        # - Integrate with agent availability/presence systems.
        # - Implement skill-based routing.
        # - Consider queue statistics (e.g., current wait times, number of agents available).
        # - Implement more complex rule engine or policy-based routing.
        # - Time-of-day routing, VIP caller routing, etc.

        nlu_data = call_context.get("nlu_data", {})
        # Assuming NLU output structure like: {"intent": {"name": "sales_inquiry", "confidence": 0.9}, ...}
        nlu_intent = nlu_data.get("intent", {}).get("name") 
        
        if nlu_intent == "sales_inquiry":
            decision = {
                "route_type": "QUEUE", 
                "target_id": "queue_sales", 
                "priority": 1,
                "reason": "Sales intent detected"
            }
        elif nlu_intent == "technical_support":
            decision = {
                "route_type": "QUEUE", 
                "target_id": "queue_support_tier1", 
                "priority": 2,
                "reason": "Support intent detected"
            }
        else:
            decision = {
                "route_type": "QUEUE", 
                "target_id": "queue_general", 
                "priority": 5,
                "reason": "Default routing for unclassified or general intent"
            }
            if not nlu_intent:
                 decision["reason"] = "Default routing due to missing NLU intent"
        
        print(f"Call Routing: Decision made: {decision}")
        return decision

# Example usage (optional, for testing or demonstration)
if __name__ == "__main__":
    routing_service = CallRoutingService()

    # Example 1: Sales Inquiry
    call_context_sales = {
        "nlu_data": {"intent": {"name": "sales_inquiry", "confidence": 0.95}, "entities": {}},
        "caller_id": "1234567890",
        "session_id": "session_abc_123",
        "media_type": "voice"
    }
    decision_sales = routing_service.get_routing_decision(call_context_sales)
    # print(f"Sales Call Decision: {decision_sales}")

    # Example 2: Technical Support
    call_context_support = {
        "nlu_data": {"intent": {"name": "technical_support", "confidence": 0.88}, "entities": {"product": "gizmo_x"}},
        "caller_id": "0987654321",
        "session_id": "session_xyz_789",
        "media_type": "voice",
        "current_queue_time": 120 # seconds
    }
    decision_support = routing_service.get_routing_decision(call_context_support)
    # print(f"Support Call Decision: {decision_support}")

    # Example 3: Undefined Intent
    call_context_general = {
        "nlu_data": {"intent": {"name": "general_query", "confidence": 0.70}, "entities": {}},
        "caller_id": "1122334455",
        "session_id": "session_def_456",
        "media_type": "chat"
    }
    decision_general = routing_service.get_routing_decision(call_context_general)
    # print(f"General Query Decision: {decision_general}")

    # Example 4: Missing NLU data
    call_context_no_nlu = {
        "caller_id": "5555555555",
        "session_id": "session_ghi_000",
        "media_type": "voice"
    }
    decision_no_nlu = routing_service.get_routing_decision(call_context_no_nlu)
    # print(f"No NLU Data Decision: {decision_no_nlu}")

    # Example 5: Missing call_context entirely
    decision_no_context = routing_service.get_routing_decision(None)
    # print(f"No Context Decision: {decision_no_context}")
