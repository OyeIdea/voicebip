# ai_ml_services/dialogue_management_service/service.py

import grpc
from concurrent import futures
import time # Required for server.wait_for_termination() in a loop

# Import generated protobuf and gRPC modules
# These should be in the same directory or Python path
import dialogue_management_service_pb2
import dialogue_management_service_pb2_grpc

# Import NLU message definitions (needed for request.nlu_result)
# This assumes nlu_service_pb2.py is in the same directory or path.
import nlu_service_pb2

# Optional: for more advanced logging
# import logging
# logging.basicConfig(level=logging.INFO)


class DialogueManagementServicer(dialogue_management_service_pb2_grpc.DialogueManagementServiceServicer):
    """
    Implements the DialogueManagementService gRPC interface.
    """
    def ManageTurn(self, request: dialogue_management_service_pb2.DialogueRequest, context):
        """
        Manages a turn in the conversation based on NLU input.
        Returns a text response.
        """
        nlu_result = request.nlu_result
        session_id = request.session_id
        
        # Log the received NLU result
        # Using print for simplicity, replace with logging in a real application
        entities_str = ', '.join([f"{e.name}='{e.value}' (conf: {e.confidence:.2f})" for e in nlu_result.entities])
        print(f"DialogueManagementService: Received NLU for session '{session_id}': Intent='{nlu_result.intent}' (Conf: {nlu_result.intent_confidence:.2f}), Entities=[{entities_str}], ProcessedText='{nlu_result.processed_text}'")
        
        text_response = "I'm sorry, I didn't quite understand that. Could you say it again?" # Default fallback

        # Placeholder logic to determine response based on intent
        if nlu_result.intent == "greeting":
            text_response = "Hello there! How can I help you today?"
        elif nlu_result.intent == "get_help":
            text_response = "I understand you need help. I'll do my best to assist you."
        elif nlu_result.intent == "get_weather":
            location = "your area" # Default location
            date_info = ""
            for entity in nlu_result.entities:
                if entity.name == "location":
                    location = entity.value
                elif entity.name == "date": # Assuming NLU might extract a date entity
                    date_info = f" for {entity.value}"
            
            # More sophisticated weather logic would go here
            text_response = f"I'm sorry, I can't fetch the actual weather for {location}{date_info} right now, but I hope it's a pleasant day!"
        elif not nlu_result.intent: # No intent detected or empty intent string
             text_response = "I'm not sure what you mean. Can you try rephrasing?"

        print(f"DialogueManagementService: Responding to session '{session_id}' with: '{text_response}'")
        
        return dialogue_management_service_pb2.DialogueResponse(
            session_id=session_id,
            text_response=text_response
            # action_code and updated_conversation_context could be set here if needed
        )

def serve():
    """
    Starts the gRPC server for the DialogueManagementService.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    dialogue_management_service_pb2_grpc.add_DialogueManagementServiceServicer_to_server(
        DialogueManagementServicer(), server
    )

    port = "50054" # Define the port for Dialogue Management service
    listen_addr = f'[::]:{port}'
    server.add_insecure_port(listen_addr)

    server.start()
    print(f"DialogueManagementService server started, listening on port {port}")

    try:
        # Keep the server running until interrupted
        # server.wait_for_termination() # This blocks indefinitely
        while True:
            time.sleep(86400) # Sleep for a day, effectively keeping the thread alive
    except KeyboardInterrupt:
        print("DialogueManagementService server stopping...")
        server.stop(0) # Graceful stop
        print("DialogueManagementService server stopped.")

# Main guard to run the server when the script is executed
if __name__ == '__main__':
    # The original DialogueManagementService class (business logic) can be instantiated here
    # if the servicer needs to use it, e.g., for complex state tracking or policy rules.
    # For this task, the servicer implements logic directly.
    # e.g., dm_logic = DialogueManagementService()
    # servicer_instance = DialogueManagementServicer(logic=dm_logic)
    # and then pass `servicer_instance` to add_DialogueManagementServiceServicer_to_server.

    serve()
