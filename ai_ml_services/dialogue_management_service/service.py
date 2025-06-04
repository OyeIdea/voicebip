# ai_ml_services/dialogue_management_service/service.py

# from ..nlu_service.service import NLUService # Assuming NLUService output structure

class DialogueManagementService:
    """
    Dialogue Management (DM) Service.
    Responsible for managing the flow of conversation, tracking state,
    applying policies, and determining system actions.
    """

    def __init__(self, config=None):
        """
        Initializes the DialogueManagementService.

        Args:
            config: Configuration object or dictionary.
                    (Placeholder for future configuration loading, e.g.,
                     paths to dialogue rules from config.py or rules/)
        """
        self.config = config
        # Future initialization for dialogue rules, state trackers, etc.
        print("DialogueManagementService initialized.")

    def _track_state(self, current_state: dict = None, nlu_output: dict = None) -> dict:
        """
        Updates and returns the dialogue state based on NLU output and current state.
        """
        if current_state is None:
            updated_state = {"history": []}
        else:
            updated_state = current_state.copy() # Avoid modifying the original object directly

        if nlu_output:
            # For simplicity, just store the NLU output in history.
            # A more sophisticated tracker might extract specific slots, intents, etc.
            updated_state["history"].append(nlu_output)
            # Potentially update other state variables based on NLU output
            updated_state["last_nlu_intent"] = nlu_output.get("intent", {}).get("intent")
            updated_state["last_nlu_entities"] = nlu_output.get("entities", {}).get("entities")
        
        print(f"DM Service: State tracked. New state: {updated_state}")
        return updated_state

    def _apply_policy(self, state: dict) -> dict:
        """
        Applies dialogue policy to determine the next system action based on the current state.
        """
        # Simple rule-based policy
        last_intent = state.get("last_nlu_intent")
        
        if last_intent == "greet":
            policy_decision = {"action": "reply", "message_key": "greeting_response"}
        elif last_intent == "order_pizza":
            # Example for a more specific intent
            policy_decision = {"action": "reply", "message_key": "pizza_order_confirmation"}
        else:
            policy_decision = {"action": "reply", "message_key": "default_fallback"}
            
        print(f"DM Service: Policy applied. Decision: {policy_decision}")
        return policy_decision

    def _generate_response(self, action_details: dict) -> str:
        """
        Generates or retrieves a response string based on action details (e.g., a message key).
        """
        message_key = action_details.get("message_key")
        
        if message_key == "greeting_response":
            response = "Hello! How can I help you today?"
        elif message_key == "pizza_order_confirmation":
            response = "Sure, I can help with that. What kind of pizza would you like?"
        elif message_key == "default_fallback":
            response = "Sorry, I didn't understand that. Can you please rephrase?"
        else:
            response = "I'm not sure how to respond to that."
            
        print(f"DM Service: Response generated for key '{message_key}': '{response}'")
        return response

    def determine_next_action(self, nlu_output: dict, current_state: dict = None) -> dict:
        """
        Determines the next system action based on NLU output and current dialogue state.

        Args:
            nlu_output (dict): The output from the NLUService.
                               Example: {"text": "...", "intent": {"intent": "greet", ...}, ...}
            current_state (dict, optional): The current dialogue state. Defaults to None.

        Returns:
            dict: A dictionary specifying the system action.
                  Example: {
                      "action_type": "reply",
                      "message_to_user": "Hello!",
                      "new_state": {...},
                      "action_details": {"action": "reply", "message_key": "greeting_response"}
                  }
        """
        print(f"\nDM Service: Determining next action for NLU output: {nlu_output}")
        if current_state:
            print(f"DM Service: Current state: {current_state}")

        updated_state = self._track_state(current_state, nlu_output)
        policy_decision = self._apply_policy(updated_state)
        
        message_to_user = None
        if policy_decision.get("action") == "reply":
            message_to_user = self._generate_response(policy_decision)
            
        system_action = {
            "action_type": policy_decision.get("action"),
            "message_to_user": message_to_user,
            "new_state": updated_state,
            "action_details": policy_decision
        }
        
        print(f"DM Service: Next action determined: {system_action}")
        return system_action

# Example usage (optional, for testing or demonstration)
if __name__ == "__main__":
    dm_service = DialogueManagementService()

    # Mock NLU output for a greeting
    mock_nlu_greet = {
        "text": "Hello there",
        "intent": {"intent": "greet", "confidence": 0.95},
        "entities": {},
        "context": {"user_id": "user123"} 
    }

    print("\n--- Example 1: Greeting ---")
    action_result_1 = dm_service.determine_next_action(nlu_output=mock_nlu_greet)
    # print(f"Action Result 1: {action_result_1}")

    # Mock NLU output for an order, using the state from the previous turn
    mock_nlu_order = {
        "text": "I want to order a pizza",
        "intent": {"intent": "order_pizza", "confidence": 0.88},
        "entities": {"pizza_type": "pepperoni"}, # Simplified
        "context": action_result_1["new_state"].get("history", [{}])[-1].get("context") # Get context from NLU output in history
    }
    
    print("\n--- Example 2: Order Pizza (with state from previous turn) ---")
    action_result_2 = dm_service.determine_next_action(
        nlu_output=mock_nlu_order, 
        current_state=action_result_1["new_state"]
    )
    # print(f"Action Result 2: {action_result_2}")

    # Mock NLU output for an unknown intent
    mock_nlu_unknown = {
        "text": "Can you sing a song?",
        "intent": {"intent": "request_song", "confidence": 0.70}, # Assuming this intent is not handled
        "entities": {},
        "context": action_result_2["new_state"].get("history", [{}])[-1].get("context")
    }

    print("\n--- Example 3: Unknown Intent (with state from previous turn) ---")
    action_result_3 = dm_service.determine_next_action(
        nlu_output=mock_nlu_unknown,
        current_state=action_result_2["new_state"]
    )
    # print(f"Action Result 3: {action_result_3}")
