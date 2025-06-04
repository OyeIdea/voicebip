import unittest
from unittest import mock
import grpc # For grpc.RpcError

# Assuming the generated files are in the same directory or accessible via PYTHONPATH
# If they are in 'real_time_processing_engine.streaming_data_manager', the import is direct.
import audio_stream_pb2
import audio_stream_pb2_grpc

# The module under test
from manager import StreamIngestServicer # Assuming manager.py is in the same directory

class TestStreamIngestServicer(unittest.TestCase):

    def test_IngestAudioSegment_success(self):
        """Test successful ingestion and forwarding to STT service."""
        servicer = StreamIngestServicer()

        # Create a sample AudioSegment request
        sample_segment = audio_stream_pb2.AudioSegment(
            session_id="test_session_py_sdm",
            sequence_number=1,
            audio_format=audio_stream_pb2.AudioFormat.Value('PCMU'),
            data=b"dummy_audio_data_sdm",
            is_final=False
        )

        # Mock the gRPC context (not strictly needed for this servicer's current logic)
        mock_context = mock.Mock(spec=grpc.ServicerContext)

        # Mock the SpeechToTextStub and its TranscribeAudioSegment method
        mock_stt_stub = mock.Mock(spec=audio_stream_pb2_grpc.SpeechToTextStub)
        mock_transcription_response = audio_stream_pb2.TranscriptionResponse(
            session_id=sample_segment.session_id,
            sequence_number=sample_segment.sequence_number,
            transcript="STT mock transcript",
            is_final=True,
            confidence=0.88
        )
        mock_stt_stub.TranscribeAudioSegment.return_value = mock_transcription_response

        # Patch the grpc.insecure_channel and the SpeechToTextStub constructor
        # The target for patch should be where it's looked up (i.e., in the 'manager' module)
        with mock.patch('manager.grpc.insecure_channel') as mock_channel_constructor:
            # Configure the mock channel to return a context manager that yields the mock_stt_stub
            mock_channel_instance = mock.Mock()
            mock_channel_instance.__enter__.return_value = mock_stt_stub # This is what 'with ... as channel' uses
            mock_channel_constructor.return_value = mock_channel_instance

            # Patch the SpeechToTextStub to return our mock_stt_stub
            with mock.patch('manager.audio_stream_pb2_grpc.SpeechToTextStub', return_value=mock_stt_stub) as mock_StubConstructor:

                # Call the method under test
                response = servicer.IngestAudioSegment(sample_segment, mock_context)

                # Assertions
                # Check that grpc.insecure_channel was called with the correct address
                mock_channel_constructor.assert_called_once_with(servicer.stt_service_address)

                # Check that SpeechToTextStub was instantiated with the channel from insecure_channel
                # This is implicitly tested if TranscribeAudioSegment is called on mock_stt_stub correctly
                mock_StubConstructor.assert_called_once_with(mock_channel_instance.__enter__.return_value)


                # Check that TranscribeAudioSegment was called on the stub with the correct request
                mock_stt_stub.TranscribeAudioSegment.assert_called_once_with(sample_segment, timeout=10)

                # Check the IngestResponse
                self.assertEqual(response.session_id, sample_segment.session_id)
                self.assertEqual(response.sequence_number, sample_segment.sequence_number)
                self.assertIn("Segment received and forwarded to STT", response.status_message)
                self.assertIn("STT mock transcript", response.status_message)

    def test_IngestAudioSegment_stt_rpc_error(self):
        """Test handling of gRPC RpcError when calling STT service."""
        servicer = StreamIngestServicer()
        sample_segment = audio_stream_pb2.AudioSegment(session_id="test_rpc_error", sequence_number=2)
        mock_context = mock.Mock(spec=grpc.ServicerContext)

        # Mock the STT stub to raise an RpcError
        mock_stt_stub = mock.Mock(spec=audio_stream_pb2_grpc.SpeechToTextStub)

        # Create a mock RpcError
        # Note: Creating a real RpcError for testing can be complex as it's usually raised by the grpc library.
        # We can simulate it by having the mock method raise it.
        # For simplicity, we'll mock the error object itself if needed for specific attributes.
        mock_rpc_error = grpc.RpcError("Simulated RpcError from STT")
        # If you need to simulate specific codes or details:
        # mock_rpc_error = mock.Mock(spec=grpc.RpcError)
        # mock_rpc_error.code.return_value = grpc.StatusCode.UNAVAILABLE
        # mock_rpc_error.details.return_value = "STT service down"
        # For this test, a generic RpcError is fine.
        mock_stt_stub.TranscribeAudioSegment.side_effect = mock_rpc_error


        with mock.patch('manager.grpc.insecure_channel') as mock_channel_constructor:
            mock_channel_instance = mock.Mock()
            mock_channel_instance.__enter__.return_value = mock_stt_stub
            mock_channel_constructor.return_value = mock_channel_instance
            with mock.patch('manager.audio_stream_pb2_grpc.SpeechToTextStub', return_value=mock_stt_stub):
                response = servicer.IngestAudioSegment(sample_segment, mock_context)

                mock_stt_stub.TranscribeAudioSegment.assert_called_once_with(sample_segment, timeout=10)
                self.assertEqual(response.session_id, sample_segment.session_id)
                self.assertIn("Segment received, but failed to forward to STT", response.status_message)
                # self.assertIn("Simulated RpcError from STT", response.status_message) # Check specific error message if needed

    def test_IngestAudioSegment_stt_unexpected_error(self):
        """Test handling of unexpected Python errors when calling STT service."""
        servicer = StreamIngestServicer()
        sample_segment = audio_stream_pb2.AudioSegment(session_id="test_unexpected_error", sequence_number=3)
        mock_context = mock.Mock(spec=grpc.ServicerContext)

        mock_stt_stub = mock.Mock(spec=audio_stream_pb2_grpc.SpeechToTextStub)
        mock_stt_stub.TranscribeAudioSegment.side_effect = ValueError("Unexpected Python error in STT call chain")

        with mock.patch('manager.grpc.insecure_channel') as mock_channel_constructor:
            mock_channel_instance = mock.Mock()
            mock_channel_instance.__enter__.return_value = mock_stt_stub
            mock_channel_constructor.return_value = mock_channel_instance
            with mock.patch('manager.audio_stream_pb2_grpc.SpeechToTextStub', return_value=mock_stt_stub):
                response = servicer.IngestAudioSegment(sample_segment, mock_context)

                mock_stt_stub.TranscribeAudioSegment.assert_called_once_with(sample_segment, timeout=10)
                self.assertEqual(response.session_id, sample_segment.session_id)
                self.assertIn("Segment received, but an unexpected error occurred during STT call", response.status_message)
                self.assertIn("Unexpected Python error", response.status_message)


if __name__ == '__main__':
    unittest.main()
