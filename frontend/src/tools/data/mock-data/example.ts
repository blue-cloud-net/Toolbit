/**
 * mock-data 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface MockDataExample {
  /** 示例名称。 */
  label: string
  /** types。 */
  types: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: MockDataExample = {
  label: 'mock-data 示例',
  types: 'email',
}

/** 全部示例。 */
export const EXAMPLES: readonly MockDataExample[] = [DEFAULT_EXAMPLE]
