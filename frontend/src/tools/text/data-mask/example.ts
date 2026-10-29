/**
 * data-mask 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface DataMaskExample {
  /** 示例名称。 */
  label: string
  /** text。 */
  text: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: DataMaskExample = {
  label: 'data-mask 示例',
  text: '手机号 13800138000，邮箱 a@example.com',
}

/** 全部示例。 */
export const EXAMPLES: readonly DataMaskExample[] = [DEFAULT_EXAMPLE]
