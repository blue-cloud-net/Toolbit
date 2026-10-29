/**
 * ip-range-expand 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface IPRangeExample {
  /** 示例名称。 */
  label: string
  /** cidr。 */
  cidr: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: IPRangeExample = {
  label: 'ip-range-expand 示例',
  cidr: '192.168.1.0/24',
}

/** 全部示例。 */
export const EXAMPLES: readonly IPRangeExample[] = [DEFAULT_EXAMPLE]
