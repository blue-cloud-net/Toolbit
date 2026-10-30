/**
 * mime-lookup 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface MimeLookupExample {
  /** 示例名称。 */
  label: string
  /** 查询词：扩展名、文件名或媒体类型。 */
  query: string
}

/** 默认示例：按扩展名查 PNG。 */
export const DEFAULT_EXAMPLE: MimeLookupExample = {
  label: 'PNG 扩展名',
  query: 'png',
}

/** 文件名形式：按文件名里的扩展名查。 */
export const FILENAME_EXAMPLE: MimeLookupExample = {
  label: '按文件名查 PDF',
  query: 'report.pdf',
}

/** 按媒体类型查。 */
export const TYPE_EXAMPLE: MimeLookupExample = {
  label: '按媒体类型查 JSON',
  query: 'application/json',
}

/** 全部示例。 */
export const EXAMPLES: readonly MimeLookupExample[] = [
  DEFAULT_EXAMPLE,
  FILENAME_EXAMPLE,
  TYPE_EXAMPLE,
]
