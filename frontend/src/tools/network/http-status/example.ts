/**
 * http-status 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 查询方式。 */
export type StatusQueryKind = 'code' | 'keyword'

/** 一个示例。 */
export interface HttpStatusExample {
  /** 示例名称。 */
  label: string
  /** 查询方式。 */
  kind: StatusQueryKind
  /** 状态码；仅 `kind` 为 `code` 时有意义。 */
  code: string
  /** 关键字；仅 `kind` 为 `keyword` 时有意义。 */
  keyword: string
}

/** 默认示例：按状态码查 404。 */
export const DEFAULT_EXAMPLE: HttpStatusExample = {
  label: '404 未找到',
  kind: 'code',
  code: '404',
  keyword: '',
}

/** 服务端错误示例。 */
export const SERVER_ERROR_EXAMPLE: HttpStatusExample = {
  label: '500 服务端错误',
  kind: 'code',
  code: '500',
  keyword: '',
}

/** 关键字示例。 */
export const KEYWORD_EXAMPLE: HttpStatusExample = {
  label: '按关键字查重定向',
  kind: 'keyword',
  code: '',
  keyword: '重定向',
}

/** 全部示例。 */
export const EXAMPLES: readonly HttpStatusExample[] = [
  DEFAULT_EXAMPLE,
  SERVER_ERROR_EXAMPLE,
  KEYWORD_EXAMPLE,
]
