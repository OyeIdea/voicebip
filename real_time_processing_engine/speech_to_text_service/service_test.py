import unittest
from unittest import mock
import grpc # For mock_context and RpcError
import asyncio # For asyncio.Queue and asyncio.TimeoutError

import audio_stream_pb2
import nlu_service_pb2
import nlu_service_pb2_grpc

# Import the servicer from the service module
# Assuming 'service.py' is in the same directory or PYTHONPATH is set up
from service import SpeechToTextServicer
from service import DEEPGRAM_API_KEY # To check if it's mocked or has a value

# Mock DeepgramClient and its related classes if they are directly used for type hinting
# or if their methods are called directly on mock objects.
# For this test, we primarily mock at the point of `DeepgramClient()` instantiation.
from deepgram import LiveTranscriptionEvents, LiveOptions, DeepgramClientOptions


# It's good practice to ensure DEEPGRAM_API_KEY is handled for tests,
# either by setting a mock environment variable or by directly patching 'service.DEEPGRAM_API_KEY'.
@mock.patch('service.DEEPGRAM_API_KEY', "test_deepgram_api_key_for_unit_tests") # Patch API key for all tests in this class
class TestSpeechToTextServicerWithDeepgram(unittest.TestCase):

    def setUp(self):
        """Set up for each test method."""
        # Mock the DeepgramClient constructor and its instance methods
        self.deepgram_client_patcher = mock.patch('service.DeepgramClient')
        self.MockDeepgramClientConstructor = self.deepgram_client_patcher.start()

        self.mock_dg_client_instance = self.MockDeepgramClientConstructor.return_value
        self.mock_dg_live_connection = mock.Mock() # This is the object returned by self.deepgram_client.listen.live.v("1")

        # Configure the mock client instance to return the mock live connection
        self.mock_dg_client_instance.listen.live.v.return_value = self.mock_dg_live_connection

        # Mock methods of the live connection object
        # .start() and .finish() are async, so use AsyncMock
        self.mock_dg_live_connection.start = mock.AsyncMock()
        self.mock_dg_live_connection.finish = mock.AsyncMock()
        # .send() is synchronous in the Deepgram SDK if the connection is managed by their ConnectionManager
        # or if called from a synchronous context after async start.
        self.mock_dg_live_connection.send = mock.Mock()
        # .on() is used to register event handlers
        self.mock_dg_live_connection.on = mock.Mock()

        # Mock the NLU client chain (channel and stub)
        self.grpc_channel_patcher = mock.patch('service.grpc.insecure_channel')
        self.mock_grpc_insecure_channel = self.grpc_channel_patcher.start()

        self.nlu_stub_patcher = mock.patch('service.nlu_service_pb2_grpc.NLUServiceStub')
        self.MockNLUServiceStubConstructor = self.nlu_stub_patcher.start()

        self.mock_nlu_channel_instance = mock.Mock()
        self.mock_nlu_channel_instance.__enter__.return_value = self.mock_nlu_channel_instance # For 'with' statement
        self.mock_grpc_insecure_channel.return_value = self.mock_nlu_channel_instance

        self.mock_nlu_stub_instance = self.MockNLUServiceStubConstructor.return_value
        self.mock_nlu_stub_instance.ProcessText.return_value = nlu_service_pb2.NLUResponse(
            intent="mock_nlu_intent",
            session_id="test_session_123" # Ensure this matches or is handled
        )

        # Instantiate the servicer. This will now use the mocked DeepgramClient.
        self.servicer = SpeechToTextServicer()

        # Mock the servicer's asyncio event loop and transcription_results queue for control
        # The servicer's __init__ tries to start its own loop. We might need to control this.
        # For simplicity, we'll assume the loop handling in servicer is okay for testing,
        # and directly manipulate the transcription_results queue if needed.
        # Or, mock the loop related methods if they cause issues during test setup.
        # self.servicer.loop = mock.MagicMock(spec=asyncio.AbstractEventLoop) # If direct control is needed
        # self.servicer.loop.is_running.return_value = True
        # self.servicer.loop.create_task = lambda coro: coro # simple passthrough for testing

        # Ensure the session's queue is created for testing `is_final` logic
        self.test_session_id = "test_session_123"
        self.servicer.transcription_results[self.test_session_id] = asyncio.Queue()

        self.mock_grpc_context = mock.Mock(spec=grpc.ServicerContext)

    def tearDown(self):
        """Clean up after each test method."""
        self.deepgram_client_patcher.stop()
        self.nlu_stub_patcher.stop()
        self.grpc_channel_patcher.stop()

        # Clean up any tasks that might have been created by the servicer's event loop
        # This is important if the servicer's own loop was running.
        if self.servicer.loop and self.servicer.loop.is_running():
            # Cancel all tasks on the loop
            for task in asyncio.all_tasks(loop=self.servicer.loop):
                task.cancel()
            # Give a moment for tasks to cancel - this is a bit of a hack for tests
            # In a real scenario, shutdown is more graceful.
            # self.servicer.loop.call_soon_threadsafe(self.servicer.loop.stop)
            # self.servicer.event_loop_thread.join(timeout=0.1) # if thread was started
            pass


    @mock.patch('asyncio.run_coroutine_threadsafe') # To control future.result()
    def test_transcribe_audio_segment_success_final(self, mock_run_coro_threadsafe):
        """Test successful transcription for a final segment."""

        # --- Setup Mocks for this specific test ---
        # Mock _get_or_create_deepgram_connection to return the connection directly
        # This avoids re-testing its internal logic here.
        # The mock_dg_live_connection is already set up in self.setUp.
        async def mock_get_conn(sid, af): return self.mock_dg_live_connection
        self.servicer._get_or_create_deepgram_connection = mock.MagicMock(side_effect=mock_get_conn)

        # Simulate future.result() for _get_or_create_deepgram_connection
        mock_future_conn = mock.Mock()
        mock_future_conn.result.return_value = self.mock_dg_live_connection
        mock_run_coro_threadsafe.side_effect = [mock_future_conn] # First call is for get_conn

        # Simulate the Deepgram callback putting a result on the queue
        # We will use asyncio.run_coroutine_threadsafe for the queue.get() call
        async def queue_putter_and_closer():
            await self.servicer.transcription_results[self.test_session_id].put({
                "transcript": "Hello Deepgram", "confidence": 0.99, "is_final": True
            })
            # Simulate _close_deepgram_stream being called successfully
            await self.servicer._close_deepgram_stream(self.test_session_id) # This calls finish on mock

        # Mock the future for queue.get() and _close_deepgram_stream()
        mock_future_queue_get = mock.Mock()
        mock_future_queue_get.result.return_value = {"transcript": "Hello Deepgram", "confidence": 0.99, "is_final": True}

        mock_future_close_stream = mock.Mock()
        mock_future_close_stream.result.return_value = None # _close_deepgram_stream returns None

        # Configure run_coroutine_threadsafe:
        # 1st call: _get_or_create_deepgram_connection
        # 2nd call: asyncio.wait_for(self.transcription_results[session_id].get(), timeout=5.0)
        # 3rd call: self._close_deepgram_stream(session_id)
        mock_run_coro_threadsafe.side_effect = [
            mock_future_conn,
            mock_future_queue_get,
            mock_future_close_stream
        ]

        # We also need to ensure the _on_deepgram_message callback (if it was directly called by a mock)
        # correctly uses run_coroutine_threadsafe if it puts to queue.
        # However, for this test, we are directly controlling the queue content via the future mock.

        # --- Prepare Request ---
        request = audio_stream_pb2.AudioSegment(
            session_id=self.test_session_id,
            data=b"audio data",
            is_final=True,
            audio_format=audio_stream_pb2.AudioFormat.Value('PCMU')
        )

        # --- Call Method ---
        response = self.servicer.TranscribeAudioSegment(request, self.mock_grpc_context)

        # --- Assertions ---
        # Deepgram connection and send
        self.servicer._get_or_create_deepgram_connection.assert_called_once_with(self.test_session_id, request.audio_format)
        self.mock_dg_live_connection.send.assert_called_once_with(request.data)

        # Transcript result
        self.assertEqual(response.transcript, "Hello Deepgram")
        self.assertEqual(response.confidence, 0.99)
        self.assertTrue(response.is_final)
        self.assertEqual(response.session_id, self.test_session_id)

        # NLU call
        self.mock_nlu_stub_instance.ProcessText.assert_called_once()
        called_nlu_request = self.mock_nlu_stub_instance.ProcessText.call_args[0][0]
        self.assertEqual(called_nlu_request.text, "Hello Deepgram")
        self.assertEqual(called_nlu_request.session_id, self.test_session_id)

        # Deepgram finish (due to is_final=True and successful transcript retrieval)
        # This is asserted by _close_deepgram_stream being called, which calls finish.
        # The mock_future_close_stream ensures this path is taken.
        self.mock_dg_live_connection.finish.assert_called_once() # finish is async, check if it was awaited


    @mock.patch('asyncio.run_coroutine_threadsafe')
    def test_transcribe_audio_segment_deepgram_connection_error(self, mock_run_coro_threadsafe):
        # Simulate _get_or_create_deepgram_connection failing (returning None)
        async def mock_get_conn_fail(sid, af): return None
        self.servicer._get_or_create_deepgram_connection = mock.MagicMock(side_effect=mock_get_conn_fail)

        mock_future_conn_fail = mock.Mock()
        mock_future_conn_fail.result.return_value = None # Simulate connection failure
        mock_run_coro_threadsafe.return_value = mock_future_conn_fail

        request = audio_stream_pb2.AudioSegment(session_id="dg_conn_error_session", data=b"data", is_final=True)
        response = self.servicer.TranscribeAudioSegment(request, self.mock_grpc_context)

        self.assertEqual(response.transcript, "[STT Error: Failed to connect to Deepgram]")
        self.assertTrue(response.is_final)
        self.assertEqual(response.confidence, 0.0)
        self.mock_nlu_stub_instance.ProcessText.assert_called_once_with(
            nlu_service_pb2.NLURequest(text="[STT Error: Failed to connect to Deepgram]", session_id="dg_conn_error_session"),
            timeout=10
        )

    @mock.patch('asyncio.run_coroutine_threadsafe')
    def test_transcribe_audio_segment_timeout_getting_transcript(self, mock_run_coro_threadsafe):
        # Simulate successful connection
        async def mock_get_conn(sid, af): return self.mock_dg_live_connection
        self.servicer._get_or_create_deepgram_connection = mock.MagicMock(side_effect=mock_get_conn)
        mock_future_conn = mock.Mock()
        mock_future_conn.result.return_value = self.mock_dg_live_connection

        # Simulate timeout when getting from queue
        mock_future_queue_get_timeout = mock.Mock()
        mock_future_queue_get_timeout.result.side_effect = asyncio.TimeoutError # Simulate timeout on future.result()

        mock_future_close_stream = mock.Mock() # For the close call in finally/except
        mock_future_close_stream.result.return_value = None

        mock_run_coro_threadsafe.side_effect = [
            mock_future_conn,
            mock_future_queue_get_timeout,
            mock_future_close_stream
        ]

        request = audio_stream_pb2.AudioSegment(session_id=self.test_session_id, data=b"data", is_final=True)
        response = self.servicer.TranscribeAudioSegment(request, self.mock_grpc_context)

        self.assertEqual(response.transcript, "[STT Timeout]")
        self.assertTrue(response.is_final)
        self.mock_dg_live_connection.send.assert_called_once_with(request.data) # Ensure data was sent
        self.mock_dg_live_connection.finish.assert_called_once() # Ensure stream was closed
        self.mock_nlu_stub_instance.ProcessText.assert_called_once_with(
            nlu_service_pb2.NLURequest(text="[STT Timeout]", session_id=self.test_session_id),
            timeout=10
        )

    def test_transcribe_audio_segment_non_final_no_interim(self, mock_dg_api_key_is_set_for_tests_in_class_decorator):
        # This test relies on the class-level DEEPGRAM_API_KEY mock
        # It also needs to mock the async parts like other tests.
        # We'll use a simplified path where get_nowait() on the queue raises QueueEmpty.

        with mock.patch('asyncio.run_coroutine_threadsafe') as mock_run_coro_threadsafe_local:
            # Simulate successful connection
            async def mock_get_conn(sid, af): return self.mock_dg_live_connection
            self.servicer._get_or_create_deepgram_connection = mock.MagicMock(side_effect=mock_get_conn)
            mock_future_conn = mock.Mock()
            mock_future_conn.result.return_value = self.mock_dg_live_connection
            mock_run_coro_threadsafe_local.return_value = mock_future_conn # For get_connection

            # Ensure the queue is empty for get_nowait()
            # For session_id "non_final_session", a new queue will be created by _get_or_create_deepgram_connection
            # if it doesn't exist. We need to ensure it's empty.
            # The servicer's __init__ now creates a queue if not present, so this should be fine.
            # Or, we can explicitly ensure it's empty for this test's session_id.
            non_final_session_id = "non_final_session"
            self.servicer.transcription_results[non_final_session_id] = asyncio.Queue()


            request = audio_stream_pb2.AudioSegment(session_id=non_final_session_id, data=b"interim data", is_final=False)
            response = self.servicer.TranscribeAudioSegment(request, self.mock_grpc_context)

            self.assertEqual(response.transcript, "") # Expect empty transcript for non-final if no interim
            self.assertFalse(response.is_final)
            self.mock_dg_live_connection.send.assert_called_once_with(request.data)
            self.mock_dg_live_connection.finish.assert_not_called() # Finish should not be called for non-final client segment
            self.mock_nlu_stub_instance.ProcessText.assert_called_once_with(
                 nlu_service_pb2.NLURequest(text="[No transcript from STT]", session_id=non_final_session_id), # Empty transcript sent to NLU
                 timeout=10
            )


    # Test for API key not set
    @mock.patch('service.DEEPGRAM_API_KEY', None) # Override class-level patch for this test
    def test_transcribe_audio_segment_no_api_key(self, _): # Param for the DEEPGRAM_API_KEY patch
        # Re-initialize servicer with API key as None
        # The __init__ has a check, but TranscribeAudioSegment also checks.
        # We need to ensure the servicer instance used by the test reflects this.
        # For simplicity, we assume the class-level patch makes DEEPGRAM_API_KEY None *before* servicer init.
        # However, setUp runs first. So, we might need to re-init or patch differently.

        # Let's test the check within TranscribeAudioSegment directly by re-patching for this method's scope
        with mock.patch('service.DEEPGRAM_API_KEY', None): # Shadow the class-level mock for this test
             # If servicer is created in setUp, it already has the test key.
             # This shows a limitation if API key is only checked at init.
             # The current servicer code checks DEEPGRAM_API_KEY also at the start of TranscribeAudioSegment.
            servicer_no_key = SpeechToTextServicer() # This instance will see DEEPGRAM_API_KEY as None during its __init__

            request = audio_stream_pb2.AudioSegment(session_id="no_api_key_session", data=b"data", is_final=True)
            response = servicer_no_key.TranscribeAudioSegment(request, self.mock_grpc_context) # Use the new instance

            self.assertEqual(response.transcript, "[STT Error: API key not configured]")
            self.assertTrue(response.is_final)
            # NLU call should still happen with the error message
            self.mock_nlu_stub_instance.ProcessText.assert_called_once_with(
                 nlu_service_pb2.NLURequest(text="[STT Error: API key not configured]", session_id="no_api_key_session"),
                 timeout=10
            )


if __name__ == '__main__':
    unittest.main()
