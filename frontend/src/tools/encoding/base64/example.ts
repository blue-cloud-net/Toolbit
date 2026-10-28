/**
 * base64 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface Base64Example {
  /** 示例名称。 */
  label: string
  /** 操作：encode / decode。 */
  operation: 'encode' | 'decode'
  /** 输入内容。 */
  input: string
  /** 输入编码。 */
  inputEncoding: 'utf8' | 'hex'
  /** Base64 变体。 */
  variant: 'standard' | 'url'
  /** 是否保留填充。 */
  padding: boolean
  /** 输出编码。 */
  outputEncoding: 'utf8' | 'hex'
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
}

/** 全部示例。 */
export const EXAMPLES: readonly Base64Example[] = [DEFAULT_EXAMPLE, DECODE_EXAMPLE]
