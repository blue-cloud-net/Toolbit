/**
 * protobuf-decoder 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实业务数据。
 */
export interface ProtobufDecoderExample {
  /** 示例名称。 */
  label: string
  /** 解析模式：wire（无 schema）或 schema（按 .proto 描述）。 */
  mode: 'wire' | 'schema'
  /** 输入内容。 */
  input: string
  /** 输入的传输编码。 */
  inputEncoding: 'hex' | 'base64' | 'base64url'
  /** .proto 描述文本。 */
  schema: string
  /** 顶层消息名。 */
  messageType: string
}

/** 线格式示例：字段 1 是文本 "Alice"，字段 2 是整数 7。 */
export const WIRE_EXAMPLE: ProtobufDecoderExample = {
  label: '线格式解码',
  mode: 'wire',
  input: '0a05416c6963651007',
  inputEncoding: 'hex',
  schema: '',
  messageType: '',
}

/** schema 模式示例：按 Person 描述解析同一段字节。 */
export const SCHEMA_EXAMPLE: ProtobufDecoderExample = {
  label: '按 .proto 描述解码',
  mode: 'schema',
  input: '0a05416c6963651007',
  inputEncoding: 'hex',
  schema: [
    'syntax = "proto3";',
    '',
    'message Person {',
    '  string name = 1;',
    '  int32 id = 2;',
    '}',
  ].join('\n'),
  messageType: 'Person',
}

/** 全部示例。 */
export const EXAMPLES: ProtobufDecoderExample[] = [WIRE_EXAMPLE, SCHEMA_EXAMPLE]
