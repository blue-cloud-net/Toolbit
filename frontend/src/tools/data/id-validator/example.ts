/**
 * id-validator 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface IDValidateExample {
  /** 示例名称。 */
  label: string
  /** type。 */
  type: string
  /** value。 */
  value: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: IDValidateExample = {
  label: 'id-validator 示例',
  type: 'idCard',
  value: '110101199003072516',
}

/** 全部示例。 */
export const EXAMPLES: readonly IDValidateExample[] = [DEFAULT_EXAMPLE]
