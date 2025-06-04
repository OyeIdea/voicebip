import unittest
from unittest import mock
import grpc # For mocking context

# Import service and message types
from service import DialogueManagementServicer
import dialogue_management_service_pb2
import nlu_service_pb2 # Needed to construct NLUResponse for the DialogueRequest

class TestDialogueManagementServicer(unittest.TestCase):

    def setUp(self):
        """Setup method to create a servicer instance for each test."""
        self.servicer = DialogueManagementServicer()
        self.mock_context = mock.Mock(spec=grpc.ServicerContext)

    def test_manage_turn_greeting(self):
        """Test response for 'greeting' intent."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_greeting_dm_test",
            intent="greeting",
            intent_confidence=0.98
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_greeting_dm_test",
            nlu_result=nlu_res
        )

        response = self.servicer.ManageTurn(request, self.mock_context)

        self.assertEqual(response.text_response, "Hello there! How can I help you today?")
        self.assertEqual(response.session_id, "session_greeting_dm_test")

    def test_manage_turn_get_help(self):
        """Test response for 'get_help' intent."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_help_dm_test",
            intent="get_help",
            intent_confidence=0.92
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_help_dm_test",
            nlu_result=nlu_res
        )

        response = self.servicer.ManageTurn(request, self.mock_context)

        self.assertEqual(response.text_response, "I understand you need help. I'll do my best to assist you.")
        self.assertEqual(response.session_id, "session_help_dm_test")

    def test_manage_turn_get_weather_no_location(self):
        """Test 'get_weather' intent without a location entity."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_weather_no_loc_test",
            intent="get_weather",
            intent_confidence=0.89
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_weather_no_loc_test",
            nlu_result=nlu_res
        )

        response = self.servicer.ManageTurn(request, self.mock_context)

        expected_response_part = "I'm sorry, I can't fetch the actual weather for your area"
        self.assertIn(expected_response_part, response.text_response)
        self.assertEqual(response.session_id, "session_weather_no_loc_test")

    def test_manage_turn_get_weather_with_location(self):
        """Test 'get_weather' intent with a location entity."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_weather_with_loc_test",
            intent="get_weather",
            intent_confidence=0.91,
            entities=[nlu_service_pb2.Entity(name="location", value="London", confidence=0.85)]
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_weather_with_loc_test",
            nlu_result=nlu_res
        )

        response = self.servicer.ManageTurn(request, self.mock_context)

        expected_response_part = "I'm sorry, I can't fetch the actual weather for London"
        self.assertIn(expected_response_part, response.text_response)
        self.assertEqual(response.session_id, "session_weather_with_loc_test")

    def test_manage_turn_get_weather_with_location_and_date(self):
        """Test 'get_weather' with location and date entities."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_weather_loc_date_test",
            intent="get_weather",
            intent_confidence=0.93,
            entities=[
                nlu_service_pb2.Entity(name="location", value="Paris", confidence=0.88),
                nlu_service_pb2.Entity(name="date", value="tomorrow", confidence=0.80)
            ]
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_weather_loc_date_test",
            nlu_result=nlu_res
        )
        response = self.servicer.ManageTurn(request, self.mock_context)
        expected_response_part = "I'm sorry, I can't fetch the actual weather for Paris for tomorrow"
        self.assertIn(expected_response_part, response.text_response)

    def test_manage_turn_default_response_unknown_intent(self):
        """Test default response for an unknown intent."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_unknown_intent_test",
            intent="unknown_intent", # An intent not specifically handled
            intent_confidence=0.75
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_unknown_intent_test",
            nlu_result=nlu_res
        )

        response = self.servicer.ManageTurn(request, self.mock_context)

        self.assertEqual(response.text_response, "I'm sorry, I didn't quite understand that. Could you say it again?")
        self.assertEqual(response.session_id, "session_unknown_intent_test")

    def test_manage_turn_default_response_no_intent(self):
        """Test default response when NLU provides no intent."""
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id="session_no_intent_test",
            intent="", # Empty intent string
            intent_confidence=0.0 # No confidence or intent
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id="session_no_intent_test",
            nlu_result=nlu_res
        )
        response = self.servicer.ManageTurn(request, self.mock_context)
        self.assertEqual(response.text_response, "I'm not sure what you mean. Can you try rephrasing?")

    def test_manage_turn_session_passthrough(self):
        """Verify session_id from request is correctly passed to the DialogueResponse."""
        test_session_id = "custom_dm_session_passthrough_test"
        nlu_res = nlu_service_pb2.NLUResponse(
            session_id=test_session_id, # NLU response might also have session_id
            intent="greeting"
        )
        request = dialogue_management_service_pb2.DialogueRequest(
            session_id=test_session_id,
            nlu_result=nlu_res
        )

        response = self.servicer.ManageTurn(request, self.mock_context)

        self.assertEqual(response.session_id, test_session_id)

if __name__ == '__main__':
    unittest.main()
