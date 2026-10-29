/**
 * url-codec 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface URLCodecExample {
  /** 示例名称。 */
  label: string
  /** 操作：encode / decode / parse。 */
  operation: 'encode' | 'decode' | 'parse'
  /** 输入内容。 */
  input: string
  /** 组件模式。 */
  component: 'full' | 'path' | 'query' | 'fragment'
  /** 是否把 `+` 视为空格。 */
  plusAsSpace: boolean
}

/** 默认示例：编码一段带中文的查询串。 */
export const DEFAULT_EXAMPLE: URLCodecExample = {
  label: '编码查询串',
  operation: 'encode',
  input: 'https://example.com/search?q=蓝云 工具箱&page=2',
  component: 'full',
  plusAsSpace: true,
}

/** 解码示例。 */
export const DECODE_EXAMPLE: URLCodecExample = {
  label: '解码百分号编码',
  operation: 'decode',
  input: 'q=%E8%93%9D%E4%BA%91%E5%B7%A5%E5%85%B7%E7%AE%B1&page=2',
  component: 'full',
  plusAsSpace: true,
}

/** 解析示例。 */
export const PARSE_EXAMPLE: URLCodecExample = {
  label: '解析 URL 结构',
  operation: 'parse',
  input: 'https://user@example.com:8443/path/to/page?q=1&lang=zh#section',
  component: 'full',
  plusAsSpace: true,
}

/** 全部示例。 */
export const EXAMPLES: readonly URLCodecExample[] = [
  DEFAULT_EXAMPLE,
  DECODE_EXAMPLE,
  PARSE_EXAMPLE,
]
