/**
 * sql-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface SqlFormatExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 操作：format / minify / validate。 */
  action: 'format' | 'minify' | 'validate'
  /** 缩进宽度。 */
  indent: number
  /** SQL 方言。 */
  dialect: 'standard' | 'mysql' | 'pg' | 'sqlite'
  /** 关键字大小写策略。 */
  keywordCase: 'keep' | 'upper' | 'lower'
}

/** 默认示例：美化一条查询。 */
export const DEFAULT_EXAMPLE: SqlFormatExample = {
  label: '美化 SQL',
  input: 'select id, name from users where id = 1 order by name',
  action: 'format',
  indent: 2,
  dialect: 'standard',
  keywordCase: 'keep',
}

/** 关键字转大写示例。 */
export const UPPER_EXAMPLE: SqlFormatExample = {
  label: '关键字转大写',
  input: 'select id, name from users where id = 1',
  action: 'format',
  indent: 4,
  dialect: 'standard',
  keywordCase: 'upper',
}

/** 全部示例。 */
export const EXAMPLES: readonly SqlFormatExample[] = [
  DEFAULT_EXAMPLE,
  UPPER_EXAMPLE,
]
