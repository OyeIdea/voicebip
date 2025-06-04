import unittest
from unittest import mock
import grpc

# Import NLU service and message types
from service import NLUServiceServicer
import nlu_service_pb2

# Import DM service message types (for mocking DM calls)
import dialogue_management_service_pb2
import dialogue_management_service_pb2_grpc

# Import Dialogflow CX specific types for creating mock responses
from google.cloud.dialogflowcx_v3beta1 import types as dialogflowcx_types
from google.protobuf import struct_pb2

# To patch config values loaded at module level in service.py
# We need to patch them where they are looked up (i.e., in the 'service' module)
@mock.patch.dict('service.__dict__', {
    'DIALOGFLOW_PROJECT_ID': 'test-project',
    'DIALOGFLOW_AGENT_ID': 'test-agent',
    'DIALOGFLOW_LOCATION_ID': 'global',
    'DIALOGFLOW_LANGUAGE_CODE': 'en-US',
    'GOOGLE_APP_CREDS': 'fake_creds_path.json' # Needs to be non-empty for client init
})
class TestNLUServiceServicerWithDialogflow(unittest.TestCase):

    def setUp(self):
        # Mock the Dialogflow SessionsClient constructor and its instance
        self.sessions_client_patcher = mock.patch('service.dialogflowcx.SessionsClient')
        self.MockDialogflowSessionsClientConstructor = self.sessions_client_patcher.start()
        self.mock_df_sessions_client_instance = self.MockDialogflowSessionsClientConstructor.return_value
        self.mock_df_sessions_client_instance.detect_intent = mock.Mock()

        # Mock the Dialogue Management (DM) client stub
        self.dm_channel_patcher = mock.patch('service.grpc.insecure_channel')
        self.mock_grpc_insecure_channel = self.dm_channel_patcher.start()

        self.dm_stub_patcher = mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub')
        self.MockDMServiceStubConstructor = self.dm_stub_patcher.start()

        self.mock_dm_channel_instance = mock.Mock()
        self.mock_dm_channel_instance.__enter__.return_value = self.mock_dm_channel_instance
        self.mock_grpc_insecure_channel.return_value = self.mock_dm_channel_instance

        self.mock_dm_stub_instance = self.MockDMServiceStubConstructor.return_value
        self.mock_dm_stub_instance.ManageTurn.return_value = dialogue_management_service_pb2.DialogueResponse(
            text_response="DM mock response"
        )

        # Instantiate the servicer. This will use the mocked SessionsClient.
        self.servicer = NLUServiceServicer()
        self.mock_grpc_context = mock.Mock(spec=grpc.ServicerContext)

    def tearDown(self):
        self.sessions_client_patcher.stop()
        self.dm_stub_patcher.stop()
        self.dm_channel_patcher.stop()

    def test_process_text_success_maps_dialogflow_response(self):
        # 1. Define a mock Dialogflow CX response
        mock_df_query_result = dialogflowcx_types.QueryResult()
        mock_df_query_result.intent.display_name = "order_pizza"
        mock_df_query_result.intent_detection_confidence = 0.85
        mock_df_query_result.text = "I want a large pepperoni pizza"

        # Parameters: string, number, boolean, struct, list
        params = struct_pb2.Struct()
        params.fields['size'].string_value = "large"
        params.fields['quantity'].number_value = 1
        params.fields['is_urgent'].bool_value = False

        topping_struct = struct_pb2.Struct()
        topping_struct.fields['name'].string_value = "pepperoni"
        topping_struct.fields['type'].string_value = "meat"
        params.fields['topping_details'].struct_value.CopyFrom(topping_struct)

        extra_list = struct_pb2.ListValue()
        extra_list.values.add().string_value = "extra cheese"
        extra_list.values.add().string_value = "onions"
        params.fields['extras'].list_value.CopyFrom(extra_list)

        mock_df_query_result.parameters = params

        mock_df_response = dialogflowcx_types.DetectIntentResponse(query_result=mock_df_query_result)
        self.mock_df_sessions_client_instance.detect_intent.return_value = mock_df_response

        # 2. Create an NLURequest
        nlu_request = nlu_service_pb2.NLURequest(
            text="I want a large pepperoni pizza with extra cheese and onions, make it snappy!",
            session_id="test_df_session_1"
        )

        # 3. Call servicer.ProcessText
        nlu_response = self.servicer.ProcessText(nlu_request, self.mock_grpc_context)

        # 4. Assert Dialogflow call
        self.mock_df_sessions_client_instance.detect_intent.assert_called_once()
        called_df_request = self.mock_df_sessions_client_instance.detect_intent.call_args[1]['request'] # request is a kwarg
        self.assertIn(f"projects/test-project/locations/global/agents/test-agent/sessions/{nlu_request.session_id}", called_df_request.session)
        self.assertEqual(called_df_request.query_input.text.text, nlu_request.text)
        self.assertEqual(called_df_request.query_input.language_code, 'en-US')

        # 5. Assert NLUResponse mapping
        self.assertEqual(nlu_response.session_id, nlu_request.session_id)
        self.assertEqual(nlu_response.intent, "order_pizza")
        self.assertAlmostEqual(nlu_response.intent_confidence, 0.85)
        self.assertEqual(nlu_response.processed_text, "I want a large pepperoni pizza") # From df_response.query_result.text

        expected_entities = {
            "size": "large",
            "quantity": "1.0", # str(1.0)
            "is_urgent": "false", # str(False).lower()
            "topping_details": '{\n  "name": "pepperoni",\n  "type": "meat"\n}', # JSON of struct
            "extras": '[\n  "extra cheese",\n  "onions"\n]' # JSON of list
        }
        self.assertEqual(len(nlu_response.entities), len(expected_entities))
        for entity in nlu_response.entities:
            self.assertIn(entity.name, expected_entities)
            # For struct/list, MessageToJson adds newlines and spaces, so direct compare might be tricky
            # if not exactly matching the format. For now, let's assume it's okay or use JSON parsing for comparison.
            if entity.name == "topping_details" or entity.name == "extras":
                 import json
                 self.assertEqual(json.loads(entity.value), json.loads(expected_entities[entity.name]))
            else:
                self.assertEqual(entity.value, expected_entities[entity.name])
            self.assertAlmostEqual(entity.confidence, 1.0) # Default for DF params

        # 6. Assert DM call
        self.mock_dm_stub_instance.ManageTurn.assert_called_once()
        called_dm_request = self.mock_dm_stub_instance.ManageTurn.call_args[0][0]
        self.assertEqual(called_dm_request.nlu_result, nlu_response)


    def test_process_text_dialogflow_api_error(self):
        self.mock_df_sessions_client_instance.detect_intent.side_effect = Exception("Dialogflow API Error")
        nlu_request = nlu_service_pb2.NLURequest(text="test error", session_id="s_api_error")

        with mock.patch('builtins.print') as mock_print:
            nlu_response = self.servicer.ProcessText(nlu_request, self.mock_grpc_context)

        self.assertEqual(nlu_response.intent, "error_calling_dialogflow")
        self.assertEqual(nlu_response.entities[0].name, "error_message")
        self.assertIn("Dialogflow API Error", nlu_response.entities[0].value)

        self.mock_dm_stub_instance.ManageTurn.assert_called_once_with(
            dialogue_management_service_pb2.DialogueRequest(session_id="s_api_error", nlu_result=nlu_response)
        )
        mock_print.assert_any_call("NLUService: Dialogflow API error for session s_api_error: Dialogflow API Error")


    def test_process_text_no_intent_matched(self):
        mock_df_query_result = dialogflowcx_types.QueryResult(text="unrecognized input")
        # No intent set in mock_df_query_result.intent
        mock_df_response = dialogflowcx_types.DetectIntentResponse(query_result=mock_df_query_result)
        self.mock_df_sessions_client_instance.detect_intent.return_value = mock_df_response

        nlu_request = nlu_service_pb2.NLURequest(text="unrecognized input", session_id="s_no_intent")
        nlu_response = self.servicer.ProcessText(nlu_request, self.mock_grpc_context)

        self.assertEqual(nlu_response.intent, "no_intent_matched")
        self.assertEqual(nlu_response.intent_confidence, 0.0)


    def test_process_text_no_parameters(self):
        mock_df_query_result = dialogflowcx_types.QueryResult(
            text="text with no params",
            intent=dialogflowcx_types.Intent(display_name="simple_intent")
        )
        # No parameters set in mock_df_query_result.parameters
        mock_df_response = dialogflowcx_types.DetectIntentResponse(query_result=mock_df_query_result)
        self.mock_df_sessions_client_instance.detect_intent.return_value = mock_df_response

        nlu_request = nlu_service_pb2.NLURequest(text="text with no params", session_id="s_no_params")
        nlu_response = self.servicer.ProcessText(nlu_request, self.mock_grpc_context)

        self.assertEqual(len(nlu_response.entities), 0)
        self.assertEqual(nlu_response.intent, "simple_intent")


    @mock.patch.dict('service.__dict__', {'GOOGLE_APP_CREDS': None}) # Override class-level patch
    def test_process_text_dialogflow_client_not_initialized(self, MockDialogflowSessionsClient):
        # Re-initialize servicer with GOOGLE_APP_CREDS as None to test client non-initialization
        # This requires SessionsClient to be passed to NLUServiceServicer or for __init__ to re-check config
        # The current __init__ checks global config vars.
        servicer_no_client = NLUServiceServicer() # This will print warning and set sessions_client to None

        nlu_request = nlu_service_pb2.NLURequest(text="test no client", session_id="s_no_client")

        with mock.patch('builtins.print') as mock_print: # To check for the warning
            nlu_response = servicer_no_client.ProcessText(nlu_request, self.mock_grpc_context)

        self.assertEqual(nlu_response.intent, "error_no_dialogflow_client")
        # Check that the DM call is still made with the error NLUResponse
        self.mock_dm_stub_instance.ManageTurn.assert_called_once()
        called_dm_request = self.mock_dm_stub_instance.ManageTurn.call_args[0][0]
        self.assertEqual(called_dm_request.nlu_result.intent, "error_no_dialogflow_client")

        # Check if the warning from __init__ was printed (if it's the first time for this config)
        # Or check the print from _call_dialogflow_cx
        mock_print.assert_any_call("NLUService Error: Dialogflow CX client not available. Returning error response.")


if __name__ == '__main__':
    unittest.main()
