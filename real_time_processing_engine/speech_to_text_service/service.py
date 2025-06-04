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
        if not DEEPGRAM_API_KEY:
            print("CRITICAL: DEEPGRAM_API_KEY is not set. SpeechToTextServicer cannot function.")
            # In a real scenario, you might raise an exception or prevent server startup

        # It's important that DeepgramClientOptions is initialized correctly.
        # If DEEPGRAM_API_KEY is None, DeepgramClient will raise an error.
        # We ensure it's at least an empty string if not set, though Deepgram will reject it.
        dg_api_key_to_use = DEEPGRAM_API_KEY if DEEPGRAM_API_KEY else ""

        self.deepgram_config_options = DeepgramClientOptions(api_key=dg_api_key_to_use, verbose=0) # verbose=0 for less SDK logging
        self.deepgram_client = DeepgramClient(self.deepgram_config_options)

        self.active_streams = {} # {session_id: dg_connection}
        self.transcription_results = {} # {session_id: asyncio.Queue}
        self.loop = None # Will be set in ensure_event_loop
        self._ensure_event_loop_is_running_in_thread()

        # Register cleanup function to be called on exit
        atexit.register(self.cleanup_all_streams_on_exit)


    def _ensure_event_loop_is_running_in_thread(self):
        """Ensures an asyncio event loop is running in a daemon thread."""
        try:
            self.loop = asyncio.get_event_loop()
        except RuntimeError:
            self.loop = asyncio.new_event_loop()
            asyncio.set_event_loop(self.loop)

        if not self.loop.is_running():
            import threading
            self.event_loop_thread = threading.Thread(target=self.loop.run_forever, daemon=True)
            self.event_loop_thread.start()
            print("Asyncio event loop started in a new thread.")


    async def _on_deepgram_message(self, session_id, result, **kwargs):
        transcript = ""
        confidence = 0.0
        is_final_dg = False

        if result.is_final and result.channel and result.channel.alternatives and result.channel.alternatives[0].transcript:
            transcript = result.channel.alternatives[0].transcript
            confidence = result.channel.alternatives[0].confidence
            is_final_dg = True
            print(f"Deepgram final transcript for {session_id}: '{transcript}' (Confidence: {confidence})")
        elif not result.is_final and result.channel and result.channel.alternatives and result.channel.alternatives[0].transcript:
            transcript = result.channel.alternatives[0].transcript
            # For interim results, confidence might not be as relevant or stable
            print(f"Deepgram interim transcript for {session_id}: '{transcript}'")

        if transcript: # Only put if there's something to put
            if session_id in self.transcription_results:
                await self.transcription_results[session_id].put({
                    "transcript": transcript,
                    "confidence": confidence,
                    "is_final": is_final_dg
                })
            else:
                print(f"Warning: Received transcript for {session_id} but no active queue.")


    async def _get_or_create_deepgram_connection(self, session_id: str, audio_format_enum):
        if session_id not in self.active_streams:
            if not DEEPGRAM_API_KEY: # Double check before attempting to connect
                print(f"Cannot create Deepgram connection for {session_id}: API key not set.")
                return None
            try:
                print(f"Attempting to start Deepgram connection for {session_id} with format {audio_format_enum}")
                dg_connection = self.deepgram_client.listen.live.v("1")

                # Basic mapping from our AudioFormat enum to Deepgram encoding options
                # This needs to be more robust based on actual RTP payload types / SDP negotiation.
                encoding = "linear16" # Default
                sample_rate = 16000 # Default
                channels = 1

                if audio_format_enum == audio_stream_pb2.AudioFormat.Value('PCMU'):
                    encoding = "mulaw"
                    sample_rate = 8000
                elif audio_format_enum == audio_stream_pb2.AudioFormat.Value('PCMA'):
                    # Deepgram might expect PCMA as 'mulaw' or might have specific 'alaw'
                    # For now, assuming it might be handled by 'mulaw' or needs specific check.
                    # Let's default to linear16 if unsure, or research Deepgram's PCMA handling.
                    # For simplicity, we'll use mulaw for PCMA too, but this is a common point of error.
                    # A better approach is to ensure correct encoding string if supported, or transcode.
                    encoding = "mulaw"
                    sample_rate = 8000
                elif audio_format_enum == audio_stream_pb2.AudioFormat.Value('OPUS'):
                    # Deepgram's live streaming typically expects raw audio like PCM, not compressed Opus.
                    # This means Opus would need to be decoded *before* sending to Deepgram.
                    # This is a significant gap if Opus is directly sent.
                    # For now, this will likely lead to errors or poor transcription if Opus bytes are sent.
                    # A real implementation would need an Opus decoder here.
                    # Setting to linear16 as a placeholder if Opus was decoded.
                    print(f"Warning: Received OPUS format for {session_id}. Deepgram expects uncompressed audio. Assuming pre-decoded to linear16 for now.")
                    encoding = "linear16" # Assuming Opus is decoded to PCM elsewhere. This is key.
                    sample_rate = 16000 # Common for Opus wideband

                options = LiveOptions(
                    model="nova-2",
                    language="en-US",
                    smart_format=True,
                    interim_results=True,
                    utterance_end_ms="1000",
                    vad_events=False, # True if you want VAD events from Deepgram
                    encoding=encoding,
                    sample_rate=sample_rate,
                    channels=channels
                )

                # Run the start method in the event loop
                await dg_connection.start(options)

                dg_connection.on(LiveTranscriptionEvents.Transcript, lambda _, result, **kwargs: asyncio.create_task(self._on_deepgram_message(session_id, result, **kwargs)))
                dg_connection.on(LiveTranscriptionEvents.Error, lambda _, error, **kwargs: print(f"Deepgram error for {session_id}: {error}"))
                dg_connection.on(LiveTranscriptionEvents.Close, lambda _, **kwargs: print(f"Deepgram connection closed for {session_id}."))

                self.active_streams[session_id] = dg_connection
                self.transcription_results[session_id] = asyncio.Queue()
                print(f"Deepgram connection started for {session_id} with encoding {encoding}, sample rate {sample_rate}")
            except Exception as e:
                print(f"Error starting Deepgram connection for {session_id}: {e}")
                if session_id in self.active_streams: # Clean up if partial setup
                    del self.active_streams[session_id]
                return None
        return self.active_streams.get(session_id)

    async def _close_deepgram_stream(self, session_id):
        if session_id in self.active_streams:
            connection = self.active_streams.pop(session_id)
            print(f"Deepgram connection finishing for {session_id}...")
            await connection.finish()
            print(f"Deepgram connection finished for {session_id}.")
        if session_id in self.transcription_results:
            del self.transcription_results[session_id]
            print(f"Transcription queue removed for {session_id}.")


    def TranscribeAudioSegment(self, request: audio_stream_pb2.AudioSegment, context):
        session_id = request.session_id
        audio_data = request.data
        is_final_segment_from_client = request.is_final # Client signals end of their stream/utterance

        if not DEEPGRAM_API_KEY:
             return self._handle_stt_error(session_id, "[STT Error: API key not configured]")

        # Ensure event loop is available for async operations
        if not self.loop or not self.loop.is_running():
            print("CRITICAL: Asyncio event loop not running in STT servicer.")
            return self._handle_stt_error(session_id, "[STT Error: Internal event loop issue]")

        # Schedule Deepgram connection and data sending on the event loop
        future_connection = asyncio.run_coroutine_threadsafe(
            self._get_or_create_deepgram_connection(session_id, request.audio_format), self.loop
        )
        dg_connection = future_connection.result(timeout=10) # Wait for connection setup

        if not dg_connection:
            return self._handle_stt_error(session_id, "[STT Error: Failed to connect to Deepgram]")

        # Send audio data (this is a blocking call on the dg_connection object if not wrapped)
        # The SDK's send is designed to be called from a sync context if dg_connection was started in async.
        try:
            dg_connection.send(audio_data)
            # print(f"Sent {len(audio_data)} bytes to Deepgram for session {session_id}")
        except Exception as e:
            print(f"Error sending data to Deepgram for {session_id}: {e}")
            # Potentially close and try to re-establish on next segment, or mark session as error
            asyncio.run_coroutine_threadsafe(self._close_deepgram_stream(session_id), self.loop)
            return self._handle_stt_error(session_id, "[STT Error: Failed to send audio data]")


        transcript_text = ""
        transcript_confidence = 0.0
        # is_final_transcript_from_dg should reflect if Deepgram returned a final transcript for *this utterance*
        is_final_transcript_from_dg = False

        # Only wait for a transcript if the client indicates this is the end of their segment stream / utterance
        if is_final_segment_from_client:
            print(f"Client marked segment as is_final for {session_id}. Attempting to get final transcript from Deepgram.")
            try:
                result = asyncio.run_coroutine_threadsafe(
                    asyncio.wait_for(self.transcription_results[session_id].get(), timeout=5.0), self.loop
                ).result(timeout=5.5) # Wait for the async task to complete

                transcript_text = result.get("transcript", "")
                transcript_confidence = result.get("confidence", 0.0)
                is_final_transcript_from_dg = result.get("is_final", False) # This should be true if DG said it's final
                self.transcription_results[session_id].task_done()
                print(f"Retrieved from queue for {session_id}: '{transcript_text}', final_dg: {is_final_transcript_from_dg}")

                # Important: Close the Deepgram stream *only* if this `is_final_segment_from_client` truly means
                # the end of all audio for this session_id. If more utterances are expected for the same session_id,
                # we might not want to close it yet. For this implementation, assume one utterance per "session"
                # or that is_final_segment_from_client implies end of the whole interaction with STT for this session_id.
                asyncio.run_coroutine_threadsafe(self._close_deepgram_stream(session_id), self.loop).result(timeout=5)

            except (asyncio.TimeoutError, TimeoutError): # TimeoutError for future.result()
                print(f"Timeout waiting for Deepgram transcript for {session_id}")
                transcript_text = "[STT Timeout]"
                is_final_transcript_from_dg = True # Consider timeout as end of this attempt
                asyncio.run_coroutine_threadsafe(self._close_deepgram_stream(session_id), self.loop).result(timeout=5)
            except Exception as e:
                print(f"Error getting transcript from queue for {session_id}: {e}")
                transcript_text = f"[STT Error: {type(e).__name__}]"
                is_final_transcript_from_dg = True # Consider error as end of this attempt
                asyncio.run_coroutine_threadsafe(self._close_deepgram_stream(session_id), self.loop).result(timeout=5)
        else:
            # Non-final segment from client. We don't block for a DG final transcript.
            # We could try to get an interim transcript non-blockingly if available.
            try:
                interim_result = self.transcription_results[session_id].get_nowait()
                transcript_text = interim_result.get("transcript", "")
                transcript_confidence = interim_result.get("confidence", 0.0) # usually 0 for interim
                is_final_transcript_from_dg = interim_result.get("is_final", False)
                self.transcription_results[session_id].task_done()
                print(f"Got interim for {session_id}: '{transcript_text}'")
            except asyncio.QueueEmpty:
                # print(f"No interim transcript available yet for {session_id}")
                pass # No transcript to return for this non-final segment yet.
            except Exception as e:
                print(f"Error getting interim transcript from queue for {session_id}: {e}")


        # Call NLUService
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
        self._call_nlu_service(session_id, error_transcript_text) # Notify NLU even on STT error
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
                print(f"SpeechToTextService: Calling NLUService at {nlu_service_address} for SID {session_id} with text: '{nlu_request.text}'")
                nlu_response = nlu_stub.ProcessText(nlu_request, timeout=10)
                if nlu_response:
                    entities_log = [(e.name, e.value, f"{e.confidence:.2f}") for e in nlu_response.entities]
                    print(f"SpeechToTextService: NLU response for SID {nlu_response.session_id}: Intent='{nlu_response.intent}' (Conf: {nlu_response.intent_confidence:.2f}), Entities={entities_log}")
                else:
                    print(f"SpeechToTextService: No response from NLUService for SID {session_id}")
        except grpc.RpcError as e:
            print(f"SpeechToTextService: Error calling NLUService for SID {session_id}: Code={e.code()}, Details='{e.details()}'")
        except Exception as e:
            print(f"SpeechToTextService: Unexpected Python error calling NLUService for SID {session_id}: {e}")

    def cleanup_all_streams_on_exit(self):
        print("SpeechToTextServicer: Cleaning up all active Deepgram streams on server exit...")
        if self.loop and self.loop.is_running():
            for session_id in list(self.active_streams.keys()):
                print(f"Stopping Deepgram stream for session: {session_id}")
                future = asyncio.run_coroutine_threadsafe(self._close_deepgram_stream(session_id), self.loop)
                try:
                    future.result(timeout=5) # Wait for cleanup to complete
                except TimeoutError:
                    print(f"Timeout cleaning up stream for session {session_id}")
                except Exception as e:
                     print(f"Exception during cleanup of stream {session_id}: {e}")

            # Gently stop the event loop
            self.loop.call_soon_threadsafe(self.loop.stop)
            # Wait for the thread to finish if it was started by this class
            if hasattr(self, 'event_loop_thread') and self.event_loop_thread.is_alive():
                 print("Waiting for event loop thread to exit...")
                 self.event_loop_thread.join(timeout=5)
                 if self.event_loop_thread.is_alive():
                     print("Event loop thread did not exit cleanly.")
                 else:
                     print("Event loop thread exited.")
        else:
            print("No active event loop to clean up streams with.")
        print("SpeechToTextServicer: Cleanup complete.")


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
