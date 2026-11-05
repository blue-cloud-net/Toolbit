/**
 * dnssec-check 工具的示例数据。
 */

/** 一个示例。 */
export interface DNSSECCheckExample {
  /** 示例名称。 */
  label: string
  /** 域名。 */
  domain: string
}

/** 默认示例：检查 example.com 的 DNSSEC 状态。 */
export const DEFAULT_EXAMPLE: DNSSECCheckExample = {
  label: '检查 example.com',
  domain: 'example.com',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSSECCheckExample[] = [DEFAULT_EXAMPLE]
