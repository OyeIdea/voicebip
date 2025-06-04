# ai_ml_services/nlu_service/service.py

import grpc
from concurrent import futures
import time # Required for server.wait_for_termination() in a loop, though not strictly for the current serve()

# Import generated protobuf and gRPC modules
# These should be in the same directory or Python path
import nlu_service_pb2
import nlu_service_pb2_grpc

# The NLUService class below is the gRPC servicer.
# The existing NLUService class (for business logic) can be kept if desired,
# and the servicer could instantiate or use it. For this subtask,
# the servicer will implement the logic directly as per instructions.

class NLUServiceServicer(nlu_service_pb2_grpc.NLUServiceServicer):
    """
    Implements the NLUService gRPC interface.
    """
    def ProcessText(self, request: nlu_service_pb2.NLURequest, context):
        """
        Processes a text input and returns NLU results.
        """
        print(f"NLUService: Received text '{request.text}' for session_id '{request.session_id}'")
        
        # Placeholder NLU logic
        intent = "greeting"
        intent_confidence = 0.95
        entities = [
            nlu_service_pb2.Entity(name="name", value="User", confidence=0.8)
        ]
        processed_text = request.text # Or some cleaned version
        
        # Simple rule-based modification for demonstration
        if "help" in request.text.lower():
            intent = "get_help"
            entities.append(nlu_service_pb2.Entity(name="topic", value="general", confidence=0.7))
        elif "weather" in request.text.lower():
            intent = "get_weather"
            entities.append(nlu_service_pb2.Entity(name="query_type", value="general_weather", confidence=0.85))
            if "tomorrow" in request.text.lower():
                 entities.append(nlu_service_pb2.Entity(name="date", value="tomorrow", confidence=0.9))
        
        print(f"NLUService: Intent='{intent}', Entities={entities}")

        return nlu_service_pb2.NLUResponse(
            session_id=request.session_id,
            intent=intent,
            entities=entities,
            processed_text=processed_text,
            intent_confidence=intent_confidence
        )

def serve():
    """
    Starts the gRPC server for the NLUService.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    nlu_service_pb2_grpc.add_NLUServiceServicer_to_server(NLUServiceServicer(), server)
    
    port = "50053" # Define the port for NLU service
    listen_addr = f'[::]:{port}'
    server.add_insecure_port(listen_addr)
    
    server.start()
    print(f"NLUService server started, listening on port {port}")
    
    try:
        # Keep the server running until interrupted
        # server.wait_for_termination() # This blocks indefinitely
        while True:
            time.sleep(86400) # Sleep for a day, effectively keeping the thread alive
    except KeyboardInterrupt:
        print("NLUService server stopping...")
        server.stop(0) # Graceful stop
        print("NLUService server stopped.")

# Main guard to run the server when the script is executed
if __name__ == '__main__':
    # The original NLUService class (business logic) can be instantiated here
    # if the servicer needs to use it. For this task, it's not directly used by the servicer.
    # e.g., nlu_logic_instance = NLUService()
    # servicer = NLUServiceServicer(nlu_logic=nlu_logic_instance)
    # and then pass `servicer` to add_NLUServiceServicer_to_server.
    
    serve()
