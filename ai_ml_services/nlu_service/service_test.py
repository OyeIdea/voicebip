import unittest
from unittest import mock
import grpc # For mocking context and RpcError

# Assuming generated files and service module are accessible
import nlu_service_pb2
import dialogue_management_service_pb2 # For DM Response
import dialogue_management_service_pb2_grpc # For DM Stub spec

from service import NLUServiceServicer # The module under test

class TestNLUServiceServicer(unittest.TestCase):

    def setUp(self):
        """Setup method to create a servicer instance for each test."""
        self.servicer = NLUServiceServicer()
        self.mock_context = mock.Mock(spec=grpc.ServicerContext)

    # Keep existing tests for NLU logic (greeting, get_help, etc.)
    # These will now also implicitly test that the DM call is made,
    # but we'll add specific assertions for the DM call.

    @mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_process_text_greeting_calls_dm(self, mock_grpc_channel_constructor, MockDMServiceStubConstructor):
        """Test basic greeting intent and that DM service is called."""

        # Configure DM mocks
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance # For 'with' statement
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_dm_stub_instance = mock.Mock(spec=dialogue_management_service_pb2_grpc.DialogueManagementServiceStub)
        MockDMServiceStubConstructor.return_value = mock_dm_stub_instance
        mock_dm_stub_instance.ManageTurn.return_value = dialogue_management_service_pb2.DialogueResponse(
            session_id="session_greet_123",
            text_response="DM says hello!"
        )

        request = nlu_service_pb2.NLURequest(
            text="hello there friend",
            session_id="session_greet_123"
        )

        # Call the method under test
        nlu_response = self.servicer.ProcessText(request, self.mock_context)

        # Assert NLU's own response
        self.assertEqual(nlu_response.intent, "greeting")
        self.assertAlmostEqual(nlu_response.intent_confidence, 0.95)
        self.assertEqual(nlu_response.session_id, "session_greet_123")

        # Assert DM call
        mock_grpc_channel_constructor.assert_called_once_with(self.servicer.dm_service_address)
        MockDMServiceStubConstructor.assert_called_once_with(mock_channel_instance)
        mock_dm_stub_instance.ManageTurn.assert_called_once()

        called_dm_request = mock_dm_stub_instance.ManageTurn.call_args[0][0]
        self.assertIsInstance(called_dm_request, dialogue_management_service_pb2.DialogueRequest)
        self.assertEqual(called_dm_request.session_id, nlu_response.session_id)
        self.assertEqual(called_dm_request.nlu_result, nlu_response) # Check if the NLU response was passed

    @mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_process_text_get_help_calls_dm(self, mock_grpc_channel_constructor, MockDMServiceStubConstructor):
        """Test 'get_help' intent and that DM service is called."""
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_dm_stub_instance = mock.Mock(spec=dialogue_management_service_pb2_grpc.DialogueManagementServiceStub)
        MockDMServiceStubConstructor.return_value = mock_dm_stub_instance
        mock_dm_stub_instance.ManageTurn.return_value = dialogue_management_service_pb2.DialogueResponse(
            session_id="session_help_456",
            text_response="DM will help."
        )

        request = nlu_service_pb2.NLURequest(
            text="I really need some help with this thing.",
            session_id="session_help_456"
        )
        nlu_response = self.servicer.ProcessText(request, self.mock_context)

        self.assertEqual(nlu_response.intent, "get_help")

        mock_dm_stub_instance.ManageTurn.assert_called_once()
        called_dm_request = mock_dm_stub_instance.ManageTurn.call_args[0][0]
        self.assertEqual(called_dm_request.nlu_result, nlu_response)


    @mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_process_text_handles_dm_rpc_error(self, mock_grpc_channel_constructor, MockDMServiceStubConstructor):
        """Test that NLU service handles RpcError from DM gracefully."""
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_dm_stub_instance = mock.Mock(spec=dialogue_management_service_pb2_grpc.DialogueManagementServiceStub)
        MockDMServiceStubConstructor.return_value = mock_dm_stub_instance

        simulated_rpc_error = grpc.RpcError("DM unavailable")
        # If specific code/details needed:
        # simulated_rpc_error.code = lambda: grpc.StatusCode.UNAVAILABLE
        # simulated_rpc_error.details = lambda: "DM service is down"
        mock_dm_stub_instance.ManageTurn.side_effect = simulated_rpc_error

        request = nlu_service_pb2.NLURequest(
            text="Any text for error test",
            session_id="session_dm_error_test"
        )

        with mock.patch('builtins.print') as mock_print:
            nlu_response = self.servicer.ProcessText(request, self.mock_context)

            # NLU should still return its own response
            self.assertIsNotNone(nlu_response)
            self.assertEqual(nlu_response.session_id, request.session_id)
            self.assertEqual(nlu_response.intent, "greeting") # Default intent in this case

            # Check that DM call was attempted
            mock_dm_stub_instance.ManageTurn.assert_called_once()

            # Check that the error was logged
            error_logged = False
            for call_arg in mock_print.call_args_list:
                if "Error calling DialogueManagementService" in str(call_arg) and "DM unavailable" in str(call_arg):
                    error_logged = True
                    break
            self.assertTrue(error_logged, "Expected DM RpcError to be logged by NLU service")

    # --- Keep original NLU logic tests, slightly adapted if needed or rely on above tests for DM call checks ---

    def test_original_process_text_greeting(self): # Renamed to avoid conflict if running all tests
        """Original test for basic greeting intent and entity, DM call will happen."""
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub'):
            request = nlu_service_pb2.NLURequest(text="hello", session_id="s1")
            response = self.servicer.ProcessText(request, self.mock_context)
            self.assertEqual(response.intent, "greeting")

    def test_original_process_text_get_help(self):
        """Original test for 'get_help' intent, DM call will happen."""
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub'):
            request = nlu_service_pb2.NLURequest(text="I need help", session_id="s2")
            response = self.servicer.ProcessText(request, self.mock_context)
            self.assertEqual(response.intent, "get_help")
            self.assertTrue(any(e.name == "topic" and e.value == "general" for e in response.entities))

    def test_original_process_text_get_weather(self):
        """Original test for 'get_weather' intent, DM call will happen."""
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub'):
            request = nlu_service_pb2.NLURequest(text="weather tomorrow", session_id="s3")
            response = self.servicer.ProcessText(request, self.mock_context)
            self.assertEqual(response.intent, "get_weather")
            self.assertTrue(any(e.name == "date" and e.value == "tomorrow" for e in response.entities))

    def test_original_process_text_session_passthrough(self):
        """Original test for session_id passthrough, DM call will happen."""
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub'):
            test_session_id = "custom_session_passthrough"
            request = nlu_service_pb2.NLURequest(text="Test", session_id=test_session_id)
            response = self.servicer.ProcessText(request, self.mock_context)
            self.assertEqual(response.session_id, test_session_id)

    def test_original_process_text_default_case(self):
        """Original test for default case, DM call will happen."""
        with mock.patch('service.grpc.insecure_channel'), \
             mock.patch('service.dialogue_management_service_pb2_grpc.DialogueManagementServiceStub'):
            request = nlu_service_pb2.NLURequest(text="Product info", session_id="s_default")
            response = self.servicer.ProcessText(request, self.mock_context)
            self.assertEqual(response.intent, "greeting")


if __name__ == '__main__':
    unittest.main()
