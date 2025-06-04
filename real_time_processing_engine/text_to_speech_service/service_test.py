import unittest
from unittest import mock
import grpc # For mocking context

# Import service and message types
from service import TextToSpeechServicer # The module under test
import tts_service_pb2

class TestTextToSpeechServicer(unittest.TestCase):

    def setUp(self):
        """Setup method to create a servicer instance for each test."""
        self.servicer = TextToSpeechServicer()
        self.mock_context = mock.Mock(spec=grpc.ServicerContext)

    def test_synthesize_text_placeholder_response_default_voice(self):
        """Test placeholder response with default voice_config_id."""
        session_id = "test_tts_session_default_voice"
        text_to_synthesize = "Hello world, this is a test."

        request = tts_service_pb2.TTSRequest(
            text_to_synthesize=text_to_synthesize,
            session_id=session_id,
            voice_config_id="default_voice" # Explicitly using a common default
        )

        response = self.servicer.SynthesizeText(request, self.mock_context)

        expected_status_message = f"Text for session '{session_id}' (voice: 'default_voice') received by TTS. Placeholder synthesis initiated."
        self.assertEqual(response.session_id, session_id)
        self.assertEqual(response.status_message, expected_status_message)

    def test_synthesize_text_placeholder_response_custom_voice(self):
        """Test placeholder response with a custom voice_config_id."""
        session_id = "test_tts_session_custom_voice"
        text_to_synthesize = "Another test with a different voice."
        voice_config = "custom_voice_abc_123"

        request = tts_service_pb2.TTSRequest(
            text_to_synthesize=text_to_synthesize,
            session_id=session_id,
            voice_config_id=voice_config
        )

        response = self.servicer.SynthesizeText(request, self.mock_context)

        expected_status_message = f"Text for session '{session_id}' (voice: '{voice_config}') received by TTS. Placeholder synthesis initiated."
        self.assertEqual(response.session_id, session_id)
        self.assertEqual(response.status_message, expected_status_message)

    def test_synthesize_text_placeholder_response_empty_voice_config(self):
        """Test placeholder response when voice_config_id is empty."""
        session_id = "test_tts_session_empty_voice"
        text_to_synthesize = "Testing with no specific voice configuration."

        request = tts_service_pb2.TTSRequest(
            text_to_synthesize=text_to_synthesize,
            session_id=session_id,
            voice_config_id="" # Empty voice_config_id
        )

        response = self.servicer.SynthesizeText(request, self.mock_context)

        # The service logic uses 'default' if voice_config_id is empty
        expected_status_message = f"Text for session '{session_id}' (voice: 'default') received by TTS. Placeholder synthesis initiated."
        self.assertEqual(response.session_id, session_id)
        self.assertEqual(response.status_message, expected_status_message)

    def test_synthesize_text_different_texts_and_sessions(self):
        """Test with a couple of different inputs to ensure no state leakage (though servicer is stateless)."""
        session_id_1 = "tts_session_001"
        text_1 = "First message."
        request_1 = tts_service_pb2.TTSRequest(text_to_synthesize=text_1, session_id=session_id_1, voice_config_id="voice_a")
        response_1 = self.servicer.SynthesizeText(request_1, self.mock_context)
        expected_status_1 = f"Text for session '{session_id_1}' (voice: 'voice_a') received by TTS. Placeholder synthesis initiated."
        self.assertEqual(response_1.session_id, session_id_1)
        self.assertEqual(response_1.status_message, expected_status_1)

        session_id_2 = "tts_session_002"
        text_2 = "Second message, quite different."
        request_2 = tts_service_pb2.TTSRequest(text_to_synthesize=text_2, session_id=session_id_2, voice_config_id="voice_b")
        response_2 = self.servicer.SynthesizeText(request_2, self.mock_context)
        expected_status_2 = f"Text for session '{session_id_2}' (voice: 'voice_b') received by TTS. Placeholder synthesis initiated."
        self.assertEqual(response_2.session_id, session_id_2)
        self.assertEqual(response_2.status_message, expected_status_2)


if __name__ == '__main__':
    unittest.main()
