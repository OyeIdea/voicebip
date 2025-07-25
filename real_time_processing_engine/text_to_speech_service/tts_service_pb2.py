# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: tts_service.proto
# Protobuf Python Version: 6.30.0
"""Generated protocol buffer code."""
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import runtime_version as _runtime_version
from google.protobuf import symbol_database as _symbol_database
from google.protobuf.internal import builder as _builder
_runtime_version.ValidateProtobufRuntimeVersion(
    _runtime_version.Domain.PUBLIC,
    6,
    30,
    0,
    '',
    'tts_service.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x11tts_service.proto\x12\x1freal_time_processing_engine.tts\"U\n\nTTSRequest\x12\x1a\n\x12text_to_synthesize\x18\x01 \x01(\t\x12\x12\n\nsession_id\x18\x02 \x01(\t\x12\x17\n\x0fvoice_config_id\x18\x03 \x01(\t\"9\n\x0bTTSResponse\x12\x12\n\nsession_id\x18\x01 \x01(\t\x12\x16\n\x0estatus_message\x18\x02 \x01(\t2\x82\x01\n\x13TextToSpeechService\x12k\n\x0eSynthesizeText\x12+.real_time_processing_engine.tts.TTSRequest\x1a,.real_time_processing_engine.tts.TTSResponseB<Z:revovoiceai/real_time_processing_engine/protos/tts_serviceb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'tts_service_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z:revovoiceai/real_time_processing_engine/protos/tts_service'
  _globals['_TTSREQUEST']._serialized_start=54
  _globals['_TTSREQUEST']._serialized_end=139
  _globals['_TTSRESPONSE']._serialized_start=141
  _globals['_TTSRESPONSE']._serialized_end=198
  _globals['_TEXTTOSPEECHSERVICE']._serialized_start=201
  _globals['_TEXTTOSPEECHSERVICE']._serialized_end=331
# @@protoc_insertion_point(module_scope)
