/**
 * password-policy 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface PasswordPolicyExample {
  /** 示例名称。 */
  label: string
  /** password。 */
  password: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: PasswordPolicyExample = {
  label: 'password-policy 示例',
  password: 'Example123',
}

/** 全部示例。 */
export const EXAMPLES: readonly PasswordPolicyExample[] = [DEFAULT_EXAMPLE]
