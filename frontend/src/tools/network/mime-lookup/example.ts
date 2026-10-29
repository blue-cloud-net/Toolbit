/**
 * mime-lookup 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 查询方式。 */
export type MimeQueryKind = 'extension' | 'mimeType'

/** 一个示例。 */
export interface MimeLookupExample {
  /** 示例名称。 */
  label: string
  /** 查询方式。 */
  kind: MimeQueryKind
  /** 扩展名查询词；仅 `kind` 为 `extension` 时有意义。 */
  extension: string
  /** 媒体类型查询词；仅 `kind` 为 `mimeType` 时有意义。 */
  mimeType: string
}

/** 默认示例：按扩展名查 PNG。 */
export const DEFAULT_EXAMPLE: MimeLookupExample = {
  label: 'PNG 扩展名',
  kind: 'extension',
  extension: 'png',
  mimeType: '',
}

/** 文件名形式：按文件名里的扩展名查。 */
export const FILENAME_EXAMPLE: MimeLookupExample = {
  label: '按文件名查 PDF',
  kind: 'extension',
  extension: 'report.pdf',
  mimeType: '',
}

/** 按媒体类型查。 */
export const TYPE_EXAMPLE: MimeLookupExample = {
  label: '按媒体类型查 JSON',
  kind: 'mimeType',
  extension: '',
  mimeType: 'application/json',
}

/** 全部示例。 */
export const EXAMPLES: readonly MimeLookupExample[] = [
  DEFAULT_EXAMPLE,
  FILENAME_EXAMPLE,
  TYPE_EXAMPLE,
]
