/**
 * dns-ptr 工具的示例数据。
 */

/** 一个示例。 */
export interface DNSPTRExample {
  /** 示例名称。 */
  label: string
  /** 域名；可为空。 */
  domain: string
  /** IP 地址；可为空。 */
  ip: string
}

/** 默认示例：反查 8.8.8.8。 */
export const DEFAULT_EXAMPLE: DNSPTRExample = {
  label: '反查 8.8.8.8',
  domain: '',
  ip: '8.8.8.8',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSPTRExample[] = [DEFAULT_EXAMPLE]
