/**
 * ct-search 工具的示例数据。
 */
export interface CTSearchExample {
  /** 示例名称。 */
  label: string
  /** 域名。 */
  domain: string
  /** 匹配方式。 */
  matchType: 'exact' | 'contains' | 'wildcard'
  /** 返回条数。 */
  limit: number
}

/** 默认示例：查询 example.com。 */
export const DEFAULT_EXAMPLE: CTSearchExample = {
  label: '查询 example.com',
  domain: 'example.com',
  matchType: 'exact',
  limit: 20,
}

/** 全部示例。 */
export const EXAMPLES: readonly CTSearchExample[] = [DEFAULT_EXAMPLE]
