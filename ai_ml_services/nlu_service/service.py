# ai_ml_services/nlu_service/service.py

import grpc
from concurrent import futures
import time

# Import generated protobuf and gRPC modules for NLU Service
import nlu_service_pb2
import nlu_service_pb2_grpc

# Import generated protobuf and gRPC modules for Dialogue Management Service (client)
import dialogue_management_service_pb2
import dialogue_management_service_pb2_grpc

# Google Cloud Dialogflow CX specific imports
from google.cloud import dialogflowcx_v3beta1 as dialogflowcx
# from google.protobuf import struct_pb2 # For complex parameters, if needed later

# Import configuration
from .config import (
    DIALOGFLOW_PROJECT_ID,
    DIALOGFLOW_AGENT_ID,
    DIALOGFLOW_LOCATION_ID,
    DIALOGFLOW_LANGUAGE_CODE,
    GOOGLE_APP_CREDS # Used here for an initial check/warning
)

class NLUServiceServicer(nlu_service_pb2_grpc.NLUServiceServicer):
    """
    Implements the NLUService gRPC interface using Dialogflow CX.
    After processing text with Dialogflow CX, it calls the DialogueManagementService.
    """
    def __init__(self):
        self.sessions_client = None
        if GOOGLE_APP_CREDS and DIALOGFLOW_PROJECT_ID and DIALOGFLOW_AGENT_ID:
            try:
                client_options = None
                if DIALOGFLOW_LOCATION_ID and DIALOGFLOW_LOCATION_ID.lower() != "global":
                    client_options = {"api_endpoint": f"{DIALOGFLOW_LOCATION_ID}-dialogflow.googleapis.com"}

                self.sessions_client = dialogflowcx.SessionsClient(client_options=client_options)
                print("NLUService: Dialogflow CX SessionsClient initialized successfully.")
            except Exception as e:
                print(f"NLUService Error: Failed to initialize Dialogflow CX SessionsClient: {e}")
                self.sessions_client = None # Ensure it's None if init fails
        else:
            print("NLUService Warning: Dialogflow CX client not initialized due to missing configuration (PROJECT_ID, AGENT_ID, or GOOGLE_APPLICATION_CREDENTIALS). NLUService will use placeholder logic or fail.")

        # Configuration for Dialogue Management service endpoint
        self.dm_service_address = 'localhost:50054' # Make this configurable
        # Consider creating the channel once if NLUServiceServicer is long-lived,
        # or per-call if requests are infrequent or to handle transient DM issues.
        # For now, per-call via 'with' statement is robust.

    def _call_dialogflow_cx(self, request_text: str, request_session_id: str) -> nlu_service_pb2.NLUResponse:
        """
        Helper method to call Dialogflow CX and map its response.
        """
        if not self.sessions_client:
            print("NLUService Error: Dialogflow CX client not available. Returning error response.")
            return nlu_service_pb2.NLUResponse(
                session_id=request_session_id,
                intent="error_no_dialogflow_client",
                processed_text=request_text,
                intent_confidence=0.0
            )

        session_path = self.sessions_client.session_path(
            project=DIALOGFLOW_PROJECT_ID,
            location=DIALOGFLOW_LOCATION_ID,
            agent=DIALOGFLOW_AGENT_ID,
            session=request_session_id # Using the incoming session_id for Dialogflow session
        )

        text_input = dialogflowcx.TextInput(text=request_text)
        query_input = dialogflowcx.QueryInput(
            text=text_input,
            language_code=DIALOGFLOW_LANGUAGE_CODE
        )

        try:
            df_request = dialogflowcx.DetectIntentRequest(session=session_path, query_input=query_input)
            print(f"NLUService: Sending request to Dialogflow CX for session {session_path}: Text='{request_text}'")
            df_response = self.sessions_client.detect_intent(request=df_request)
            query_result = df_response.query_result

            print(f"NLUService: Received Dialogflow CX response: Intent='{query_result.intent.display_name if query_result.intent else 'N/A'}', Confidence={query_result.intent_detection_confidence:.2f}")

            nlu_entities = []
            if query_result.parameters and query_result.parameters.fields:
                for entity_name, entity_value_struct in query_result.parameters.fields.items():
                    entity_value_str = ""
                    kind = entity_value_struct.WhichOneof('kind') # 'string_value', 'number_value', 'bool_value', 'struct_value', 'list_value', 'null_value'

                    if kind == 'string_value':
                        entity_value_str = entity_value_struct.string_value
                    elif kind == 'number_value':
                        entity_value_str = str(entity_value_struct.number_value)
                    elif kind == 'bool_value':
                        entity_value_str = str(entity_value_struct.bool_value).lower()
                    elif kind == 'struct_value' or kind == 'list_value':
                         # For complex types, serialize to JSON string or handle specific structures
                        from google.protobuf.json_format import MessageToJson
                        entity_value_str = MessageToJson(entity_value_struct)
                    elif kind == 'null_value':
                        entity_value_str = "null"
                    else: # Parameter has a value, but its kind is not one of the simple ones above
                        entity_value_str = "[Complex Value]"


                    nlu_entities.append(nlu_service_pb2.Entity(
                        name=entity_name,
                        value=entity_value_str,
                        confidence=1.0 # DF CX parameters usually don't have per-param confidence. Intent confidence is key.
                    ))

            return nlu_service_pb2.NLUResponse(
                session_id=request_session_id,
                intent=query_result.intent.display_name if query_result.intent else "no_intent_matched",
                entities=nlu_entities,
                processed_text=query_result.text if query_result.text else request_text,
                intent_confidence=query_result.intent_detection_confidence if query_result.intent else 0.0
            )

        except Exception as e:
            print(f"NLUService: Dialogflow API error for session {request_session_id}: {e}")
            return nlu_service_pb2.NLUResponse(
                session_id=request_session_id,
                intent="error_calling_dialogflow",
                processed_text=request_text,
                intent_confidence=0.0,
                entities=[nlu_service_pb2.Entity(name="error_message", value=str(e))]
            )


    def ProcessText(self, request: nlu_service_pb2.NLURequest, context):
        print(f"NLUService: Received ProcessText request for SID '{request.session_id}', Text: '{request.text}'")

        # Use Dialogflow CX for NLU processing
        nlu_response = self._call_dialogflow_cx(request.text, request.session_id)

        entities_str = ', '.join([f"{e.name}='{e.value}' (conf: {e.confidence:.2f})" for e in nlu_response.entities])
        print(f"NLUService: NLU Response from Dialogflow for SID {nlu_response.session_id}: Intent='{nlu_response.intent}' (Conf: {nlu_response.intent_confidence:.2f}), Entities=[{entities_str}]")

        # Call DialogueManagementService
        try:
            with grpc.insecure_channel(self.dm_service_address) as channel:
                dm_stub = dialogue_management_service_pb2_grpc.DialogueManagementServiceStub(channel)
                dialogue_request = dialogue_management_service_pb2.DialogueRequest(
                    session_id=nlu_response.session_id,
                    nlu_result=nlu_response
                )
                print(f"NLUService: Calling DMService at {self.dm_service_address} for SID {dialogue_request.session_id}")
                dm_response = dm_stub.ManageTurn(dialogue_request, timeout=10)
                if dm_response:
                    print(f"NLUService: Received DM response for SID {dm_response.session_id}: TextResponse='{dm_response.text_response}'")
                else:
                    print(f"NLUService: No response from DMService for SID {dialogue_request.session_id}")
        except grpc.RpcError as e:
            print(f"NLUService: Error calling DMService for SID {request.session_id}: Code={e.code()}, Details='{e.details()}'")
        except Exception as e:
            print(f"NLUService: Unexpected Python error calling DMService for SID {request.session_id}: {e}")

        return nlu_response # Return the NLU response obtained from Dialogflow

def serve():
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
