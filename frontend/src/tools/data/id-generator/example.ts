/**
 * id-generator 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface IDGeneratorExample {
  /** 示例名称。 */
  label: string
  /** format。 */
  format: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: IDGeneratorExample = {
  label: 'id-generator 示例',
  format: 'uuid',
}

/** 全部示例。 */
export const EXAMPLES: readonly IDGeneratorExample[] = [DEFAULT_EXAMPLE]
