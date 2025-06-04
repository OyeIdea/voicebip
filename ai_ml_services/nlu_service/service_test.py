import unittest
from unittest import mock
import grpc # For mocking context if needed for more advanced tests

# Assuming generated files and service module are accessible
import nlu_service_pb2
# nlu_service_pb2_grpc is not strictly needed for testing servicer logic directly,
# but good to have if we were testing client instantiation within these tests.
# import nlu_service_pb2_grpc

from service import NLUServiceServicer # The module under test

class TestNLUServiceServicer(unittest.TestCase):

    def setUp(self):
        """Setup method to create a servicer instance for each test."""
        self.servicer = NLUServiceServicer()
        # Mock the gRPC context object. For many simple servicers,
        # its methods might not be called, but it's good practice to mock it.
        self.mock_context = mock.Mock(spec=grpc.ServicerContext)

    def test_process_text_greeting(self):
        """Test basic greeting intent and entity."""
        request = nlu_service_pb2.NLURequest(
            text="hello there friend",
            session_id="session_greet_123"
        )
        response = self.servicer.ProcessText(request, self.mock_context)

        self.assertEqual(response.intent, "greeting")
        self.assertAlmostEqual(response.intent_confidence, 0.95)
        self.assertEqual(response.session_id, "session_greet_123")
        self.assertEqual(response.processed_text, request.text)

        # Check for default "User" entity
        found_user_entity = False
        for entity in response.entities:
            if entity.name == "name" and entity.value == "User":
                found_user_entity = True
                self.assertAlmostEqual(entity.confidence, 0.8)
                break
        self.assertTrue(found_user_entity, "Default 'User' entity not found for greeting.")

    def test_process_text_get_help(self):
        """Test 'get_help' intent and specific entity for help."""
        request = nlu_service_pb2.NLURequest(
            text="I really need some help with this thing.",
            session_id="session_help_456"
        )
        response = self.servicer.ProcessText(request, self.mock_context)

        self.assertEqual(response.intent, "get_help")
        # Default confidence for greeting is 0.95, this will be used as intent_confidence
        # The logic in service.py sets intent="get_help" but doesn't update intent_confidence for it.
        # This is a nuance of the current placeholder logic.
        self.assertAlmostEqual(response.intent_confidence, 0.95)
        self.assertEqual(response.session_id, "session_help_456")
        self.assertEqual(response.processed_text, request.text)

        found_topic_entity = False
        for entity in response.entities:
            if entity.name == "topic" and entity.value == "general":
                found_topic_entity = True
                self.assertAlmostEqual(entity.confidence, 0.7)
                break
        self.assertTrue(found_topic_entity, "'topic':'general' entity not found for get_help intent.")

        # Ensure default "User" entity is also present
        self.assertTrue(any(e.name == "name" and e.value == "User" for e in response.entities), "Default 'User' entity missing with get_help.")


    def test_process_text_get_weather(self):
        """Test 'get_weather' intent and related entities."""
        request = nlu_service_pb2.NLURequest(
            text="what is the weather like tomorrow",
            session_id="session_weather_789"
        )
        response = self.servicer.ProcessText(request, self.mock_context)

        self.assertEqual(response.intent, "get_weather")
        self.assertAlmostEqual(response.intent_confidence, 0.95) # Same as above, intent_confidence isn't updated for this rule
        self.assertEqual(response.session_id, "session_weather_789")

        expected_entities = {
            "query_type": ("general_weather", 0.85),
            "date": ("tomorrow", 0.9)
        }

        found_entities_count = 0
        for entity in response.entities:
            if entity.name in expected_entities:
                expected_value, expected_confidence = expected_entities[entity.name]
                if entity.value == expected_value:
                    self.assertAlmostEqual(entity.confidence, expected_confidence, msg=f"Confidence mismatch for entity {entity.name}")
                    found_entities_count +=1

        # Check if all expected entities (query_type, date) were found.
        # The default "User" entity will also be there.
        self.assertEqual(found_entities_count, len(expected_entities), f"Expected to find specific weather entities, found {found_entities_count}")
        self.assertTrue(any(e.name == "name" and e.value == "User" for e in response.entities), "Default 'User' entity missing with get_weather.")


    def test_process_text_session_passthrough(self):
        """Verify session_id from request is correctly passed to the response."""
        test_session_id = "custom_session_id_for_passthrough_test"
        request = nlu_service_pb2.NLURequest(
            text="Just a simple test.",
            session_id=test_session_id
        )
        response = self.servicer.ProcessText(request, self.mock_context)
        self.assertEqual(response.session_id, test_session_id)

    def test_process_text_default_case(self):
        """Test a case that doesn't match specific rules, expecting default greeting."""
        request = nlu_service_pb2.NLURequest(
            text="Tell me about your product.",
            session_id="session_default_case"
        )
        response = self.servicer.ProcessText(request, self.mock_context)

        self.assertEqual(response.intent, "greeting") # Default fallback intent
        self.assertAlmostEqual(response.intent_confidence, 0.95)
        self.assertEqual(response.session_id, "session_default_case")
        # Check for only the default "User" entity
        self.assertEqual(len(response.entities), 1)
        self.assertEqual(response.entities[0].name, "name")
        self.assertEqual(response.entities[0].value, "User")

if __name__ == '__main__':
    unittest.main()
