# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# NO CHECKED-IN PROTOBUF GENCODE
# source: dialogue_management_service.proto
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
    'dialogue_management_service.proto'
)
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


import nlu_service_pb2 as nlu__service__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n!dialogue_management_service.proto\x12\"ai_ml_services.dialogue_management\x1a\x11nlu_service.proto\"Z\n\x0f\x44ialogueRequest\x12\x12\n\nsession_id\x18\x01 \x01(\t\x12\x33\n\nnlu_result\x18\x02 \x01(\x0b\x32\x1f.ai_ml_services.nlu.NLUResponse\"=\n\x10\x44ialogueResponse\x12\x12\n\nsession_id\x18\x01 \x01(\t\x12\x15\n\rtext_response\x18\x02 \x01(\t2\x94\x01\n\x19\x44ialogueManagementService\x12w\n\nManageTurn\x12\x33.ai_ml_services.dialogue_management.DialogueRequest\x1a\x34.ai_ml_services.dialogue_management.DialogueResponseB?Z=revovoiceai/ai_ml_services/protos/dialogue_management_serviceb\x06proto3')

_globals = globals()
_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, _globals)
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'dialogue_management_service_pb2', _globals)
if not _descriptor._USE_C_DESCRIPTORS:
  _globals['DESCRIPTOR']._loaded_options = None
  _globals['DESCRIPTOR']._serialized_options = b'Z=revovoiceai/ai_ml_services/protos/dialogue_management_service'
  _globals['_DIALOGUEREQUEST']._serialized_start=92
  _globals['_DIALOGUEREQUEST']._serialized_end=182
  _globals['_DIALOGUERESPONSE']._serialized_start=184
  _globals['_DIALOGUERESPONSE']._serialized_end=245
  _globals['_DIALOGUEMANAGEMENTSERVICE']._serialized_start=248
  _globals['_DIALOGUEMANAGEMENTSERVICE']._serialized_end=396
# @@protoc_insertion_point(module_scope)
