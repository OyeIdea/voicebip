import unittest
from unittest import mock
import grpc # For mocking context and RpcError

# Import service and message types
from service import DialogueManagementServicer
import dialogue_management_service_pb2
import nlu_service_pb2 # Needed to construct NLUResponse for the DialogueRequest
import tts_service_pb2 # For TTS Response
import tts_service_pb2_grpc # For TTS Stub spec

class TestDialogueManagementServicer(unittest.TestCase):

    def setUp(self):
        """Setup method to create a servicer instance for each test."""
        self.servicer = DialogueManagementServicer()
        self.mock_context = mock.Mock(spec=grpc.ServicerContext)

    # --- Existing tests for DM logic can be kept ---
    # They will now also implicitly test that TTS call is made,
    # but we'll add specific assertions for the TTS call in new/adapted tests.

    @mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_manage_turn_greeting_calls_tts(self, mock_grpc_channel_constructor, MockTTSServiceStubConstructor):
        """Test 'greeting' intent response and that TTS service is called."""

        # Configure TTS mocks
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance # For 'with' statement
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_tts_stub_instance = mock.Mock(spec=tts_service_pb2_grpc.TextToSpeechServiceStub)
        MockTTSServiceStubConstructor.return_value = mock_tts_stub_instance
        mock_tts_stub_instance.SynthesizeText.return_value = tts_service_pb2.TTSResponse(
            session_id="session_greeting_dm_test",
            status_message="TTS says hello!"
        )

        # Prepare DM request
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_greeting_dm_test",
            intent="greeting",
            intent_confidence=0.98
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_greeting_dm_test",
            nlu_result=nlu_res
        )

        # Call the method under test
        dm_response = self.servicer.ManageTurn(request, self.mock_context)

        # Assert DM's own response
        expected_dm_text_response = "Hello there! How can I help you today?"
        self.assertEqual(dm_response.text_response, expected_dm_text_response)
        self.assertEqual(dm_response.session_id, "session_greeting_dm_test")

        # Assert TTS call
        mock_grpc_channel_constructor.assert_called_once_with(self.servicer.tts_service_address)
        MockTTSServiceStubConstructor.assert_called_once_with(mock_channel_instance)
        mock_tts_stub_instance.SynthesizeText.assert_called_once()

        called_tts_request = mock_tts_stub_instance.SynthesizeText.call_args[0][0]
        self.assertIsInstance(called_tts_request, tts_service_pb2.TTSRequest)
        self.assertEqual(called_tts_request.text_to_synthesize, expected_dm_text_response)
        self.assertEqual(called_tts_request.session_id, request.session_id)
        self.assertEqual(called_tts_request.voice_config_id, "default_voice")


    @mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_manage_turn_get_weather_calls_tts(self, mock_grpc_channel_constructor, MockTTSServiceStubConstructor):
        """Test 'get_weather' intent and that TTS service is called with the correct text."""
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_tts_stub_instance = mock.Mock(spec=tts_service_pb2_grpc.TextToSpeechServiceStub)
        MockTTSServiceStubConstructor.return_value = mock_tts_stub_instance
        mock_tts_stub_instance.SynthesizeText.return_value = tts_service_pb2.TTSResponse(
            status_message="TTS processed weather response"
        )

        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_weather_tts_test",
            intent="get_weather",
            entities=[nlu_service_pb2.Entity(name="location", value="London", confidence=0.85)]
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_weather_tts_test",
            nlu_result=nlu_res
        )

        dm_response = self.servicer.ManageTurn(request, self.mock_context)

        expected_dm_text_response_part = "I'm sorry, I can't fetch the actual weather for London"
        self.assertIn(expected_dm_text_response_part, dm_response.text_response) # DM response check

        mock_tts_stub_instance.SynthesizeText.assert_called_once()
        called_tts_request = mock_tts_stub_instance.SynthesizeText.call_args[0][0]
        self.assertEqual(called_tts_request.text_to_synthesize, dm_response.text_response) # Ensure DM's text is sent to TTS


    @mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_manage_turn_handles_tts_rpc_error(self, mock_grpc_channel_constructor, MockTTSServiceStubConstructor):
        """Test that DM service handles RpcError from TTS gracefully."""
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_tts_stub_instance = mock.Mock(spec=tts_service_pb2_grpc.TextToSpeechServiceStub)
        MockTTSServiceStubConstructor.return_value = mock_tts_stub_instance

        simulated_rpc_error = grpc.RpcError("TTS unavailable")
        # If specific code/details needed:
        # simulated_rpc_error.code = lambda: grpc.StatusCode.UNAVAILABLE
        # simulated_rpc_error.details = lambda: "TTS service is down"
        mock_tts_stub_instance.SynthesizeText.side_effect = simulated_rpc_error

        nlu_res = nlu_service_pb2.NLUResponse(intent="greeting", session_id="session_tts_error_test")
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_tts_error_test",
            nlu_result=nlu_res
        )

        with mock.patch('builtins.print') as mock_print:
            dm_response = self.servicer.ManageTurn(request, self.mock_context)

            # DM should still return its own determined response
            self.assertIsNotNone(dm_response)
            self.assertEqual(dm_response.session_id, request.session_id)
            self.assertEqual(dm_response.text_response, "Hello there! How can I help you today?")

            # Check that TTS call was attempted
            mock_tts_stub_instance.SynthesizeText.assert_called_once()

            # Check that the error was logged
            error_logged = False
            for call_arg in mock_print.call_args_list:
                if "Error calling TextToSpeechService" in str(call_arg) and "TTS unavailable" in str(call_arg):
                    error_logged = True
                    break
            self.assertTrue(error_logged, "Expected TTS RpcError to be logged by DM service")

    # --- Keep original DM logic tests ---
    # These tests will now also ensure that the TTS call is made,
    # but the primary assertion remains on the DM's text_response.
    # We can add a generic mock for TTS to prevent actual calls if these are run standalone.

    def test_original_manage_turn_get_help(self):
        """Original test for 'get_help' intent response. TTS call will be made."""
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub'):
            nlu_res = nlu_service_pb2.NLUResponse(intent="get_help")
            request = dialogue_management_service_pb2.DialogueRequest(session_id="s_help", nlu_result=nlu_res)
            response = self.servicer.ManageTurn(request, self.mock_context)
            self.assertEqual(response.text_response, "I understand you need help. I'll do my best to assist you.")

    def test_original_manage_turn_get_weather_no_location(self):
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub'):
            nlu_res = nlu_service_pb2.NLUResponse(intent="get_weather")
            request = dialogue_management_service_pb2.DialogueRequest(session_id="s_weather_no_loc", nlu_result=nlu_res)
            response = self.servicer.ManageTurn(request, self.mock_context)
            self.assertIn("weather for your area", response.text_response)

    def test_original_manage_turn_default_response_unknown_intent(self):
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub'):
            nlu_res = nlu_service_pb2.NLUResponse(intent="unknown")
            request = dialogue_management_service_pb2.DialogueRequest(session_id="s_unknown", nlu_result=nlu_res)
            response = self.servicer.ManageTurn(request, self.mock_context)
            self.assertEqual(response.text_response, "I'm sorry, I didn't quite understand that. Could you say it again?")

    def test_original_manage_turn_session_passthrough(self):
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.tts_service_pb2_grpc.TextToSpeechServiceStub'):
            test_session_id = "custom_passthrough"
            nlu_res = nlu_service_pb2.NLUResponse(session_id=test_session_id, intent="greeting")
            request = dialogue_management_service_pb2.DialogueRequest(session_id=test_session_id, nlu_result=nlu_res)
            response = self.servicer.ManageTurn(request, self.mock_context)
            self.assertEqual(response.session_id, test_session_id)


if __name__ == '__main__':
    unittest.main()
