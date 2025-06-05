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
        pending_question = session_state.get('pending_question')
        pending_intent_data = session_state.get('pending_intent_data', {})

        if pending_question == 'ask_location_for_weather':
            location_response_found = False
            location_value = ""
            for entity in nlu_result.entities:
                if entity.name == "location":
                    location_value = entity.value
                    location_response_found = True
                    break
            if not location_response_found and nlu_result.processed_text:
                location_value = nlu_result.processed_text
                location_response_found = True
                print(f"DialogueManagementService: SID '{session_id}', No location entity, using processed_text '{location_value}' as location.")

            if location_response_found:
                text_response = f"Got it. I'm sorry, I can't fetch the actual weather for {location_value} right now, but I hope it's a pleasant day!"
                session_state.pop('pending_question', None)
                # No pending_intent_data for weather currently, so no pop needed here
                session_state['last_intent'] = "get_weather_location_answered"
            else:
                text_response = "I still need a location for the weather forecast. Could you please tell me where?"
                session_state['last_intent'] = "get_weather_still_asking_location"

        elif pending_question == 'ask_drink_name':
            drink_name_response = nlu_result.processed_text # Or extract from a "drink_name" entity if available
            if drink_name_response: # Assuming any non-empty text is the drink name
                pending_intent_data['drink_name'] = drink_name_response
                session_state['pending_intent_data'] = pending_intent_data

                # Check if size is already known (e.g., from initial order_drink intent or if provided with name)
                drink_size_value = pending_intent_data.get('drink_size')
                if not drink_size_value: # Try to get from current entities if not already in pending_intent_data
                    for entity in nlu_result.entities:
                        if entity.name == "drink_size": # Assuming "drink_size" entity
                            drink_size_value = entity.value
                            pending_intent_data['drink_size'] = drink_size_value
                            break

                if not drink_size_value:
                    text_response = f"What size for the {pending_intent_data['drink_name']}?"
                    session_state['pending_question'] = 'ask_drink_size'
                    session_state['last_intent'] = 'order_drink_asking_size'
                else: # Size is known
                    text_response = f"Okay, one {pending_intent_data['drink_size']} {pending_intent_data['drink_name']} coming up!"
                    session_state.pop('pending_question', None)
                    session_state.pop('pending_intent_data', None)
                    session_state['last_intent'] = 'order_drink_complete'
            else:
                text_response = "I didn't catch the drink name. What would you like to order?"
                session_state['last_intent'] = 'order_drink_still_asking_name'
                # pending_question remains 'ask_drink_name'

        elif pending_question == 'ask_drink_size':
            drink_size_response = nlu_result.processed_text # Or extract from a "drink_size" entity
            if drink_size_response: # Assuming any non-empty text is the size
                pending_intent_data['drink_size'] = drink_size_response # Though not strictly needed if completing
                drink_name_value = pending_intent_data.get('drink_name', 'previously mentioned drink')
                text_response = f"Okay, one {drink_size_response} {drink_name_value} coming up!"
                session_state.pop('pending_question', None)
                session_state.pop('pending_intent_data', None)
                session_state['last_intent'] = 'order_drink_complete'
            else:
                drink_name_value = pending_intent_data.get('drink_name', 'that drink')
                text_response = f"I didn't catch the size for {drink_name_value}. What size would you like?"
                session_state['last_intent'] = 'order_drink_still_asking_size'
                # pending_question remains 'ask_drink_size'

        # 2. Handle intents if no pending question was exclusively handled above
        elif nlu_result.intent == "greeting":
            text_response = "Hello there! How can I help you today?"
            session_state['last_intent'] = "greeting"
            session_state.pop('pending_question', None)
            session_state.pop('pending_intent_data', None)
        elif nlu_result.intent == "get_help":
            text_response = "I understand you need help. I'll do my best to assist you."
            session_state['last_intent'] = "get_help"
            session_state.pop('pending_question', None)
            session_state.pop('pending_intent_data', None)
        elif nlu_result.intent == "get_weather":
            session_state.pop('pending_intent_data', None) # Clear drink context
            location = "your area"
            date_info = ""
            location_entity_found = False
            for entity in nlu_result.entities:
                if entity.name == "location":
                    location = entity.value
                    location_entity_found = True
                elif entity.name == "date":
                    date_info = f" for {entity.value}"

            if location_entity_found:
                text_response = f"I'm sorry, I can't fetch the actual weather for {location}{date_info} right now, but I hope it's a pleasant day!"
                session_state.pop('pending_question', None)
                session_state['last_intent'] = "get_weather_location_provided"
            else:
                text_response = "For which location would you like the weather?"
                session_state['pending_question'] = "ask_location_for_weather"
                session_state['last_intent'] = "get_weather_asking_location"

        elif nlu_result.intent == "order_drink":
            # Initialize or retrieve pending_intent_data for this intent
            current_pending_data = session_state.get('pending_intent_data', {})

            # Extract entities if present
            for entity in nlu_result.entities:
                if entity.name == "drink_name": # Assuming entity name from NLU
                    current_pending_data['drink_name'] = entity.value
                elif entity.name == "drink_size": # Assuming entity name from NLU
                    current_pending_data['drink_size'] = entity.value

            drink_name_known = 'drink_name' in current_pending_data
            drink_size_known = 'drink_size' in current_pending_data

            if not drink_name_known:
                text_response = "What drink would you like to order?"
                session_state['pending_question'] = 'ask_drink_name'
                session_state['last_intent'] = 'order_drink_asking_name'
                session_state['pending_intent_data'] = current_pending_data # Save any partial data (e.g. size if provided without name)
            elif not drink_size_known:
                text_response = f"What size would you like for your {current_pending_data['drink_name']}?"
                session_state['pending_question'] = 'ask_drink_size'
                session_state['last_intent'] = 'order_drink_asking_size'
                session_state['pending_intent_data'] = current_pending_data # Save drink_name
            else: # Both name and size are known
                text_response = f"Okay, one {current_pending_data['drink_size']} {current_pending_data['drink_name']} coming up!"
                session_state.pop('pending_question', None)
                session_state.pop('pending_intent_data', None)
                session_state['last_intent'] = 'order_drink_complete'

        elif nlu_result.intent == "goodbye":
            text_response = "Goodbye! Have a great day."
            session_state['last_intent'] = "goodbye"
            self.session_states.pop(session_id, None)
        elif nlu_result.intent == "error_no_dialogflow_client" or nlu_result.intent == "error_calling_dialogflow":
            text_response = "I'm having a little trouble with my understanding capabilities at the moment. Please try again later."
            session_state.pop('pending_question', None)
            session_state.pop('pending_intent_data', None)
            session_state['last_intent'] = nlu_result.intent
        elif not nlu_result.intent:
             text_response = "I'm not sure what you mean. Can you try rephrasing?"
             session_state['last_intent'] = "unknown_intent_or_empty"
             session_state.pop('pending_question', None)
             session_state.pop('pending_intent_data', None)


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
