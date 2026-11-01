/**
 * base64 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface Base64Example {
  /** 示例名称。 */
  label: string
  /** 操作：encode / decode / convert。 */
  operation: 'encode' | 'decode' | 'convert'
  /** 输入内容。 */
  input: string
  /** 输入编码。 */
  inputEncoding: 'utf8' | 'hex'
  /** 输入变体；`auto` 表示自动识别。 */
  variant: 'standard' | 'url' | 'auto'
  /** 是否保留填充。 */
  padding: boolean
  /** 输出编码（decode 时有效）。 */
  outputEncoding: 'utf8' | 'hex'
  /** 输出变体（convert 时有效）。 */
  targetVariant: 'standard' | 'url'
}

/** 默认示例：编码一段中文。 */
export const DEFAULT_EXAMPLE: Base64Example = {
  label: '编码一段中文',
  operation: 'encode',
  input: '蓝云工具箱 Toolbit',
  inputEncoding: 'utf8',
  variant: 'standard',
  padding: true,
  outputEncoding: 'utf8',
  targetVariant: 'url',
}

/** 解码示例。 */
export const DECODE_EXAMPLE: Base64Example = {
  label: '解码一段 Base64',
  operation: 'decode',
  input: 'SGVsbG8sIFRvb2xiaXQh',
  inputEncoding: 'utf8',
  variant: 'standard',
  padding: true,
  outputEncoding: 'utf8',
  targetVariant: 'url',
}

/** 变体互转示例：标准转 URL 安全。 */
export const CONVERT_EXAMPLE: Base64Example = {
  label: '标准转 URL 安全',
  operation: 'convert',
  input: '++//Pw==',
  inputEncoding: 'utf8',
  variant: 'auto',
  padding: true,
  outputEncoding: 'utf8',
  targetVariant: 'url',
}

/** 全部示例。 */
export const EXAMPLES: readonly Base64Example[] = [
  DEFAULT_EXAMPLE,
  DECODE_EXAMPLE,
  CONVERT_EXAMPLE,
]
