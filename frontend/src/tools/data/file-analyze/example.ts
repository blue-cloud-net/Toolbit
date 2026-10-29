/**
 * file-analyze 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface FileAnalyzeExample {
  /** 示例名称。 */
  label: string
  /** data。 */
  data: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: FileAnalyzeExample = {
  label: 'file-analyze 示例',
  data: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly FileAnalyzeExample[] = [DEFAULT_EXAMPLE]
