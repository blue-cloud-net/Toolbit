/**
 * encoding-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface EncodingConvertExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 输入的传输编码。 */
  inputEncoding: 'utf8' | 'hex' | 'base64' | 'base64url'
  /** 源字符集。 */
  from: string
  /** 目标字符集。 */
  to: string
  /** 是否写入字节序标记。 */
  bom: boolean
  /** 输出编码。 */
  outputEncoding: 'utf8' | 'hex' | 'base64' | 'base64url'
}

/** 默认示例：把 GBK 字节还原成 UTF-8 文本。 */
export const DEFAULT_EXAMPLE: EncodingConvertExample = {
  label: 'GBK 转 UTF-8',
  input: '1tDOxA==',
  inputEncoding: 'base64',
  from: 'gbk',
  to: 'utf-8',
  bom: false,
  outputEncoding: 'utf8',
}

/** 自动探测示例。 */
export const AUTO_EXAMPLE: EncodingConvertExample = {
  label: '自动探测源字符集',
  input: 'hello',
  inputEncoding: 'utf8',
  from: 'auto',
  to: 'utf-8',
  bom: false,
  outputEncoding: 'utf8',
}

/** 全部示例。 */
export const EXAMPLES: readonly EncodingConvertExample[] = [
  DEFAULT_EXAMPLE,
  AUTO_EXAMPLE,
]
