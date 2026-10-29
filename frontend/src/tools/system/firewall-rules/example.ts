/**
 * firewall-rules 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface FirewallRulesExample {
  /** 示例名称。 */
  label: string
  /** scenario。 */
  scenario: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: FirewallRulesExample = {
  label: 'firewall-rules 示例',
  scenario: 'https',
}

/** 全部示例。 */
export const EXAMPLES: readonly FirewallRulesExample[] = [DEFAULT_EXAMPLE]
