import unittest
from unittest import mock
import grpc # For mock_context and RpcError

# Assuming the generated files are in the same directory or accessible via PYTHONPATH
import audio_stream_pb2
# audio_stream_pb2_grpc is not strictly needed for testing the servicer's direct logic,
# unless we were also testing client creation within this file.

# NLU service protos are needed for mocking the NLU call
import nlu_service_pb2
import nlu_service_pb2_grpc # For spec on NLUServiceStub

# The module under test
from service import SpeechToTextServicer # Assuming service.py is in the same directory

class TestSpeechToTextServicer(unittest.TestCase):

    def setUp(self):
        self.servicer = SpeechToTextServicer()
        self.mock_context = mock.Mock(spec=grpc.ServicerContext)

    def test_TranscribeAudioSegment_placeholder_response(self):
        """
        Tests that TranscribeAudioSegment returns the expected placeholder STT response
        (and that NLU call is made, though not primary focus here).
        """
        sample_segment = audio_stream_pb2.AudioSegment(
            session_id="test_session_py_stt",
            sequence_number=123,
            audio_format=audio_stream_pb2.AudioFormat.Value('OPUS'),
            data=b"dummy_opus_data_stt",
            is_final=False
        )

        # Mock the NLU call chain
        mock_nlu_stub_instance = mock.Mock(spec=nlu_service_pb2_grpc.NLUServiceStub)
        mock_nlu_stub_instance.ProcessText.return_value = nlu_service_pb2.NLUResponse(
            session_id=sample_segment.session_id,
            intent="mock_intent_from_stt_test",
            entities=[],
            processed_text="Placeholder transcript...",
            intent_confidence=0.99
        )

        # Patch the grpc.insecure_channel and the NLUServiceStub constructor
        with mock.patch('service.grpc.insecure_channel') as mock_channel_constructor:
            mock_channel_instance = mock.Mock()
            # This mock_channel_instance needs to behave like a context manager for the 'with' statement
            mock_channel_instance.__enter__.return_value = mock_nlu_stub_instance # This is what 'with ... as channel' uses
            mock_channel_constructor.return_value = mock_channel_instance

            with mock.patch('service.nlu_service_pb2_grpc.NLUServiceStub', return_value=mock_nlu_stub_instance) as MockNLUServiceStubConstructor:

                response = self.servicer.TranscribeAudioSegment(sample_segment, self.mock_context)

                # Assertions for STT's own response
                self.assertEqual(response.session_id, sample_segment.session_id)
                self.assertEqual(response.sequence_number, sample_segment.sequence_number)
                expected_transcript_part = f"Placeholder transcript for segment {sample_segment.sequence_number} of session {sample_segment.session_id}."
                self.assertEqual(response.transcript, expected_transcript_part)
                self.assertTrue(response.is_final, "is_final in STT response should be True for placeholder")
                self.assertAlmostEqual(response.confidence, 0.90)

                # Assert NLU call was made
                MockNLUServiceStubConstructor.assert_called_once_with(mock_channel_instance.__enter__.return_value)
                mock_nlu_stub_instance.ProcessText.assert_called_once()
                called_nlu_request = mock_nlu_stub_instance.ProcessText.call_args[0][0]
                self.assertEqual(called_nlu_request.text, expected_transcript_part)
                self.assertEqual(called_nlu_request.session_id, sample_segment.session_id)


    def test_TranscribeAudioSegment_is_final_in_request(self):
        """
        Tests that STT placeholder response includes "(Final)" if request.is_final is True,
        and NLU call is made with this transcript.
        """
        sample_segment_final = audio_stream_pb2.AudioSegment(
            session_id="test_session_py_stt_final",
            sequence_number=456,
            is_final=True # Test with is_final=True
        )

        mock_nlu_stub_instance = mock.Mock(spec=nlu_service_pb2_grpc.NLUServiceStub)
        mock_nlu_stub_instance.ProcessText.return_value = nlu_service_pb2.NLUResponse(session_id=sample_segment_final.session_id)

        with mock.patch('service.grpc.insecure_channel') as mock_channel_constructor:
            mock_channel_instance = mock.Mock()
            mock_channel_instance.__enter__.return_value = mock_nlu_stub_instance
            mock_channel_constructor.return_value = mock_channel_instance
            with mock.patch('service.nlu_service_pb2_grpc.NLUServiceStub', return_value=mock_nlu_stub_instance):

                response = self.servicer.TranscribeAudioSegment(sample_segment_final, self.mock_context)

                expected_transcript = f"Placeholder transcript for segment {sample_segment_final.sequence_number} of session {sample_segment_final.session_id}. (Final)"
                self.assertEqual(response.transcript, expected_transcript)
                self.assertTrue(response.is_final)
                self.assertEqual(response.session_id, sample_segment_final.session_id)

                # Assert NLU call
                mock_nlu_stub_instance.ProcessText.assert_called_once()
                called_nlu_request = mock_nlu_stub_instance.ProcessText.call_args[0][0]
                self.assertEqual(called_nlu_request.text, expected_transcript)


    @mock.patch('service.nlu_service_pb2_grpc.NLUServiceStub') # Patch where NLUServiceStub is looked up (in 'service' module)
    @mock.patch('service.grpc.insecure_channel') # Patch where insecure_channel is looked up
    def test_transcribe_audio_segment_calls_nlu(self, mock_grpc_channel_constructor, MockNLUServiceStubConstructor):
        """Explicitly test the NLU call details."""

        # Configure the mock for grpc.insecure_channel()
        mock_channel_instance = mock.Mock() # This is the channel object
        mock_channel_instance.__enter__.return_value = mock_channel_instance # To support 'with ... as channel:'
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        # Configure the mock for NLUServiceStub constructor to return our specific stub mock
        mock_nlu_stub_instance = mock.Mock(spec=nlu_service_pb2_grpc.NLUServiceStub)
        MockNLUServiceStubConstructor.return_value = mock_nlu_stub_instance

        # Define what the NLU's ProcessText method should return
        stt_session_id = "test_stt_to_nlu_session"
        mock_nlu_stub_instance.ProcessText.return_value = nlu_service_pb2.NLUResponse(
            intent="mock_intent",
            session_id=stt_session_id,
            entities=[nlu_service_pb2.Entity(name="item", value="test")]
        )

        stt_request = audio_stream_pb2.AudioSegment(session_id=stt_session_id, sequence_number=1)

        # Call the method under test
        self.servicer.TranscribeAudioSegment(stt_request, self.mock_context)

        # Assertions
        mock_grpc_channel_constructor.assert_called_once_with(self.servicer.nlu_service_address)
        MockNLUServiceStubConstructor.assert_called_once_with(mock_channel_instance) # Check it was called with the channel
        mock_nlu_stub_instance.ProcessText.assert_called_once()

        called_with_nlu_request = mock_nlu_stub_instance.ProcessText.call_args[0][0]
        self.assertIsInstance(called_with_nlu_request, nlu_service_pb2.NLURequest)
        self.assertEqual(called_with_nlu_request.session_id, stt_session_id)
        self.assertIn("Placeholder transcript", called_with_nlu_request.text)


    @mock.patch('service.nlu_service_pb2_grpc.NLUServiceStub')
    @mock.patch('service.grpc.insecure_channel')
    def test_transcribe_audio_segment_handles_nlu_rpc_error(self, mock_grpc_channel_constructor, MockNLUServiceStubConstructor):
        """Test that STT service handles RpcError from NLU gracefully."""
        mock_channel_instance = mock.Mock()
        mock_channel_instance.__enter__.return_value = mock_channel_instance
        mock_grpc_channel_constructor.return_value = mock_channel_instance

        mock_nlu_stub_instance = mock.Mock(spec=nlu_service_pb2_grpc.NLUServiceStub)
        MockNLUServiceStubConstructor.return_value = mock_nlu_stub_instance

        # Configure ProcessText to raise an RpcError
        # A real RpcError is tricky to instantiate directly for mocks.
        # Using a general Exception or a custom mock that quacks like RpcError.
        # For grpc.RpcError, it's often better to check for its side effects (e.g., logs or default return values).
        # Let's simulate it by raising a generic grpc.RpcError.
        # Note: grpc.RpcError is an exception class, so it should be raised, not just returned.
        simulated_rpc_error = grpc.RpcError("NLU unavailable")
        # To add code and details, you would typically need a more complex setup or a custom error class.
        # For this test, we'll assume the error object itself is sufficient.
        # If specific code() or details() methods are called in the SUT, the mock needs to support them.
        # simulated_rpc_error.code = lambda: grpc.StatusCode.UNAVAILABLE # Example if needed
        # simulated_rpc_error.details = lambda: "NLU service is down"     # Example if needed
        mock_nlu_stub_instance.ProcessText.side_effect = simulated_rpc_error

        stt_request = audio_stream_pb2.AudioSegment(session_id="test_nlu_error_session", sequence_number=1)

        # Call the method and expect it not to crash, and to return the default STT response
        with mock.patch('builtins.print') as mock_print: # To check logs
            response = self.servicer.TranscribeAudioSegment(stt_request, self.mock_context)

            # Check that STT still returns its own response
            self.assertEqual(response.session_id, stt_request.session_id)
            self.assertIn("Placeholder transcript", response.transcript)

            # Check that the NLU error was logged (by checking print calls)
            error_logged = False
            for call_arg in mock_print.call_args_list:
                if "Error calling NLUService" in str(call_arg) and "NLU unavailable" in str(call_arg):
                    error_logged = True
                    break
            self.assertTrue(error_logged, "Expected NLU RpcError to be logged by STT service")

if __name__ == '__main__':
    unittest.main()
