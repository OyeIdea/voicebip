# real_time_processing_engine/speech_to_text_service/service.py

import grpc
from concurrent import futures
import time # For the main loop of serve()

# Import generated protobuf and gRPC modules for STT
import audio_stream_pb2
import audio_stream_pb2_grpc

# Import generated protobuf and gRPC modules for NLU
# These files (nlu_service_pb2.py, nlu_service_pb2_grpc.py)
# should have been copied into this directory in the previous subtask.
import nlu_service_pb2
import nlu_service_pb2_grpc

class SpeechToTextServicer(audio_stream_pb2_grpc.SpeechToTextServicer):
    """
    Implements the SpeechToText gRPC service.
    """
    def __init__(self):
        # Configuration for NLU service endpoint
        self.nlu_service_address = 'localhost:50053' # Make this configurable in a real app

    def TranscribeAudioSegment(self, request: audio_stream_pb2.AudioSegment, context):
        """
        Receives an audio segment for transcription and then calls NLUService.
        """
        print(f"SpeechToTextService: Received AudioSegment: SID={request.session_id}, Seq={request.sequence_number}, Format={request.audio_format}, DataLen={len(request.data)}, IsFinalReq={request.is_final}")

        # Placeholder transcription logic
        transcript_text = f"Placeholder transcript for segment {request.sequence_number} of session {request.session_id}."
        if request.is_final:
            transcript_text += " (Final)"

        print(f"SpeechToTextService: Generated transcript: '{transcript_text}'")

        # Call NLUService
        try:
            with grpc.insecure_channel(self.nlu_service_address) as nlu_channel:
                nlu_stub = nlu_service_pb2_grpc.NLUServiceStub(nlu_channel)
                nlu_request = nlu_service_pb2.NLURequest(
                    text=transcript_text,
                    session_id=request.session_id
                )

                print(f"SpeechToTextService: Calling NLUService at {self.nlu_service_address} with text: '{nlu_request.text}'")
                nlu_response = nlu_stub.ProcessText(nlu_request, timeout=10) # Adding a timeout

                if nlu_response:
                    entities_log_format = [(e.name, e.value, f"{e.confidence:.2f}") for e in nlu_response.entities]
                    print(f"SpeechToTextService: Received NLU response for SID {nlu_response.session_id}: Intent='{nlu_response.intent}' (Conf: {nlu_response.intent_confidence:.2f}), Entities={entities_log_format}, ProcessedText='{nlu_response.processed_text}'")
                else:
                    print(f"SpeechToTextService: Received no response from NLUService for SID {request.session_id}")

        except grpc.RpcError as e:
            print(f"SpeechToTextService: Error calling NLUService for SID {request.session_id}: {e.code()} - {e.details()}")
            # Decide if this error should affect the TranscriptionResponse or just be logged.
            # For now, it's just logged, and STT proceeds with its own response.
        except Exception as e:
            print(f"SpeechToTextService: An unexpected error occurred while calling NLUService for SID {request.session_id}: {e}")

        # Return the original placeholder TranscriptionResponse for now
        return audio_stream_pb2.TranscriptionResponse(
            session_id=request.session_id,
            sequence_number=request.sequence_number,
            transcript=transcript_text,
            is_final=True, # For this placeholder, we'll always return a final transcript from STT
            confidence=0.90
        )

class SpeechToTextService:
    """
    Service for converting speech to text.
    This class can hold STT models and configurations.
    The gRPC servicer can instantiate or use this class.
    (Currently not directly used by the servicer's core logic in this simplified version)
    """
    def __init__(self, config=None):
        self.config = config
        print("SpeechToTextService (business logic class) initialized.")

def serve():
    """
    Starts the gRPC server for the SpeechToTextService.
    """
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    # If SpeechToTextServicer had dependencies (like an NLU client class instance),
    # they would be injected here.
    audio_stream_pb2_grpc.add_SpeechToTextServicer_to_server(SpeechToTextServicer(), server)

    listen_addr = '[::]:50052'
    server.add_insecure_port(listen_addr)

    print(f"SpeechToTextService gRPC server starting on {listen_addr}")
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
