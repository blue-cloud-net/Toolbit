/**
 * protobuf-encoder 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实业务数据。
 */
import { createNode, type ProtoFieldNode } from '@/tools/encoding/protobuf-encoder/fields'

export interface ProtobufEncoderExample {
  /** 示例名称。 */
  label: string
  /** 编码模式：wire（按字段规格）或 schema（按 protojson）。 */
  mode: 'wire' | 'schema'
  /** 字段规格。 */
  fields: ProtoFieldNode[]
  /** .proto 描述文本。 */
  schema: string
  /** 顶层消息名。 */
  messageType: string
  /** protojson 文本。 */
  message: string
  /** 输出编码。 */
  outputEncoding: 'hex' | 'base64' | 'base64url'
}

/** 线格式示例：字段 1 = "Alice"，字段 2 = 7。 */
export const WIRE_EXAMPLE: ProtobufEncoderExample = {
  label: '线格式编码',
  mode: 'wire',
  fields: [
    { ...createNode(), field: 1, type: 'string', value: 'Alice' },
    { ...createNode(), field: 2, type: 'int32', value: '7' },
  ],
  schema: '',
  messageType: '',
  message: '',
  outputEncoding: 'hex',
}

/** schema 模式示例：按 Person 描述编码。 */
export const SCHEMA_EXAMPLE: ProtobufEncoderExample = {
  label: '按 .proto 描述编码',
  mode: 'schema',
  fields: [],
  schema: [
    'syntax = "proto3";',
    '',
    'message Person {',
    '  string name = 1;',
    '  int32 id = 2;',
    '}',
  ].join('\n'),
  messageType: 'Person',
  message: '{\n  "name": "Alice",\n  "id": 7\n}',
  outputEncoding: 'hex',
}

/** 全部示例。 */
export const EXAMPLES: ProtobufEncoderExample[] = [WIRE_EXAMPLE, SCHEMA_EXAMPLE]
