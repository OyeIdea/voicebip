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
        self.session_states = {} # Stores state information per session_id

    def ManageTurn(self, request: dialogue_management_service_pb2.DialogueRequest, context):
        """
        Manages a turn in the conversation based on NLU input,
        determines a text response, and then calls TextToSpeechService.
        """
        nlu_result = request.nlu_result
        session_id = request.session_id

        # Retrieve or initialize session state
        session_state = self.session_states.get(session_id, {})
        self.session_states[session_id] = session_state # Ensure it's in the main dict if new
        
        entities_str = ', '.join([f"{e.name}='{e.value}' (conf: {e.confidence:.2f})" for e in nlu_result.entities])
        print(f"DialogueManagementService: SID '{session_id}', NLU Intent='{nlu_result.intent}' (Conf: {nlu_result.intent_confidence:.2f}), Entities=[{entities_str}], Current Session State: {session_state}")
        
        # Default response
        text_response = "I'm sorry, I didn't quite understand that. Could you say it again?"
        session_state['last_intent'] = "unknown_intent_or_empty" # Default, override below

        # 1. Handle pending questions first
        if session_state.get('pending_question') == 'ask_location_for_weather':
            location_response_found = False
            location_value = ""
            # Try to find a "location" entity
            for entity in nlu_result.entities:
                if entity.name == "location":
                    location_value = entity.value
                    location_response_found = True
                    break

            if not location_response_found and nlu_result.processed_text:
                # Use non-empty processed_text as a potential location if no entity found
                location_value = nlu_result.processed_text
                location_response_found = True
                print(f"DialogueManagementService: SID '{session_id}', No location entity, using processed_text '{location_value}' as location.")

            if location_response_found:
                text_response = f"Got it. I'm sorry, I can't fetch the actual weather for {location_value} right now, but I hope it's a pleasant day!"
                session_state.pop('pending_question', None)
                session_state['last_intent'] = "get_weather_location_answered"
            else:
                text_response = "I still need a location for the weather forecast. Could you please tell me where?"
                # session_state['pending_question'] remains 'ask_location_for_weather'
                session_state['last_intent'] = "get_weather_still_asking_location"

        # 2. Handle intents if no pending question was exclusively handled
        elif nlu_result.intent == "greeting":
            text_response = "Hello there! How can I help you today?"
            session_state['last_intent'] = "greeting"
            session_state.pop('pending_question', None) # Clear any pending question on greeting
        elif nlu_result.intent == "get_help":
            text_response = "I understand you need help. I'll do my best to assist you."
            session_state['last_intent'] = "get_help"
            session_state.pop('pending_question', None)
        elif nlu_result.intent == "get_weather":
            location = "your area" # Default if no entity
            date_info = ""
            location_entity_found = False
            for entity in nlu_result.entities:
                if entity.name == "location":
                    location = entity.value
                    location_entity_found = True
                elif entity.name == "date":
                    date_info = f" for {entity.value}"

            if location_entity_found: # A specific location was provided in this turn
                text_response = f"I'm sorry, I can't fetch the actual weather for {location}{date_info} right now, but I hope it's a pleasant day!"
                session_state.pop('pending_question', None)
                session_state['last_intent'] = "get_weather_location_provided"
            else: # No location entity found in this turn
                text_response = "For which location would you like the weather?"
                session_state['pending_question'] = "ask_location_for_weather"
                session_state['last_intent'] = "get_weather_asking_location"
        elif nlu_result.intent == "goodbye":
            text_response = "Goodbye! Have a great day."
            session_state['last_intent'] = "goodbye" # Set before popping
            self.session_states.pop(session_id, None) # Clear state for this session
            # No need to pop pending_question as entire state is cleared
        elif nlu_result.intent == "error_no_dialogflow_client" or nlu_result.intent == "error_calling_dialogflow":
            text_response = "I'm having a little trouble with my understanding capabilities at the moment. Please try again later."
            session_state.pop('pending_question', None)
            session_state['last_intent'] = nlu_result.intent
        elif not nlu_result.intent: # Catches empty intents if not handled by pending_question logic
             text_response = "I'm not sure what you mean. Can you try rephrasing?"
             session_state['last_intent'] = "unknown_intent_or_empty"
             session_state.pop('pending_question', None) # Optional: clear pending q on unknown
        # else:
            # The default text_response from the top is used if no condition above is met.
            # session_state['last_intent'] is already "unknown_intent_or_empty"

        print(f"DialogueManagementService: SID '{session_id}', Determined text response: '{text_response}', Updated Session State: {self.session_states.get(session_id)}")

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
