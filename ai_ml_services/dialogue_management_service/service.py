# ai_ml_services/dialogue_management_service/service.py

import grpc
from concurrent import futures
import time # Required for server.wait_for_termination() in a loop

# Import generated protobuf and gRPC modules for DM Service
import dialogue_management_service_pb2
import dialogue_management_service_pb2_grpc

# Import NLU message definitions (needed for request.nlu_result)
import nlu_service_pb2

# Import generated protobuf and gRPC modules for TextToSpeech Service (client)
import tts_service_pb2
import tts_service_pb2_grpc

class DialogueManagementServicer(dialogue_management_service_pb2_grpc.DialogueManagementServiceServicer):
    """
    Implements the DialogueManagementService gRPC interface.
    After determining a response, it calls the TextToSpeechService.
    """
    def __init__(self):
        # Configuration for TextToSpeech service endpoint
        self.tts_service_address = 'localhost:50055' # Make this configurable

    def ManageTurn(self, request: dialogue_management_service_pb2.DialogueRequest, context):
        """
        Manages a turn in the conversation based on NLU input,
        determines a text response, and then calls TextToSpeechService.
        """
        nlu_result = request.nlu_result
        session_id = request.session_id
        
        entities_str = ', '.join([f"{e.name}='{e.value}' (conf: {e.confidence:.2f})" for e in nlu_result.entities])
        print(f"DialogueManagementService: Received NLU for SID '{session_id}': Intent='{nlu_result.intent}' (Conf: {nlu_result.intent_confidence:.2f}), Entities=[{entities_str}]")
        
        text_response = "I'm sorry, I didn't quite understand that. Could you say it again?"

        if nlu_result.intent == "greeting":
            text_response = "Hello there! How can I help you today?"
        elif nlu_result.intent == "get_help":
            text_response = "I understand you need help. I'll do my best to assist you."
        elif nlu_result.intent == "get_weather":
            location = "your area"
            date_info = ""
            for entity in nlu_result.entities:
                if entity.name == "location":
                    location = entity.value
                elif entity.name == "date":
                    date_info = f" for {entity.value}"
            text_response = f"I'm sorry, I can't fetch the actual weather for {location}{date_info} right now, but I hope it's a pleasant day!"
        elif not nlu_result.intent:
             text_response = "I'm not sure what you mean. Can you try rephrasing?"

        print(f"DialogueManagementService: Determined text response for SID '{session_id}': '{text_response}'")

        # Call TextToSpeechService
        try:
            with grpc.insecure_channel(self.tts_service_address) as channel:
                tts_stub = tts_service_pb2_grpc.TextToSpeechServiceStub(channel)
                tts_request = tts_service_pb2.TTSRequest(
                    text_to_synthesize=text_response,
                    session_id=session_id,
                    voice_config_id="default_voice" # Example voice config
                )

                print(f"DialogueManagementService: Calling TextToSpeechService at {self.tts_service_address} for SID '{session_id}' with text: '{text_response}'")
                tts_response = tts_stub.SynthesizeText(tts_request, timeout=10) # Adding a timeout

                if tts_response:
                    print(f"DialogueManagementService: Received TTS response for SID '{tts_response.session_id}': Status='{tts_response.status_message}'")
                else:
                    print(f"DialogueManagementService: Received no response from TextToSpeechService for SID '{session_id}'")

        except grpc.RpcError as e:
            print(f"DialogueManagementService: Error calling TextToSpeechService for SID '{session_id}': Code={e.code()}, Details='{e.details()}'")
        except Exception as e:
            print(f"DialogueManagementService: An unexpected Python error occurred while calling TTS for SID '{session_id}': {e}")
        
        # The DM Service still returns its own DialogueResponse (containing the determined text)
        # to its original caller (e.g., NLUService).
        return dialogue_management_service_pb2.DialogueResponse(
            session_id=session_id,
            text_response=text_response
        )

def serve():
    """
    Starts the gRPC server for the DialogueManagementService.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    dialogue_management_service_pb2_grpc.add_DialogueManagementServiceServicer_to_server(
        DialogueManagementServicer(), server
    )

    port = "50054"
    listen_addr = f'[::]:{port}'
    server.add_insecure_port(listen_addr)

    server.start()
    print(f"DialogueManagementService server started, listening on port {port}")

    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        print("DialogueManagementService server stopping...")
        server.stop(0)
        print("DialogueManagementService server stopped.")

if __name__ == '__main__':
    serve()
