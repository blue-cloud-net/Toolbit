/**
 * ua-parser 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface UAParserExample {
  /** 示例名称。 */
  label: string
  /** 待解析的 User-Agent 串。 */
  userAgent: string
}

/** 默认示例：Chrome 桌面版。 */
export const DEFAULT_EXAMPLE: UAParserExample = {
  label: 'Chrome 桌面版',
  userAgent:
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
}

/** 移动端示例：iOS Safari。 */
export const MOBILE_EXAMPLE: UAParserExample = {
  label: 'iOS Safari',
  userAgent:
    'Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1',
}

/** 爬虫示例。 */
export const BOT_EXAMPLE: UAParserExample = {
  label: 'Googlebot',
  userAgent: 'Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)',
}

/** 全部示例。 */
export const EXAMPLES: readonly UAParserExample[] = [
  DEFAULT_EXAMPLE,
  MOBILE_EXAMPLE,
  BOT_EXAMPLE,
]
