# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc
import warnings

import tts_service_pb2 as tts__service__pb2

GRPC_GENERATED_VERSION = '1.72.1'
GRPC_VERSION = grpc.__version__
_version_not_supported = False

try:
    from grpc._utilities import first_version_is_lower
    _version_not_supported = first_version_is_lower(GRPC_VERSION, GRPC_GENERATED_VERSION)
except ImportError:
    _version_not_supported = True

if _version_not_supported:
    raise RuntimeError(
        f'The grpc package installed is at version {GRPC_VERSION},'
        + f' but the generated code in tts_service_pb2_grpc.py depends on'
        + f' grpcio>={GRPC_GENERATED_VERSION}.'
        + f' Please upgrade your grpc module to grpcio>={GRPC_GENERATED_VERSION}'
        + f' or downgrade your generated code using grpcio-tools<={GRPC_VERSION}.'
    )


class TextToSpeechServiceStub(object):
    """TextToSpeechService definition
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.SynthesizeText = channel.unary_unary(
                '/real_time_processing_engine.tts.TextToSpeechService/SynthesizeText',
                request_serializer=tts__service__pb2.TTSRequest.SerializeToString,
                response_deserializer=tts__service__pb2.TTSResponse.FromString,
                _registered_method=True)


class TextToSpeechServiceServicer(object):
    """TextToSpeechService definition
    """

    def SynthesizeText(self, request, context):
        """Synthesizes text to speech
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_TextToSpeechServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'SynthesizeText': grpc.unary_unary_rpc_method_handler(
                    servicer.SynthesizeText,
                    request_deserializer=tts__service__pb2.TTSRequest.FromString,
                    response_serializer=tts__service__pb2.TTSResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'real_time_processing_engine.tts.TextToSpeechService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))
    server.add_registered_method_handlers('real_time_processing_engine.tts.TextToSpeechService', rpc_method_handlers)


 # This class is part of an EXPERIMENTAL API.
class TextToSpeechService(object):
    """TextToSpeechService definition
    """

    @staticmethod
    def SynthesizeText(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(
            request,
            target,
            '/real_time_processing_engine.tts.TextToSpeechService/SynthesizeText',
            tts__service__pb2.TTSRequest.SerializeToString,
            tts__service__pb2.TTSResponse.FromString,
            options,
            channel_credentials,
            insecure,
            call_credentials,
            compression,
            wait_for_ready,
            timeout,
            metadata,
            _registered_method=True)
