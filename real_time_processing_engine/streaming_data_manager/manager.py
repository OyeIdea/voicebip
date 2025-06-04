# real_time_processing_engine/streaming_data_manager/manager.py

import grpc
from concurrent import futures
import time # For the main loop of serve()

# Import generated protobuf and gRPC modules
import audio_stream_pb2
import audio_stream_pb2_grpc

# Placeholder for actual import path resolution if these become proper packages
# from ..speech_to_text_service.service import SpeechToTextService
# For now, we'll assume SpeechToTextService would be passed in or available
# in a way that doesn't require a direct runnable import here.

class StreamIngestServicer(audio_stream_pb2_grpc.StreamIngestServicer):
    """
    Implements the StreamIngest gRPC service.
    """
    def __init__(self):
        # Configuration for STT service endpoint
        self.stt_service_address = 'localhost:50052' # Make this configurable in a real app

    def IngestAudioSegment(self, request: audio_stream_pb2.AudioSegment, context):
        """
        Receives an audio segment from a client (e.g., Voice Gateway)
        and forwards it to the SpeechToTextService.
        """
        print(f"StreamingDataManager: Received AudioSegment: SID={request.session_id}, Seq={request.sequence_number}, Format={request.audio_format}, DataLen={len(request.data)}, IsFinal={request.is_final}")

        status_message = "Segment received by StreamingDataManager."

        try:
            # Create a channel and stub to call the SpeechToTextService
            with grpc.insecure_channel(self.stt_service_address) as channel:
                stub = audio_stream_pb2_grpc.SpeechToTextStub(channel)

                # Forward the received AudioSegment to SpeechToTextService
                # print(f"StreamingDataManager: Forwarding segment to STT service at {self.stt_service_address}")
                stt_response = stub.TranscribeAudioSegment(request, timeout=10) # Adding a timeout

                if stt_response:
                    print(f"StreamingDataManager: Received transcription from STT: SID={stt_response.session_id}, Seq={stt_response.sequence_number}, Transcript='{stt_response.transcript}', IsFinal={stt_response.is_final}")
                    status_message = "Segment received and forwarded to STT. STT Response: " + stt_response.transcript
                else:
                    print("StreamingDataManager: Received no response from STT service.")
                    status_message = "Segment received, but no response from STT service."

        except grpc.RpcError as e:
            print(f"StreamingDataManager: Error calling SpeechToTextService: {e.code()} - {e.details()}")
            status_message = f"Segment received, but failed to forward to STT: {e.details()}"
            # Optionally, you could re-raise or handle specific error codes differently
        except Exception as e:
            print(f"StreamingDataManager: An unexpected error occurred while calling STT: {e}")
            status_message = f"Segment received, but an unexpected error occurred during STT call: {e}"


        return audio_stream_pb2.IngestResponse(
            session_id=request.session_id,
            sequence_number=request.sequence_number,
            status_message=status_message
        )

class StreamingDataManager:
    """
    Manages audio streams for real-time processing.
    This includes registering, unregistering, fetching data from,
    and routing streams to other services like Speech-to-Text.
    This class is currently not directly used by the gRPC servicer but can be for future stateful logic.
    """

    def __init__(self, stt_service=None):
        """
        Initializes the StreamingDataManager.
        """
        self.stt_service = stt_service # This would be an instance of a client or logic class
        self._active_streams = {}
        print("StreamingDataManager (logic class) initialized.")
        if self.stt_service:
            print("STT Service instance provided.")
        else:
            print("No STT Service instance provided at init.")

    # ... (other methods like register_stream, unregister_stream remain for now) ...
    def register_stream(self, stream_id: str, stream_source_info: dict):
        if stream_id in self._active_streams:
            print(f"Stream {stream_id} already registered.")
            return False
        print(f"Registering stream: {stream_id} with source info: {stream_source_info}")
        self._active_streams[stream_id] = {"source_info": stream_source_info, "status": "registered"}
        return True

    def unregister_stream(self, stream_id: str):
        if stream_id not in self._active_streams:
            print(f"Stream {stream_id} not found for unregistration.")
            return False
        print(f"Unregistering stream: {stream_id}")
        del self._active_streams[stream_id]
        return True

def serve():
    """
    Starts the gRPC server for the StreamingDataManager.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    # The servicer is instantiated directly here.
    # If it needed access to a StreamingDataManager instance, you'd pass it here.
    audio_stream_pb2_grpc.add_StreamIngestServicer_to_server(StreamIngestServicer(), server)
    
    listen_addr = '[::]:50051'
    server.add_insecure_port(listen_addr)

    print(f"StreamingDataManager gRPC server starting on {listen_addr}")
    server.start()
    print(f"Server started. Waiting for termination...")
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        print("Server stopping...")
        server.stop(0)
        print("Server stopped.")

if __name__ == "__main__":
    serve()
