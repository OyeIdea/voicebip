# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: nlu_service.proto
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
    'nlu_service.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x11nlu_service.proto\x12\x12\x61i_ml_services.nlu\".\n\nNLURequest\x12\x0c\n\x04text\x18\x01 \x01(\t\x12\x12\n\nsession_id\x18\x02 \x01(\t\"9\n\x06\x45ntity\x12\x0c\n\x04name\x18\x01 \x01(\t\x12\r\n\x05value\x18\x02 \x01(\t\x12\x12\n\nconfidence\x18\x03 \x01(\x02\"\x92\x01\n\x0bNLUResponse\x12\x12\n\nsession_id\x18\x01 \x01(\t\x12\x0e\n\x06intent\x18\x02 \x01(\t\x12,\n\x08\x65ntities\x18\x03 \x03(\x0b\x32\x1a.ai_ml_services.nlu.Entity\x12\x16\n\x0eprocessed_text\x18\x04 \x01(\t\x12\x19\n\x11intent_confidence\x18\x05 \x01(\x02\x32\\\n\nNLUService\x12N\n\x0bProcessText\x12\x1e.ai_ml_services.nlu.NLURequest\x1a\x1f.ai_ml_services.nlu.NLUResponseB/Z-revovoiceai/ai_ml_services/protos/nlu_serviceb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'nlu_service_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z-revovoiceai/ai_ml_services/protos/nlu_service'
  _globals['_NLUREQUEST']._serialized_start=41
  _globals['_NLUREQUEST']._serialized_end=87
  _globals['_ENTITY']._serialized_start=89
  _globals['_ENTITY']._serialized_end=146
  _globals['_NLURESPONSE']._serialized_start=149
  _globals['_NLURESPONSE']._serialized_end=295
  _globals['_NLUSERVICE']._serialized_start=297
  _globals['_NLUSERVICE']._serialized_end=389
# @@protoc_insertion_point(module_scope)
