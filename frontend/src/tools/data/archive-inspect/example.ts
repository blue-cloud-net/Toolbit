/**
 * archive-inspect 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface ArchiveInspectExample {
  /** 示例名称。 */
  label: string
  /** data。 */
  data: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: ArchiveInspectExample = {
  label: 'archive-inspect 示例',
  data: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly ArchiveInspectExample[] = [DEFAULT_EXAMPLE]
