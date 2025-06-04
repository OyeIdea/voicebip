# ai_ml_services/nlu_service/service.py

import grpc
from concurrent import futures
import time # Required for server.wait_for_termination() in a loop

# Import generated protobuf and gRPC modules for NLU Service
import nlu_service_pb2
import nlu_service_pb2_grpc

# Import generated protobuf and gRPC modules for Dialogue Management Service (client)
import dialogue_management_service_pb2
import dialogue_management_service_pb2_grpc

class NLUServiceServicer(nlu_service_pb2_grpc.NLUServiceServicer):
    """
    Implements the NLUService gRPC interface.
    After processing text, it calls the DialogueManagementService.
    """
    def __init__(self):
        # Configuration for Dialogue Management service endpoint
        self.dm_service_address = 'localhost:50054' # Make this configurable

    def ProcessText(self, request: nlu_service_pb2.NLURequest, context):
        """
        Processes a text input, generates NLU results, and then calls DialogueManagementService.
        """
        print(f"NLUService: Received text '{request.text}' for session_id '{request.session_id}'")
        
        # Placeholder NLU logic
        intent = "greeting"
        intent_confidence = 0.95
        entities = [
            nlu_service_pb2.Entity(name="name", value="User", confidence=0.8)
        ]
        processed_text = request.text
        
        if "help" in request.text.lower():
            intent = "get_help"
            entities.append(nlu_service_pb2.Entity(name="topic", value="general", confidence=0.7))
        elif "weather" in request.text.lower():
            intent = "get_weather"
            entities.append(nlu_service_pb2.Entity(name="query_type", value="general_weather", confidence=0.85))
            if "tomorrow" in request.text.lower():
                 entities.append(nlu_service_pb2.Entity(name="date", value="tomorrow", confidence=0.9))
        
        nlu_response = nlu_service_pb2.NLUResponse(
            session_id=request.session_id,
            intent=intent,
            entities=entities,
            processed_text=processed_text,
            intent_confidence=intent_confidence
        )

        entities_str = ', '.join([f"{e.name}='{e.value}' (conf: {e.confidence:.2f})" for e in nlu_response.entities])
        print(f"NLUService: Generated NLU Response for SID {nlu_response.session_id}: Intent='{nlu_response.intent}' (Conf: {nlu_response.intent_confidence:.2f}), Entities=[{entities_str}]")

        # Call DialogueManagementService
        # Using 'with' statement for automatic channel closure
        try:
            with grpc.insecure_channel(self.dm_service_address) as channel:
                dm_stub = dialogue_management_service_pb2_grpc.DialogueManagementServiceStub(channel)

                dialogue_request = dialogue_management_service_pb2.DialogueRequest(
                    session_id=nlu_response.session_id, # Or request.session_id
                    nlu_result=nlu_response
                )

                print(f"NLUService: Calling DialogueManagementService at {self.dm_service_address} for SID {dialogue_request.session_id}")
                dm_response = dm_stub.ManageTurn(dialogue_request, timeout=10) # Adding a timeout

                if dm_response:
                    print(f"NLUService: Received DM response for SID {dm_response.session_id}: TextResponse='{dm_response.text_response}'")
                else:
                    print(f"NLUService: Received no response from DialogueManagementService for SID {dialogue_request.session_id}")

        except grpc.RpcError as e:
            # Using e.code() and e.details() for more structured error info
            print(f"NLUService: Error calling DialogueManagementService for SID {request.session_id}: Code={e.code()}, Details='{e.details()}'")
        except Exception as e:
            print(f"NLUService: An unexpected Python error occurred while calling DMService for SID {request.session_id}: {e}")

        # The NLUService still returns its own nlu_response to its original caller (e.g., SpeechToTextService)
        return nlu_response

def serve():
    """
    Starts the gRPC server for the NLUService.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    nlu_service_pb2_grpc.add_NLUServiceServicer_to_server(NLUServiceServicer(), server)
    
    port = "50053"
    listen_addr = f'[::]:{port}'
    server.add_insecure_port(listen_addr)
    
    server.start()
    print(f"NLUService server started, listening on port {port}")
    
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        print("NLUService server stopping...")
        server.stop(0)
        print("NLUService server stopped.")

if __name__ == '__main__':
    serve()
