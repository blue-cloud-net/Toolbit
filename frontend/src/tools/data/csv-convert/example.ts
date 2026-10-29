/**
 * csv-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface CSVConvertExample {
  /** 示例名称。 */
  label: string
  /** input。 */
  input: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: CSVConvertExample = {
  label: 'csv-convert 示例',
  input: 'name,age\nAlice,30\nBob,25',
}

/** 全部示例。 */
export const EXAMPLES: readonly CSVConvertExample[] = [DEFAULT_EXAMPLE]
