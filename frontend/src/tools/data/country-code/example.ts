/**
 * country-code 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** 一个国家代码查询示例。 */
export interface CountryCodeExample {
  /** 示例名称。 */
  label: string
  /** 查询关键词。 */
  query: string
}

/** 默认示例：按中文名查询。 */
export const DEFAULT_EXAMPLE: CountryCodeExample = {
  label: '按中文名查询',
  query: '中国',
}

/** 按两位码查询的示例。 */
export const ALPHA2_EXAMPLE: CountryCodeExample = {
  label: '按两位码查询',
  query: 'JP',
}

/** 按英文名查询的示例。 */
export const ENGLISH_EXAMPLE: CountryCodeExample = {
  label: '按英文名查询',
  query: 'Germany',
}

/** 全部示例。 */
export const EXAMPLES: readonly CountryCodeExample[] = [
  DEFAULT_EXAMPLE,
  ALPHA2_EXAMPLE,
  ENGLISH_EXAMPLE,
]
