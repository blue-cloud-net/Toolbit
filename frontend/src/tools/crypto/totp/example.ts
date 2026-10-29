/**
 * totp 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface TOTPExample {
  /** 示例名称。 */
  label: string
  /** secret。 */
  secret: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: TOTPExample = {
  label: 'totp 示例',
  secret: 'JBSWY3DPEHPK3PXP',
}

/** 全部示例。 */
export const EXAMPLES: readonly TOTPExample[] = [DEFAULT_EXAMPLE]
