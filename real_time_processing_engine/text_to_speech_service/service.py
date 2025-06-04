# real_time_processing_engine/text_to_speech_service/service.py

import grpc
from concurrent import futures
import time # Required for server.wait_for_termination() in a loop

# Import generated protobuf and gRPC modules
# These should be in the same directory or Python path
import tts_service_pb2
import tts_service_pb2_grpc

# Optional: for more advanced logging
# import logging
# logging.basicConfig(level=logging.INFO)


class TextToSpeechServicer(tts_service_pb2_grpc.TextToSpeechServiceServicer):
    """
    Implements the TextToSpeechService gRPC interface.
    """
    def SynthesizeText(self, request: tts_service_pb2.TTSRequest, context):
        """
        Receives text and returns a status message.
        Placeholder for actual speech synthesis.
        """
        print(f"TextToSpeechService: Received text '{request.text_to_synthesize}' for session_id '{request.session_id}'. Voice config: '{request.voice_config_id}'")

        # Placeholder TTS logic:
        # In a real implementation, this method would:
        # 1. Preprocess `request.text_to_synthesize`.
        # 2. Use a TTS engine (with `request.voice_config_id` if provided) to generate audio data.
        # 3. If streaming audio, it would return a stream of audio chunks.
        # 4. For non-streaming, it might return the audio data directly in TTSResponse (if small)
        #    or provide a way to fetch it (e.g., a URL or stream ID).
        # For this placeholder, we just acknowledge receipt.
        
        status_message = f"Text for session '{request.session_id}' (voice: '{request.voice_config_id if request.voice_config_id else 'default'}') received by TTS. Placeholder synthesis initiated."
        
        print(f"TextToSpeechService: Status for SID '{request.session_id}': {status_message}")

        return tts_service_pb2.TTSResponse(
            session_id=request.session_id,
            status_message=status_message
            # audio_data and error_message fields are omitted for this placeholder
        )

def serve():
    """
    Starts the gRPC server for the TextToSpeechService.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    tts_service_pb2_grpc.add_TextToSpeechServiceServicer_to_server(
        TextToSpeechServicer(), server
    )
    
    port = "50055" # Define the port for Text-to-Speech service
    listen_addr = f'[::]:{port}'
    server.add_insecure_port(listen_addr)
    
    server.start()
    print(f"TextToSpeechService server started, listening on port {port}")

    try:
        # Keep the server running until interrupted
        # server.wait_for_termination() # This blocks indefinitely
        while True:
            time.sleep(86400) # Sleep for a day, effectively keeping the thread alive
    except KeyboardInterrupt:
        print("TextToSpeechService server stopping...")
        server.stop(0) # Graceful stop
        print("TextToSpeechService server stopped.")

# Main guard to run the server when the script is executed
if __name__ == '__main__':
    # The original TextToSpeechService class (business logic) can be instantiated here
    # if the servicer needs to use it, e.g., for loading models, managing voices.
    # For this task, the servicer implements logic directly.
    # e.g., tts_logic = TextToSpeechService()
    # servicer_instance = TextToSpeechServicer(logic=tts_logic)
    # and then pass `servicer_instance` to add_TextToSpeechServiceServicer_to_server.

    serve()
