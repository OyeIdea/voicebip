# real_time_processing_engine/speech_to_text_service/service.py

import grpc
from concurrent import futures
import time
import asyncio
import atexit # For cleanup

# Import generated protobuf and gRPC modules for STT
import audio_stream_pb2
import audio_stream_pb2_grpc

# Import generated protobuf and gRPC modules for NLU
import nlu_service_pb2
import nlu_service_pb2_grpc

# Deepgram SDK components
from deepgram import DeepgramClient, LiveTranscriptionEvents, LiveOptions, DeepgramClientOptions
from .config import DEEPGRAM_API_KEY


class SpeechToTextServicer(audio_stream_pb2_grpc.SpeechToTextServicer):
    """
    Implements the SpeechToText gRPC service using Deepgram for transcription.
    """
    def __init__(self):
        # if not DEEPGRAM_API_KEY:
        #     print("CRITICAL: DEEPGRAM_API_KEY is not set. SpeechToTextServicer cannot function.")
            # In a real scenario, you might raise an exception or prevent server startup

        # It's important that DeepgramClientOptions is initialized correctly.
        # If DEEPGRAM_API_KEY is None, DeepgramClient will raise an error.
        # We ensure it's at least an empty string if not set, though Deepgram will reject it.
        # dg_api_key_to_use = DEEPGRAM_API_KEY if DEEPGRAM_API_KEY else ""

        # self.deepgram_config_options = DeepgramClientOptions(api_key=dg_api_key_to_use, verbose=0) # verbose=0 for less SDK logging
        # self.deepgram_client = DeepgramClient(self.deepgram_config_options)

        # self.active_streams = {} # {session_id: dg_connection}
        # self.transcription_results = {} # {session_id: asyncio.Queue}
        # self.loop = None # Will be set in ensure_event_loop
        # self._ensure_event_loop_is_running_in_thread()
        print("SpeechToTextServicer initialized for simulated Deepgram interaction.")
        # Register cleanup function to be called on exit
        # atexit.register(self.cleanup_all_streams_on_exit)


    def _ensure_event_loop_is_running_in_thread(self):
        """Ensures an asyncio event loop is running in a daemon thread. (No-op for simulation)"""
        # try:
        #     self.loop = asyncio.get_event_loop()
        # except RuntimeError:
        #     self.loop = asyncio.new_event_loop()
        #     asyncio.set_event_loop(self.loop)

        # if not self.loop.is_running():
        #     import threading
        #     self.event_loop_thread = threading.Thread(target=self.loop.run_forever, daemon=True)
        #     self.event_loop_thread.start()
        #     print("Asyncio event loop started in a new thread.")
        pass


    async def _on_deepgram_message(self, session_id, result, **kwargs):
        """Handles messages from Deepgram. (No-op for simulation)"""
        # transcript = ""
        # confidence = 0.0
        # is_final_dg = False

        # if result.is_final and result.channel and result.channel.alternatives and result.channel.alternatives[0].transcript:
        #     transcript = result.channel.alternatives[0].transcript
        #     confidence = result.channel.alternatives[0].confidence
        #     is_final_dg = True
        #     print(f"Deepgram final transcript for {session_id}: '{transcript}' (Confidence: {confidence})")
        # elif not result.is_final and result.channel and result.channel.alternatives and result.channel.alternatives[0].transcript:
        #     transcript = result.channel.alternatives[0].transcript
        #     # For interim results, confidence might not be as relevant or stable
        #     print(f"Deepgram interim transcript for {session_id}: '{transcript}'")

        # if transcript: # Only put if there's something to put
        #     if session_id in self.transcription_results:
        #         await self.transcription_results[session_id].put({
        #             "transcript": transcript,
        #             "confidence": confidence,
        #             "is_final": is_final_dg
        #         })
        #     else:
        #         print(f"Warning: Received transcript for {session_id} but no active queue.")
        pass


    async def _get_or_create_deepgram_connection(self, session_id: str, audio_format_enum):
        """
        Simulates getting or creating a Deepgram connection.
        In this simulation, it does not connect to Deepgram.
        Returns a placeholder indicating success.
        """
        print(f"Simulating _get_or_create_deepgram_connection for session {session_id}, format {audio_format_enum}")
        # For simulation, we don't need a real connection object.
        # Returning True or a simple marker.
        return True # Indicates "connection" is "successful" for simulation purposes

    async def _close_deepgram_stream(self, session_id):
        """Simulates closing a Deepgram stream. (No-op for simulation)"""
        print(f"Simulating _close_deepgram_stream for session {session_id}")
        # No actual stream to close.
        pass


    def TranscribeAudioSegment(self, request: audio_stream_pb2.AudioSegment, context):
        session_id = request.session_id
        # audio_data = request.data # Not used in simulation directly for sending
        is_final_segment_from_client = request.is_final

        # Simulate Deepgram interaction
        print(f"Simulating TranscribeAudioSegment for session {session_id}, seq {request.sequence_number}, is_final: {is_final_segment_from_client}")

        # No actual Deepgram connection or sending audio_data
        # dg_connection = self._get_or_create_deepgram_connection(session_id, request.audio_format) # Simplified call, not async here
        # if not dg_connection:
        #     return self._handle_stt_error(session_id, "[STT Sim Error: Failed to 'connect' to Deepgram]")

        # Dummy transcript generation
        transcript_text = f"Simulated transcript for session {session_id}, seq {request.sequence_number}. Audio format: {audio_stream_pb2.AudioFormat.Name(request.audio_format)}. Data length: {len(request.data)}"
        transcript_confidence = 0.95
        is_final_transcript_from_dg = request.is_final # Simulate finality based on client's is_final

        print(f"Generated simulated transcript for {session_id}: '{transcript_text}' (Confidence: {transcript_confidence}), is_final: {is_final_transcript_from_dg}")

        # Call NLUService with the dummy transcript
        final_stt_response = audio_stream_pb2.TranscriptionResponse(
            session_id=session_id,
            transcript=transcript_text,
            is_final=is_final_transcript_from_dg,
            confidence=float(transcript_confidence)
        )
        self._call_nlu_service(session_id, transcript_text)

        return final_stt_response

    def _handle_stt_error(self, session_id, error_transcript_text):
        print(f"STT Error for {session_id}: {error_transcript_text}")
        # Ensure NLU is called even if there was an STT error (simulated or otherwise)
        self._call_nlu_service(session_id, error_transcript_text)
        return audio_stream_pb2.TranscriptionResponse(
            session_id=session_id,
            transcript=error_transcript_text,
            is_final=True, # Error implies this attempt is final
            confidence=0.0
        )

    def _call_nlu_service(self, session_id, transcript_text):
        # NLU call config (should be part of __init__ or class const)
        nlu_service_address = 'localhost:50053'
        try:
            with grpc.insecure_channel(nlu_service_address) as nlu_channel:
                nlu_stub = nlu_service_pb2_grpc.NLUServiceStub(nlu_channel)
                nlu_request = nlu_service_pb2.NLURequest(
                    text=transcript_text if transcript_text else "[No transcript from STT]", # Send empty or error string
                    session_id=session_id
                )
                print(f"SpeechToTextService (Simulated): Calling NLUService at {nlu_service_address} for SID {session_id} with text: '{nlu_request.text}'")
                nlu_response = nlu_stub.ProcessText(nlu_request, timeout=10)
                if nlu_response:
                    entities_log = [(e.name, e.value, f"{e.confidence:.2f}") for e in nlu_response.entities]
                    print(f"SpeechToTextService (Simulated): NLU response for SID {nlu_response.session_id}: Intent='{nlu_response.intent}' (Conf: {nlu_response.intent_confidence:.2f}), Entities={entities_log}")
                else:
                    print(f"SpeechToTextService (Simulated): No response from NLUService for SID {session_id}")
        except grpc.RpcError as e:
            print(f"SpeechToTextService (Simulated): Error calling NLUService for SID {session_id}: Code={e.code()}, Details='{e.details()}'")
        except Exception as e:
            print(f"SpeechToTextService (Simulated): Unexpected Python error calling NLUService for SID {session_id}: {e}")

    def cleanup_all_streams_on_exit(self):
        print("SpeechToTextServicer (Simulated): cleanup_all_streams_on_exit called. No active streams to clean in simulation.")
        # if self.loop and self.loop.is_running():
        #     for session_id in list(self.active_streams.keys()):
        #         print(f"Stopping Deepgram stream for session: {session_id}")
        #         future = asyncio.run_coroutine_threadsafe(self._close_deepgram_stream(session_id), self.loop)
        #         try:
        #             future.result(timeout=5) # Wait for cleanup to complete
        #         except TimeoutError:
        #             print(f"Timeout cleaning up stream for session {session_id}")
        #         except Exception as e:
        #              print(f"Exception during cleanup of stream {session_id}: {e}")

            # Gently stop the event loop
            # self.loop.call_soon_threadsafe(self.loop.stop)
            # Wait for the thread to finish if it was started by this class
            # if hasattr(self, 'event_loop_thread') and self.event_loop_thread.is_running(): # is_alive() for Thread
            #      print("Waiting for event loop thread to exit...")
            #      self.event_loop_thread.join(timeout=5)
            #      if self.event_loop_thread.is_alive():
            #          print("Event loop thread did not exit cleanly.")
            #      else:
            #          print("Event loop thread exited.")
        # else:
        #     print("No active event loop to clean up streams with.")
        print("SpeechToTextServicer (Simulated): Cleanup complete.")


# Original SpeechToTextService class (business logic) - can be removed or kept if refactored
# class SpeechToTextService: ...

servicer_instance = None # Global instance for atexit cleanup

def serve():
    global servicer_instance
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    servicer_instance = SpeechToTextServicer() # Assign to global for cleanup
    audio_stream_pb2_grpc.add_SpeechToTextServicer_to_server(servicer_instance, server)

    listen_addr = '[::]:50052'
    server.add_insecure_port(listen_addr)

    print(f"SpeechToTextService gRPC server starting on {listen_addr}")
    server.start()
    print(f"Server started. Use Ctrl+C to stop.") # Changed message
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        print("KeyboardInterrupt received, shutting down server...")
    finally:
        # atexit will handle servicer_instance.cleanup_all_streams_on_exit()
        print("Stopping gRPC server...")
        server.stop(10).wait() # Wait 10 seconds for graceful shutdown
        print("gRPC server stopped.")


if __name__ == "__main__":
    serve()
