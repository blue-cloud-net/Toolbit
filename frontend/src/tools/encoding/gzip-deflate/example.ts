/**
 * gzip-deflate 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface GzipDeflateExample {
  /** 示例名称。 */
  label: string
  /** 操作：compress / decompress。 */
  operation: 'compress' | 'decompress'
  /** 容器格式。 */
  format: 'gzip' | 'deflate' | 'raw'
  /** 输入内容。 */
  input: string
  /** 输入的传输编码。 */
  inputEncoding: 'utf8' | 'hex' | 'base64' | 'base64url'
  /** 压缩级别。 */
  level: number
  /** 输出编码。 */
  outputEncoding: 'utf8' | 'hex' | 'base64' | 'base64url'
}

/** 默认示例：压缩一段文本。 */
export const DEFAULT_EXAMPLE: GzipDeflateExample = {
  label: 'gzip 压缩',
  operation: 'compress',
  format: 'gzip',
  input: 'Hello, Toolbit!',
  inputEncoding: 'utf8',
  level: 6,
  outputEncoding: 'base64',
}

/** 解压示例：解开一段 gzip 数据。 */
export const DECOMPRESS_EXAMPLE: GzipDeflateExample = {
  label: 'gzip 解压',
  operation: 'decompress',
  format: 'gzip',
  input: 'H4sIAAAAAAAAA/NIzcnJBwCCidH3BQAAAA==',
  inputEncoding: 'base64',
  level: 6,
  outputEncoding: 'utf8',
}

/** 全部示例。 */
export const EXAMPLES: readonly GzipDeflateExample[] = [DEFAULT_EXAMPLE, DECOMPRESS_EXAMPLE]
