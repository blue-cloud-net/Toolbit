/**
 * http-status 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface HttpStatusExample {
  /** 示例名称。 */
  label: string
  /** 查询词：状态码或关键字。 */
  query: string
}

/** 默认示例：按状态码查 404。 */
export const DEFAULT_EXAMPLE: HttpStatusExample = {
  label: '404 未找到',
  query: '404',
}

/** 服务端错误示例。 */
export const SERVER_ERROR_EXAMPLE: HttpStatusExample = {
  label: '500 服务端错误',
  query: '500',
}

/** 关键字示例。 */
export const KEYWORD_EXAMPLE: HttpStatusExample = {
  label: '按关键字查重定向',
  query: '重定向',
}

/** 全部示例。 */
export const EXAMPLES: readonly HttpStatusExample[] = [
  DEFAULT_EXAMPLE,
  SERVER_ERROR_EXAMPLE,
  KEYWORD_EXAMPLE,
]
